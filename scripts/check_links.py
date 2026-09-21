#!/usr/bin/env python3
"""Check internal Markdown links and heading anchors across the repository.

Offline, no network, no external dependencies. Checks two things for every
`[text](target)` link in every tracked `*.md` file (excluding `review-output/`,
which holds review artifacts, not repository documentation):

1. The linked file (if the target is not a bare fragment) exists on disk,
   relative to the linking file's directory.
2. If the link has a `#fragment`, that fragment matches an actual heading in
   the target file (or the same file, for same-page fragments), using
   GitHub's heading-to-anchor slug algorithm -- including its de-duplication
   suffix (`-1`, `-2`, ...) for repeated headings.

External links (http/https/mailto) and pure-fragment links to headings that
can't be resolved automatically (rare) are not checked.

Run from the repository root: python3 scripts/check_links.py
Exit code 0 if every checked link resolves, 1 otherwise.
"""
import os
import re
import subprocess
import sys

LINK_RE = re.compile(r"\[([^\]]+)\]\(([^)\s]+)\)")
HEADING_RE = re.compile(r"^(#{1,6})\s+(.+)$", re.M)
EXCLUDED_PREFIXES = ("review-output/",)


def gh_anchor(heading: str) -> str:
    """Approximate GitHub's heading-to-anchor slug algorithm."""
    h = heading.lower()
    h = re.sub(r"[^\w\s-]", "", h)
    h = re.sub(r"\s+", "-", h.strip())
    return h


def anchors_for(path: str, cache: dict) -> set:
    if path in cache:
        return cache[path]
    try:
        content = open(path, encoding="utf-8").read()
    except OSError:
        cache[path] = set()
        return cache[path]
    seen: dict = {}
    anchors = set()
    for _, text in HEADING_RE.findall(content):
        base = gh_anchor(text)
        n = seen.get(base, 0)
        seen[base] = n + 1
        anchors.add(base if n == 0 else f"{base}-{n}")
    cache[path] = anchors
    return anchors


def tracked_markdown_files() -> list:
    out = subprocess.run(
        ["git", "ls-files", "*.md"], capture_output=True, text=True, check=True
    ).stdout
    return [
        f
        for f in out.splitlines()
        if f and not any(f.startswith(p) for p in EXCLUDED_PREFIXES)
    ]


def main() -> int:
    files = tracked_markdown_files()
    anchor_cache: dict = {}
    problems = []
    checked = 0

    for src in files:
        content = open(src, encoding="utf-8").read()
        # Strip fenced code blocks and inline code spans first so an example
        # like `[text](url)` illustrating link syntax isn't treated as a real
        # link.
        content = re.sub(r"```.*?```", "", content, flags=re.S)
        content = re.sub(r"`[^`\n]*`", "", content)
        base_dir = os.path.dirname(src)
        for _text, target in LINK_RE.findall(content):
            if target.startswith(("http://", "https://", "mailto:")):
                continue
            path_part, _, frag = target.partition("#")

            if path_part:
                resolved = os.path.normpath(os.path.join(base_dir, path_part))
                checked += 1
                if not os.path.exists(resolved):
                    problems.append(f"{src}: link to {target!r} -- file not found ({resolved})")
                    continue
                target_file = resolved
            else:
                target_file = src

            if frag:
                checked += 1
                if frag not in anchors_for(target_file, anchor_cache):
                    problems.append(
                        f"{src}: link to {target!r} -- no heading in {target_file} slugs to #{frag}"
                    )

    print(f"Checked {checked} internal link/anchor references across {len(files)} tracked Markdown files.")
    if problems:
        for p in problems:
            print(f"::error::{p}")
        print(f"\n{len(problems)} broken internal link(s)/anchor(s) found.")
        return 1

    print("No broken internal links or anchors found.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
