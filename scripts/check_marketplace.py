#!/usr/bin/env python3
"""Check that the marketplaces only ship pinned, reviewable plugin content.

Checks every plugin in .claude-plugin/marketplace.json and .agents/plugins/marketplace.json:

1. A plugin from another repository pins a full 40-character commit `sha`, and
   that commit is on the default branch of the named repository. GitHub serves
   commits that only exist in a fork through the original repository's URL, so
   a pin must be checked against the branch, not just fetched. The Codex
   marketplace may only list local plugins.
2. No plugin runs code by itself. Fails on hooks, MCP, LSP and monitor servers
   (in the marketplace entry, in plugin.json, or as hooks.json, .mcp.json,
   .lsp.json, monitors.json), on `bin/` and `workflows/` at the plugin root, and
   on skills or commands that declare `hooks` or `allowed-tools` or inject shell
   output with !`command`.
3. Warns when a skill reads another repository's branch at runtime
   (raw.githubusercontent.com or blob/tree links to a branch, llms-full.txt).
   That skips the pin. Links pinned to a commit, and links to this repository,
   are fine.
4. With --base, writes what changed in each linked plugin since the base pin to
   the job summary: a compare link, the changed files and any scripts. The PR
   diff itself only shows a new sha.

Linked plugins are fetched at the pinned commit and checked on disk, like local
plugins. Needs git and the `gh` CLI with a token (GH_TOKEN).

Run from the repository root:
    python3 scripts/check_marketplace.py [--base origin/main]
Exit code 0 if every check passes, 1 otherwise. Warnings don't fail the run.
"""
import argparse
import json
import os
import re
import subprocess
import sys
import tempfile

MARKETPLACE = ".claude-plugin/marketplace.json"
CODEX_MARKETPLACE = ".agents/plugins/marketplace.json"
THIS_REPO = "entur/ai"
REMOTE_SOURCES = ("github", "url", "git-subdir")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
GITHUB_URL_RE = re.compile(r"^https://github\.com/([\w.-]+/[\w.-]+?)(?:\.git)?/?$")
CODE_KEYS = ("hooks", "mcpServers", "lspServers", "monitors", "experimental")
CODE_FILES = ("hooks.json", ".mcp.json", ".lsp.json", "monitors.json")
CODE_DIRS = ("bin", "workflows")
MANIFESTS = (".claude-plugin/plugin.json", ".codex-plugin/plugin.json")
FRONTMATTER_KEYS = ("hooks", "allowed-tools")
SHELL_INJECTION_RE = re.compile(r"!`[^`]+`|^\s*```!")
SCRIPT_EXTENSIONS = (".sh", ".bash", ".py", ".js", ".mjs", ".cjs", ".ts", ".rb", ".pl", ".ps1")
# A link to another repository's branch, captured as (owner/repo, ref).
BRANCH_LINK_RES = (
    re.compile(r"raw\.githubusercontent\.com/([\w.-]+/[\w.-]+)/(?:refs/heads/)?([\w.-]+)/"),
    re.compile(r"github\.com/([\w.-]+/[\w.-]+)/(?:blob|tree)/([\w.-]+)"),
)
MAX_COMPARE_FILES = 300

errors = []
warnings = []


def gh(*args):
    result = subprocess.run(["gh", "api", *args], capture_output=True, text=True)
    if result.returncode != 0:
        return None
    return json.loads(result.stdout)


def git(*args, cwd=None):
    return subprocess.run(["git", *args], cwd=cwd, capture_output=True, text=True)


def repo_of(source):
    if source["source"] == "github":
        return source.get("repo")
    match = GITHUB_URL_RE.match(source.get("url", ""))
    return match.group(1) if match else None


def subdir_of(source):
    return source.get("path", "").strip("/")


def is_script(path, mode):
    return mode == "100755" or path.endswith(SCRIPT_EXTENSIONS)


def check_code_keys(name, where, data):
    for key in CODE_KEYS:
        if key in data:
            errors.append(f"{name}: {where} declares `{key}`")


def frontmatter(text):
    if not text.startswith("---"):
        return ""
    end = text.find("\n---", 3)
    return text[3:end] if end != -1 else ""


def check_markdown(name, rel, text):
    is_instructions = os.path.basename(rel) == "SKILL.md" or "/commands/" in f"/{rel}"
    if is_instructions:
        for key in FRONTMATTER_KEYS:
            if re.search(rf"^{key}\s*:", frontmatter(text), re.MULTILINE):
                errors.append(f"{name}: {rel} declares `{key}` in its frontmatter")
        for number, line in enumerate(text.splitlines(), 1):
            if SHELL_INJECTION_RE.search(line):
                errors.append(f"{name}: {rel}:{number} runs a shell command when the skill loads")

    if rel == "README.md":
        return
    for number, line in enumerate(text.splitlines(), 1):
        if "llms-full" in line:
            warnings.append(f"{name}: {rel}:{number} points at llms-full.txt, which is built from a branch")
        for pattern in BRANCH_LINK_RES:
            for repo, ref in pattern.findall(line):
                if repo.lower() != THIS_REPO and not SHA_RE.match(ref):
                    warnings.append(f"{name}: {rel}:{number} reads {repo} at `{ref}`, not at a commit")


def check_tree(name, root):
    """Check a plugin on disk. `root` is the plugin's own folder."""
    if not os.path.isdir(root):
        errors.append(f"{name}: plugin folder {root} does not exist")
        return
    for entry in CODE_DIRS:
        if os.path.isdir(os.path.join(root, entry)):
            errors.append(f"{name}: {entry}/ at the plugin root runs code")
    for dirpath, dirnames, files in os.walk(root, followlinks=True):
        dirnames[:] = [d for d in dirnames if d != ".git"]
        for file in files:
            path = os.path.join(dirpath, file)
            rel = os.path.relpath(path, root)
            if file in CODE_FILES:
                errors.append(f"{name}: {rel} adds hooks or servers that run code")
            if rel.endswith(MANIFESTS):
                try:
                    with open(path, encoding="utf-8") as f:
                        check_code_keys(name, rel, json.load(f))
                except (OSError, ValueError) as error:
                    errors.append(f"{name}: could not read {rel}: {error}")
            if file.endswith(".md"):
                with open(path, encoding="utf-8", errors="replace") as f:
                    check_markdown(name, rel, f.read())


def fetch_commit(repo, sha, subdir, workdir):
    """Fetch one commit into workdir, with file contents only for subdir. True on success."""
    git("init", "-q", workdir)
    url = f"https://github.com/{repo}.git"
    fetched = git("fetch", "-q", "--depth", "1", "--filter=blob:none", url, sha, cwd=workdir)
    if fetched.returncode != 0:
        return False
    if subdir:
        git("sparse-checkout", "set", "--no-cone", f"/{subdir}/", cwd=workdir)
    return git("checkout", "-q", sha, cwd=workdir).returncode == 0


def check_remote(name, source, workdir):
    repo, sha = repo_of(source), source.get("sha", "")
    if not repo:
        errors.append(f"{name}: source must be a github.com repository")
        return None
    if not SHA_RE.match(sha):
        errors.append(f"{name}: pin a full 40-character commit `sha`, not a branch or tag")
        return None
    if "ref" in source:
        errors.append(f"{name}: remove `ref`; the `sha` is what gets installed")

    branch = (gh(f"repos/{repo}") or {}).get("default_branch")
    status = (gh(f"repos/{repo}/compare/{branch}...{sha}") or {}).get("status") if branch else None
    if status not in ("behind", "identical"):
        errors.append(f"{name}: commit {sha} is not on the default branch of {repo}")
        return None

    subdir = subdir_of(source)
    if not fetch_commit(repo, sha, subdir, workdir):
        errors.append(f"{name}: could not fetch {repo} at {sha}")
        return None
    check_tree(name, os.path.join(workdir, subdir))
    return repo, sha, subdir, workdir


def check_entry(plugin, tmp):
    name = plugin["name"]
    check_code_keys(name, "marketplace entry", plugin)
    source = plugin["source"]

    if isinstance(source, str):
        check_tree(name, source)
        return None
    if source.get("source") not in REMOTE_SOURCES:
        errors.append(f"{name}: source type `{source.get('source')}` is not allowed")
        return None
    return check_remote(name, source, os.path.join(tmp, name))


def check_codex_marketplace():
    if not os.path.exists(CODEX_MARKETPLACE):
        return
    with open(CODEX_MARKETPLACE, encoding="utf-8") as f:
        plugins = json.load(f).get("plugins", [])
    for plugin in plugins:
        name, source = plugin.get("name"), plugin.get("source", {})
        if not isinstance(source, dict) or source.get("source") != "local":
            errors.append(f"{name}: the Codex marketplace may only list local plugins")
            continue
        check_code_keys(name, "Codex marketplace entry", plugin)
        check_tree(name, source.get("path", ""))


def base_pins(base):
    shown = git("show", f"{base}:{MARKETPLACE}")
    if shown.returncode != 0:
        return {}
    pins = {}
    for plugin in json.loads(shown.stdout)["plugins"]:
        source = plugin["source"]
        if isinstance(source, dict) and source.get("source") in REMOTE_SOURCES:
            pins[plugin["name"]] = source.get("sha")
    return pins


def file_modes(workdir):
    listed = git("ls-tree", "-r", "HEAD", cwd=workdir).stdout
    modes = {}
    for line in listed.splitlines():
        meta, path = line.split("\t", 1)
        modes[path] = meta.split()[0]
    return modes


def summary(linked, base):
    old_pins = base_pins(base)
    lines = []
    for name, (repo, sha, subdir, workdir) in linked.items():
        old = old_pins.get(name)
        if old == sha:
            continue
        prefix = f"{subdir}/" if subdir else ""
        modes = file_modes(workdir)
        if old and SHA_RE.match(old):
            compare = gh(f"repos/{repo}/compare/{old}...{sha}")
            lines.append(f"### {name}: https://github.com/{repo}/compare/{old}...{sha}")
            if compare is None:
                warnings.append(f"{name}: could not compare {old}...{sha}; read the compare link")
                lines += ["", "**Could not list the changed files. Read the compare link.**", ""]
                continue
            files = [(f["filename"], f["status"]) for f in compare.get("files", [])]
            if len(files) >= MAX_COMPARE_FILES:
                warnings.append(f"{name}: the file list stops at {MAX_COMPARE_FILES}; read the compare link")
        else:
            files = [(path, "new") for path in modes]
            lines.append(f"### {name}: new pin, https://github.com/{repo}/tree/{sha}/{subdir}")
        files = [f for f in files if f[0].startswith(prefix)]
        lines.append("")
        lines.append("Read the diff before approving. Skills are instructions the agent follows with the user's access.")
        lines.append("")
        for path, status in files:
            flag = " **script**" if is_script(path, modes.get(path, "")) else ""
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

    with tempfile.TemporaryDirectory() as tmp:
        linked = {}
        for plugin in plugins:
            pinned = check_entry(plugin, tmp)
            if pinned:
                linked[plugin["name"]] = pinned
        check_codex_marketplace()

        if args.base:
            summary(linked, args.base)

    for warning in warnings:
        print(f"::warning file={MARKETPLACE}::{warning}")
    for error in errors:
        print(f"::error file={MARKETPLACE}::{error}")
    if errors:
        return 1
    print(f"OK: {len(plugins)} plugins checked, {len(linked)} pinned from other repositories.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
