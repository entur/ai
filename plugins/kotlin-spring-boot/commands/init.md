---
description: Generate .claude/entur/kotlin-spring-boot.json by scanning this repo's build files. Run once per repo, then commit the result.
argument-hint: "[--force]"
---

# Initialize Kotlin Spring Boot stack file

Goal: write `.claude/entur/kotlin-spring-boot.json` so the conventions skill applies guidance matching this repo's actual stack. The file is committed and shared with the team.

## Procedure

### 1. Spring Boot signal check

Confirm at least one of:

- `build.gradle.kts` / `build.gradle` references `org.springframework.boot` plugin or any `spring-boot-starter-*`
- `pom.xml` declares `spring-boot-starter-parent` or any `spring-boot-starter-*`
- `gradle/libs.versions.toml` defines a `spring-boot` version

If none match, tell the user this is not a Kotlin Spring Boot project, write nothing, stop.

### 2. Existing-file check

If `.claude/entur/kotlin-spring-boot.json` already exists and `$ARGUMENTS` does not contain `--force`:

- Read it and show the current values
- Ask: "Stack file exists. Overwrite all axes, edit specific axes, or cancel?"
- On cancel → stop. On edit → re-prompt only for chosen axes; keep the rest.

### 3. Detect each axis

Read `build.gradle.kts`, `build.gradle`, `pom.xml`, `gradle/libs.versions.toml`, and `settings.gradle.kts` as available. Auto-fill confident detections. Only ask the user when ambiguous or undetected.

| Axis | Signal | Default |
|---|---|---|
| `build_tool` | `gradlew` or `build.gradle.kts` → `gradle`. `mvnw` or `pom.xml` → `maven` | required, no default |
| `spring_stack` | `spring-boot-starter-webflux` → `webflux`. `spring-boot-starter-web` → `mvc` | `mvc` |
| `api_approach` | `org.openapi.generator` plugin and `specs/` directory → `contract-first` | `traditional` |
| `database` | `org.jetbrains.exposed:exposed-spring-boot-starter` → `exposed`. `spring-boot-starter-data-jdbc` → `spring-data-jdbc`. `spring-boot-starter-data-jpa` → `jpa`. None of the above → `none` | `none` |
| `test_mocking` | `com.ninja-squad:springmockk` → `mockk`. `org.mockito.kotlin:mockito-kotlin` → `mockito-kotlin` | `mockk` |
| `test_assertions` | `io.kotest:kotest-assertions-core` → `kotest` | `assertj` |
| `formatter` | `com.diffplug.spotless` Gradle plugin → `spotless-gradle`. `spotless-maven-plugin` in `pom.xml` → `spotless-maven`. `org.jlleitschuh.gradle.ktlint` plugin or `.editorconfig` ktlint section → `ktlint` | `ktlint` |

Ambiguity rules:

- Both `webflux` and `web` starters present → ask which is the primary runtime
- Multiple database starters present → ask which is canonical for this service
- No build files readable → stop and tell the user the repo state is not parseable

### 4. Legacy flag

Ask the user explicitly:

> "Is this a legacy codebase? `legacy_mode: true` tells the assistant to suppress modernization advice (no toolchain upgrades, no dependency catalog migration, no framework bumps). Match existing patterns; fix bugs, add tests, make small additions. Default `false`."

### 5. Notes

Ask one short question:

> "Any non-default choice that needs context? Example: 'Webflux because of legacy reactive pipeline; do not migrate to MVC.' Leave blank to skip."

Store in `notes`. Empty string is fine.

### 6. Write the file

Path: `.claude/entur/kotlin-spring-boot.json`. Create parent directories as needed. Format with two-space indent.

```json
{
  "$schema": "https://entur.github.io/ai/schemas/kotlin-spring-boot-stack.json",
  "version": 1,
  "generated_at": "<ISO 8601 UTC timestamp>",
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

### 7. Confirm and prompt commit

Print a one-axis-per-line summary. Then:

> "Stack file written. Commit so teammates pick up the same configuration:
>
>     git add .claude/entur/kotlin-spring-boot.json
>     git commit -m 'chore: initialize kotlin-spring-boot stack file'"

## Constraints

- Do not modify build files, source code, or anything outside `.claude/entur/`
- Read values from the actual repo files; do not infer from training data
- If a build file cannot be read or parsed, ask the user rather than guessing
- Never write secrets, internal URLs, or environment-specific values into this file
