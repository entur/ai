#!/bin/sh
# PostToolUse Write|Edit hook — warns when a build file has been edited after
# the stack file was last generated. Reads tool input from stdin (Claude Code
# hook protocol). Silent on any error so it never blocks the user.

set -eu

STACK_FILE=".entur/ai/kotlin-spring-boot.json"
[ -f "$STACK_FILE" ] || exit 0

FILE=$(jq -r '.tool_input.file_path // empty' 2>/dev/null) || exit 0
[ -n "$FILE" ] || exit 0

case "$FILE" in
  *build.gradle.kts|*build.gradle|*pom.xml|*gradle/libs.versions.toml) ;;
  *) exit 0 ;;
esac

if [ "$FILE" -nt "$STACK_FILE" ]; then
  cat <<EOF >&2
[kotlin-spring-boot] Build file changed after $STACK_FILE.
Re-run /kotlin-spring-boot:init if any stack axis (build tool, spring stack,
database, test libs, formatter) may have shifted.
EOF
fi
