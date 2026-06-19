# Scenario: Kotlin Version Follows CodeQL Support

## Description

Verifies Kotlin version guidance prefers the newest CodeQL-supported Kotlin release line instead of blindly using the newest stable Kotlin release.

## Prompt

You are choosing a Kotlin version for a new Kotlin Spring Boot service at Entur. A teammate suggests using the newest stable Kotlin release immediately.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the Kotlin and security reference guides) to answer.

Output exactly these keys:

- kotlin_version_policy: <one sentence>
- current_baseline_example: <version line or example from the docs>
- exception_owner: <who must approve an exception>

## Assertions

```json
{
  "must_contain": [
    "CodeQL",
    "2.3",
    "Team Sikkerhet"
  ],
  "must_match": [
    "newest Kotlin release line supported by Entur.*CodeQL|CodeQL.*newest Kotlin release line supported by Entur"
  ],
  "must_not_contain": [
    "latest stable (currently 2.x)"
  ]
}
```

## Budget

0.08
