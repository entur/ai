#!/usr/bin/env python3
"""Validate that every tests/framing/scenarios/*.md file has a well-formed
JSON assertions block. Offline, no network, no API calls -- safe for CI.
Run from the tests/ directory: python3 framing/check_scenarios.py
"""
import glob
import json
import re
import sys

FENCE = re.compile(r"```json\n(.*?)\n```", re.S)


def main() -> int:
    failures = []
    paths = sorted(glob.glob("framing/scenarios/*.md"))
    if not paths:
        print("::error::no files matched framing/scenarios/*.md -- run this from the tests/ directory")
        return 1

    for path in paths:
        content = open(path, encoding="utf-8").read()
        match = FENCE.search(content)
        if not match:
            failures.append(f"{path}: no JSON assertions block")
            continue
        try:
            json.loads(match.group(1))
        except Exception as exc:  # noqa: BLE001 - report any parse failure
            failures.append(f"{path}: invalid JSON: {exc}")

    if failures:
        for failure in failures:
            print(f"::error::{failure}")
        return 1

    print(f"OK: {len(paths)} framing scenario assertion blocks parse as valid JSON")
    return 0


if __name__ == "__main__":
    sys.exit(main())
