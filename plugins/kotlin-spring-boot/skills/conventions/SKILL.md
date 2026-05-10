---
name: conventions
description: >
  Entur Kotlin Spring Boot conventions: dependencies, controllers, services,
  testing, logging, build, config. Triggered when editing .kt/.kts, build.gradle.kts,
  gradle/libs.versions.toml, or pom.xml in a Spring Boot repo. Stack axes
  (mvc/webflux, jpa/exposed/jdbc, mockk/mockito, kotest/assertj, formatter)
  read from .claude/entur/kotlin-spring-boot.json. Skip for plain Java (.java),
  non-Spring Kotlin (Ktor, Android, CLI), and repos without spring-boot-starter-*.
---

# Kotlin Spring Boot Conventions

## Activation

Apply only if the repo has a Spring Boot signal:

- `org.springframework.boot` Gradle plugin or any `spring-boot-starter-*` dependency
- `pom.xml` with `spring-boot-starter-parent` or any `spring-boot-starter-*`
- `gradle/libs.versions.toml` with a `spring-boot` version

Otherwise exit silently.

## Stack file

Read `.claude/entur/kotlin-spring-boot.json` (committed, team-shared). Schema:

```json
{
  "version": 1,
  "build_tool": "gradle | maven",
  "spring_stack": "mvc | webflux",
  "api_approach": "contract-first | traditional",
  "database": "exposed | spring-data-jdbc | jpa | none",
  "test_mocking": "mockk | mockito-kotlin",
  "test_assertions": "kotest | assertj",
  "formatter": "ktlint | spotless-gradle | spotless-maven | none",
  "legacy_mode": false,
  "notes": ""
}
```

Use values as-is. Respect `notes` — if the team explained a non-default choice, don't argue with it.

`legacy_mode: true` means: no toolchain upgrades, no version catalog migrations, no framework bumps, no pattern rewrites. Match existing style; fix bugs, add tests, make small additions.

`version > 1`: stop. The plugin is older than the stack file and must be upgraded.

### When the stack file is missing

Detect axes from build files using the table in [`commands/init.md`](../../commands/init.md). Once per session, tell the user:

> `.claude/entur/kotlin-spring-boot.json` is missing. Run `/kotlin-spring-boot:init` to generate it.

Then proceed. Don't block.

## Tooling

- Kotlin LSP: `brew install kotlin-language-server` (plugin ships LSP config)
- ktlint: `brew install ktlint` — auto-runs on save when `formatter=ktlint`
- Spotless: configured in the repo's build files, not by the plugin

## References

Load the reference matching the file you're editing. Inside it, apply sections matching the active stack axes — skip the rest.

| When editing | Read |
|---|---|
| `build.gradle.kts`, `build.gradle`, `pom.xml`, `gradle/libs.versions.toml`, `settings.gradle.kts` | `references/build.md` |
| `*.kt` controllers, services, mappers, exception handlers, validators | `references/patterns.md` |
| DAOs, repositories, `@Entity`, `@Table`, `*.sql` migrations | `references/database.md` |
| `**/test/**/*.kt`, `*Test.kt`, `*IntegrationTest.kt`, `TestContainersConfig` | `references/testing.md` |
| Code using SLF4J `Logger`/`LoggerFactory` or `entur.logging.*` config | `references/logging.md` |
| `application.yml`, `application-*.yml` | `references/config.md` |

## Kotlin principles

- Primary constructor injection. Never `@Autowired` field injection.
- `val` over `var`
- `T?` over `Optional<T>`
- `data class` for domain types
- `sealed class` / `sealed interface` for closed hierarchies
- `object` for stateless singletons
- `internal` for non-public API
- `when` over long `if/else` chains
- Trailing commas on multi-line lists
- Scope functions (`let`, `apply`, `run`, `also`) only when they clarify intent
- Backtick test names: `` fun `returns 404 when route not found`() ``
- `@DisplayName` on test classes when backtick names get unwieldy

## Non-negotiable Entur rules

These apply to every project. Kept here so devs who install only this plugin still see them.

1. Secrets: Google Secret Manager + ExternalSecrets in Helm. Never hardcoded.
2. IAM roles: only from the Entur approved list. Anything else needs platform-team approval (`#talk-utviklerplattform`).
3. Terraform modules: `terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`.
4. CI/CD: Entur reusable GitHub Actions workflows. No custom CI steps.
5. K8s: Entur `common` Helm chart.
6. Pin everything — Gradle catalog or Maven `<dependencyManagement>`, Terraform `?ref=TAG`, Actions `@vN`, Docker by exact tag.
7. Every service: liveness + readiness probes, structured logging via `entur/cloud-logging`, Prometheus metrics.
8. GCP region: `europe-west1`.
9. Conventional commits (release-please needs them).
10. PRs must pass: Ktlint/Spotless, unit tests, CodeQL, Helm lint.
11. GCP projects: `.entur/*.yaml` self-service manifests only. Never `google_project` Terraform or `gcloud projects create`.

## Out of scope

Kotlin Spring Boot code. For CI/CD, Helm, Terraform, Kafka, security, scaffolding, or writing formats, send the user to the Entur marketplace at https://github.com/entur/ai (`/plugin marketplace add entur/ai`).
