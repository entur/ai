#!/usr/bin/env python3
"""Check that the Claude Code marketplace only ships pinned, reviewable plugin content.

Checks every plugin in .claude-plugin/marketplace.json:

1. A plugin from another repository pins a full 40-character commit `sha`, and
   that commit is on the default branch of the named repository. GitHub serves
   commits that only exist in a fork through the original repository's URL, so
   a pin must be checked against the branch, not just fetched.
2. No plugin declares hooks, MCP servers or LSP servers: not in its marketplace
   entry, not in its plugin.json, and not as hooks/hooks.json, .mcp.json or
   .lsp.json. These run code on the user's machine without asking.
3. With --base, writes what changed in each linked plugin since the base pin to
   the job summary: a compare link, the changed files and any scripts. The PR
   diff itself only shows a new sha.

Needs the `gh` CLI with a token (GH_TOKEN). No other dependencies.

Run from the repository root:
    python3 scripts/check_marketplace.py [--base origin/main]
Exit code 0 if every check passes, 1 otherwise.
"""
import argparse
import json
import os
import re
import subprocess
import sys

MARKETPLACE = ".claude-plugin/marketplace.json"
REMOTE_SOURCES = ("github", "url", "git-subdir")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
GITHUB_URL_RE = re.compile(r"^https://github\.com/([\w.-]+/[\w.-]+?)(?:\.git)?/?$")
CODE_KEYS = ("hooks", "mcpServers", "lspServers")
CODE_FILES = ("hooks.json", ".mcp.json", ".lsp.json")
SCRIPT_EXTENSIONS = (".sh", ".bash", ".py", ".js", ".mjs", ".cjs", ".ts", ".rb", ".pl", ".ps1")

errors = []


def gh(*args):
    result = subprocess.run(["gh", "api", *args], capture_output=True, text=True)
    if result.returncode != 0:
        return None
    return json.loads(result.stdout)


def repo_of(source):
    if source["source"] == "github":
        return source.get("repo")
    match = GITHUB_URL_RE.match(source.get("url", ""))
    return match.group(1) if match else None


def subdir_of(source):
    return source.get("path", "").strip("/") if source["source"] == "git-subdir" else ""


def is_script(path, mode):
    return mode == "100755" or path.endswith(SCRIPT_EXTENSIONS)


def check_code_keys(name, where, data):
    for key in CODE_KEYS:
        if key in data:
            errors.append(f"{name}: {where} declares `{key}`")


def check_entry(plugin):
    name = plugin["name"]
    check_code_keys(name, "marketplace entry", plugin)
    source = plugin["source"]

    if isinstance(source, str):
        check_local_plugin(name, source)
        return None
    if source.get("source") not in REMOTE_SOURCES:
        errors.append(f"{name}: source type `{source.get('source')}` is not allowed")
        return None

    repo, sha = repo_of(source), source.get("sha", "")
    if not repo:
        errors.append(f"{name}: source must be a github.com repository")
        return None
    if not SHA_RE.match(sha):
        errors.append(f"{name}: pin a full 40-character commit `sha`, not a branch or tag")
        return None

    branch = (gh(f"repos/{repo}") or {}).get("default_branch")
    status = (gh(f"repos/{repo}/compare/{branch}...{sha}") or {}).get("status") if branch else None
    if status not in ("behind", "identical"):
        errors.append(f"{name}: commit {sha} is not on the default branch of {repo}")
        return None

    check_remote_tree(name, repo, sha, subdir_of(source))
    return repo, sha, subdir_of(source)


def check_remote_tree(name, repo, sha, subdir):
    tree = gh(f"repos/{repo}/git/trees/{sha}?recursive=1")
    if not tree or tree.get("truncated"):
        errors.append(f"{name}: could not list the files of {repo} at {sha}")
        return
    prefix = f"{subdir}/" if subdir else ""
    for item in tree["tree"]:
        path = item["path"]
        if item["type"] != "blob" or not path.startswith(prefix):
            continue
        rel = path[len(prefix):]
        if os.path.basename(rel) in CODE_FILES:
            errors.append(f"{name}: {rel} adds hooks or servers that run code")
        if rel.endswith(".claude-plugin/plugin.json"):
            content = gh(f"repos/{repo}/contents/{path}?ref={sha}", "-H", "Accept: application/vnd.github.raw")
            check_code_keys(name, rel, content or {})


def check_local_plugin(name, root):
    for dirpath, _, files in os.walk(root, followlinks=True):
        for file in files:
            path = os.path.join(dirpath, file)
            if file in CODE_FILES:
                errors.append(f"{name}: {path} adds hooks or servers that run code")
            if path.endswith(".claude-plugin/plugin.json"):
                with open(path, encoding="utf-8") as f:
                    check_code_keys(name, path, json.load(f))


def base_pins(base):
    shown = subprocess.run(["git", "show", f"{base}:{MARKETPLACE}"], capture_output=True, text=True)
    if shown.returncode != 0:
        return {}
    pins = {}
    for plugin in json.loads(shown.stdout)["plugins"]:
        source = plugin["source"]
        if isinstance(source, dict) and source.get("source") in REMOTE_SOURCES:
            pins[plugin["name"]] = source.get("sha")
    return pins


def summary(linked, base):
    old_pins = base_pins(base)
    lines = []
    for name, (repo, sha, subdir) in linked.items():
        old = old_pins.get(name)
        if old == sha:
            continue
        prefix = f"{subdir}/" if subdir else ""
        if old and SHA_RE.match(old):
            compare = gh(f"repos/{repo}/compare/{old}...{sha}") or {}
            files = [(f["filename"], f["status"], "") for f in compare.get("files", [])]
            lines.append(f"### {name}: https://github.com/{repo}/compare/{old}...{sha}")
        else:
            tree = gh(f"repos/{repo}/git/trees/{sha}?recursive=1") or {"tree": []}
            files = [(t["path"], "new", t["mode"]) for t in tree["tree"] if t["type"] == "blob"]
            lines.append(f"### {name}: new pin, https://github.com/{repo}/tree/{sha}/{subdir}")
        files = [f for f in files if f[0].startswith(prefix)]
        lines.append("")
        lines.append("Read the diff before approving. Skills are instructions the agent follows with the user's access.")
        lines.append("")
        for path, status, mode in files:
            flag = " **script**" if is_script(path, mode) else ""
            lines.append(f"- `{path[len(prefix):]}` ({status}){flag}")
        lines.append("")
    if not lines:
        return
    text = "## Linked plugin changes\n\n" + "\n".join(lines) + "\n"
    target = os.environ.get("GITHUB_STEP_SUMMARY")
    if target:
        with open(target, "a", encoding="utf-8") as f:
            f.write(text)
    print(text)


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--base", help="git ref to compare pins against, e.g. origin/main")
    args = parser.parse_args()

    with open(MARKETPLACE, encoding="utf-8") as f:
        plugins = json.load(f)["plugins"]

    linked = {}
    for plugin in plugins:
        pinned = check_entry(plugin)
        if pinned:
            linked[plugin["name"]] = pinned

    if args.base:
        summary(linked, args.base)

    for error in errors:
        print(f"::error file={MARKETPLACE}::{error}")
    if errors:
        return 1
    print(f"OK: {len(plugins)} plugins checked, {len(linked)} pinned from other repositories.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
