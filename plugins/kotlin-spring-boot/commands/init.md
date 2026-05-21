---
description: Scan the repo and write .claude/entur/kotlin-spring-boot.json. Run once per repo, then commit it.
argument-hint: "[--force]"
---

# Initialize Kotlin Spring Boot stack file

Write `.claude/entur/kotlin-spring-boot.json` so the conventions skill matches this repo's stack. Commit the file.

## Steps

### Check Spring Boot signal

Confirm one of:

- `build.gradle.kts` / `build.gradle`: `org.springframework.boot` plugin or `spring-boot-starter-*`
- `pom.xml`: `spring-boot-starter-parent` or `spring-boot-starter-*`
- `gradle/libs.versions.toml`: a `spring-boot` version

No match → not a Kotlin Spring Boot project. Write nothing. Stop.

### Check for existing file

If `.claude/entur/kotlin-spring-boot.json` exists:

- **Without `--force`**: show current values, then ask: overwrite all, edit specific axes, or cancel.
  - Cancel → stop.
  - Edit → re-prompt only the chosen axes; keep all other values unchanged.
  - Overwrite → proceed to "Detect axes" as if no file existed.
- **With `--force`**: skip this step entirely and proceed directly to "Detect axes".

### Detect axes

Read `build.gradle.kts`, `build.gradle`, `pom.xml`, `gradle/libs.versions.toml`. Auto-fill confident detections; ask only when ambiguous or undetected.

| Axis | Signal | Default |
|---|---|---|
| `build_tool` | `gradlew` or `build.gradle.kts` → `gradle`. `mvnw` or `pom.xml` → `maven` | required |
| `spring_stack` | `spring-boot-starter-webflux` → `webflux`. `spring-boot-starter-web` → `mvc` | `mvc` |
| `api_approach` | `org.openapi.generator` plugin + `specs/` dir → `contract-first` | `traditional` |
| `database` | `exposed-spring-boot-starter` → `exposed`. `spring-boot-starter-data-jdbc` → `spring-data-jdbc`. `spring-boot-starter-data-jpa` → `jpa` | `none` |
| `test_mocking` | `com.ninja-squad:springmockk` → `mockk`. `org.mockito.kotlin:mockito-kotlin` → `mockito-kotlin` | `mockk` |
| `test_assertions` | `io.kotest:kotest-assertions-core` → `kotest` | `assertj` |
| `formatter` | `com.diffplug.spotless` Gradle plugin → `spotless-gradle`. `spotless-maven-plugin` → `spotless-maven`. `org.jlleitschuh.gradle.ktlint` or `.editorconfig` ktlint rules → `ktlint` | `ktlint` |

Ambiguity:

- Both `webflux` and `web` present → ask which is primary
- Multiple database starters → ask which is canonical
- Build files unreadable → stop, tell the user the repo state isn't parseable

### Ask about legacy

> Is this a legacy codebase? `legacy_mode: true` suppresses modernization advice — no toolchain upgrades, no version catalog migrations, no framework bumps. Match existing patterns; fix bugs, add tests, make small additions. Default `false`.

### Ask for notes

> Any non-default choice that needs context? Example: "Webflux for the legacy reactive pipeline; do not migrate to MVC." Leave blank to skip.

### Write the file

Path: `.claude/entur/kotlin-spring-boot.json`. Two-space indent. Create parent dirs.

```json
{
  "version": 1,
  "generated_at": "<ISO 8601 UTC>",
  "build_tool": "<value>",
  "spring_stack": "<value>",
  "api_approach": "<value>",
  "database": "<value>",
  "test_mocking": "<value>",
  "test_assertions": "<value>",
  "formatter": "<value>",
  "legacy_mode": <bool>,
  "notes": "<string>"
}
```

### Confirm

Print one axis per line. Then:

> Stack file written. Commit so teammates pick it up:
>
>     git add .claude/entur/kotlin-spring-boot.json
>     git commit -m 'chore: initialize kotlin-spring-boot stack file'

## Constraints

- Don't modify build files, source code, or anything outside `.claude/entur/`
- Read values from the actual repo files; don't infer from training data
- If a build file is unreadable or unparseable, ask the user; don't guess
- Don't write secrets, internal URLs, or environment-specific values
