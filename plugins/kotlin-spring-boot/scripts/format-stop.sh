#!/bin/sh
# Stop hook — runs spotlessApply at end of each Claude turn when the repo's
# stack file says formatter=spotless-gradle or spotless-maven. Requires both
# the stack file and a Spotless config in the repo's build files. Exits
# silently on any miss.

set -eu

STACK_FILE=".claude/entur/kotlin-spring-boot.json"
[ -f "$STACK_FILE" ] || exit 0

FORMATTER=$(jq -r '.formatter // ""' "$STACK_FILE" 2>/dev/null) || exit 0

case "$FORMATTER" in
  spotless-gradle)
    [ -x "./gradlew" ] || exit 0
    grep -qE '(spotless|com\.diffplug\.spotless)' build.gradle.kts build.gradle 2>/dev/null || exit 0
    ./gradlew spotlessApply --quiet >/dev/null 2>&1 || exit 0
    ;;
  spotless-maven)
    [ -x "./mvnw" ] || exit 0
    grep -q "spotless-maven-plugin" pom.xml 2>/dev/null || exit 0
    ./mvnw -q spotless:apply >/dev/null 2>&1 || exit 0
    ;;
  *) exit 0 ;;
esac
