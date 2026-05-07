#!/bin/sh
# PostToolUse Write|Edit hook — runs ktlint on .kt files when the repo's
# stack file says formatter=ktlint. Requires .claude/entur/kotlin-spring-boot.json
# (run /kotlin-spring-boot:init to generate). Exits silently on any miss so the
# hook never blocks editing.

set -eu

STACK_FILE=".claude/entur/kotlin-spring-boot.json"
[ -f "$STACK_FILE" ] || exit 0

FORMATTER=$(jq -r '.formatter // ""' "$STACK_FILE" 2>/dev/null) || exit 0
[ "$FORMATTER" = "ktlint" ] || exit 0

FILE=$(jq -r '.tool_input.file_path // empty' 2>/dev/null) || exit 0
[ -n "$FILE" ] || exit 0
case "$FILE" in *.kt) ;; *) exit 0 ;; esac

command -v ktlint >/dev/null 2>&1 || exit 0
ktlint --format "$FILE" >/dev/null 2>&1 || exit 0
