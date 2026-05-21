#!/bin/sh
# Stop hook — runs scoped Spotless formatting for files edited in the current
# turn when formatter=spotless-gradle or spotless-maven. Requires stack file
# plus Spotless config in build files. Exits silently on any miss.

set -eu

STACK_FILE=".entur/ai/kotlin-spring-boot.json"
CHANGED_FILES_FILE=".entur/ai/kotlin-spring-boot.changed-files"
[ -f "$STACK_FILE" ] || exit 0

FORMATTER=$(jq -r '.formatter // ""' "$STACK_FILE" 2>/dev/null) || exit 0
case "$FORMATTER" in
  spotless-gradle|spotless-maven) ;;
  *) exit 0 ;;
esac

[ -s "$CHANGED_FILES_FILE" ] || exit 0

TMP_FILES=$(mktemp "${TMPDIR:-/tmp}/kotlin-spring-boot-files.XXXXXX") || exit 0
cleanup() {
  rm -f "$TMP_FILES" "$CHANGED_FILES_FILE"
}
trap cleanup EXIT HUP INT TERM

sort -u "$CHANGED_FILES_FILE" > "$TMP_FILES" || exit 0

FILES=$(
  awk 'NF > 0' "$TMP_FILES" \
    | while IFS= read -r file; do
        case "$file" in
          "$PWD"/*) file=${file#"$PWD"/} ;;
        esac
        [ -f "$file" ] || continue
        printf '%s\n' "$file"
      done \
    | paste -sd, -
)
[ -n "${FILES:-}" ] || exit 0

case "$FORMATTER" in
  spotless-gradle)
    [ -x "./gradlew" ] || exit 0
    grep -qE '(spotless|com\.diffplug\.spotless)' build.gradle.kts build.gradle 2>/dev/null || exit 0
    ./gradlew spotlessApply -PspotlessFiles="$FILES" --quiet >/dev/null 2>&1 || exit 0
    ;;
  spotless-maven)
    [ -x "./mvnw" ] || exit 0
    grep -q "spotless-maven-plugin" pom.xml 2>/dev/null || exit 0
    ./mvnw -q spotless:apply -DspotlessFiles="$FILES" >/dev/null 2>&1 || exit 0
    ;;
esac
