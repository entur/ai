#!/bin/sh
# PostToolUse Write|Edit hook:
# - formatter=ktlint: format edited .kt/.kts files immediately
# - formatter=spotless-*: record edited files for scoped formatting in Stop hook
# Requires .entur/ai/kotlin-spring-boot.json (run /kotlin-spring-boot:init
# to generate). Exits silently on any miss so the hook never blocks editing.

set -eu

STACK_FILE=".entur/ai/kotlin-spring-boot.json"
CHANGED_FILES_FILE=".entur/ai/kotlin-spring-boot.changed-files"
[ -f "$STACK_FILE" ] || exit 0

FILE=$(jq -r '.tool_input.file_path // empty' 2>/dev/null) || exit 0
[ -n "$FILE" ] || exit 0

FORMATTER=$(jq -r '.formatter // ""' "$STACK_FILE" 2>/dev/null) || exit 0

case "$FORMATTER" in
  spotless-gradle|spotless-maven)
    mkdir -p ".entur/ai"
    printf '%s\n' "$FILE" >> "$CHANGED_FILES_FILE"
    exit 0
    ;;
  ktlint) ;;
  *) exit 0 ;;
esac

case "$FILE" in *.kt|*.kts) ;; *) exit 0 ;; esac

command -v ktlint >/dev/null 2>&1 || exit 0
ktlint --format "$FILE" >/dev/null 2>&1 || exit 0
