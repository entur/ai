---
name: conventions
description: >
  Apply when editing .kt/.kts Kotlin sources, build.gradle.kts, gradle/libs.versions.toml,
  or pom.xml in a Spring Boot repo. Provides Entur Kotlin Spring Boot conventions:
  dependencies, controllers, services, testing, logging, build, config. Stack axes
  (mvc vs webflux, jpa vs exposed vs jdbc, mockk vs mockito, kotest vs assertj, formatter)
  read from .claude/entur/kotlin-spring-boot.json. Do NOT invoke for plain Java (.java),
  non-Spring Kotlin (Ktor, Android, CLI), or repos without spring-boot-starter-* on the
  classpath.
---

# Kotlin Spring Boot Conventions

## Activation

Apply only when a Spring Boot signal is present:

- `build.gradle.kts` / `build.gradle` — `org.springframework.boot` plugin or any `spring-boot-starter-*` dependency
- `pom.xml` — `spring-boot-starter-parent` or any `spring-boot-starter-*` dependency
- `gradle/libs.versions.toml` — `spring-boot` version entry

If absent, exit silently. Ktor, plain Kotlin, and Java repos are out of scope.

## Stack file

Single source of truth for axes that vary across Entur Kotlin services: `.claude/entur/kotlin-spring-boot.json` (committed, team-shared).

Schema (version 1):

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

Treat values as authoritative. Honour `notes` as binding context — if the team documented why a non-default choice exists, do not push back against it.

`legacy_mode: true` suppresses modernization advice. Do not propose Java toolchain upgrades, dependency catalog migrations, framework version bumps, or pattern rewrites unless the user explicitly asks. Match the existing style; fix bugs, add tests, make small additions.

If `version > 1`, stop applying conventions and tell the user the plugin is older than the stack file and must be upgraded.

### Missing stack file

If `.claude/entur/kotlin-spring-boot.json` is absent, fall back to detecting axes from build files using the mapping in [`commands/init.md`](../../commands/init.md). Tell the user once per session:

> `.claude/entur/kotlin-spring-boot.json` is missing. Run `/kotlin-spring-boot:init` to generate it. The file is committed and shared with the team.

Proceed with detected values. Do not block.

## Tooling

- Kotlin LSP: `brew install kotlin-language-server` (plugin ships LSP config)
- `ktlint`: `brew install ktlint` — auto-runs on Write/Edit when `formatter=ktlint`
- Spotless: configured in the repo's `build.gradle.kts` or `pom.xml`, not the plugin

## Reference loading

Load only the reference matching the current task. Inside each reference, apply only sections matching the active stack axes; skip inactive options.

| Task | Reference |
|---|---|
| Build config, dependencies, version catalog, `build.gradle.kts`, `pom.xml`, Artifactory | `references/build.md` |
| Controllers, services, mappers, exception handling, validation | `references/patterns.md` |
| Database access (Exposed, Spring Data JDBC, JPA, Flyway, migrations) | `references/database.md` |
| Tests, mocking, assertions, TestContainers, integration tests | `references/testing.md` |
| Logging, cloud-logging, request-response logging | `references/logging.md` |
| `application.yml`, Cloud SQL, HikariCP, Redis, Flyway config | `references/config.md` |

## Kotlin language principles

Apply to all Kotlin code regardless of stack:

- Primary constructor injection — never `@Autowired` field injection
- `val` over `var`
- Kotlin null-safety (`T?`) — never `Optional<T>`
- `data class` for domain models and value objects
- `sealed class` / `sealed interface` for restricted hierarchies
- `object` for stateless singletons (validators, constants)
- `internal` visibility for implementation details
- `when` over `if-else` chains with multiple branches
- Trailing commas on multi-line argument and parameter lists
- Scope functions (`let`, `apply`, `run`, `also`) where they clarify intent — not for brevity alone
- Backtick test names: `` fun `should return 404 when route not found`() ``
- `@DisplayName` on test classes when backtick names get unwieldy

## Non-negotiable Entur rules

Apply to every Kotlin Spring Boot project regardless of stack. These remain in this plugin so they reach developers who install only `kotlin-spring-boot`.

1. Secrets: Google Secret Manager + ExternalSecrets in Helm. Never hardcoded.
2. IAM roles: only from the Entur approved list. Other roles require platform-team approval (`#talk-utviklerplattform`).
3. Terraform: Entur modules only — `terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`.
4. CI/CD: Entur reusable GitHub Actions workflows only. No custom CI steps.
5. K8s deploys: Entur `common` Helm chart.
6. Pin all dependencies — Gradle version catalog or Maven `<dependencyManagement>`; Terraform `?ref=TAG`; Actions `@vN`; Docker images by exact tag.
7. Every service: health checks (`/actuator/health/liveness`, `/actuator/health/readiness`), structured logging (`entur/cloud-logging`), Prometheus metrics.
8. Default GCP region: `europe-west1`.
9. Conventional commits — required for release-please semver automation.
10. Every PR passes: Ktlint/Spotless, unit tests, CodeQL, Helm lint.
11. GCP projects: created via `.entur/*.yaml` self-service manifests only. Never `google_project` Terraform resource or `gcloud projects create`.

## Out of scope

Kotlin Spring Boot code only. For CI/CD workflows, Helm, Terraform, Kafka, security, scaffolding, or writing formats, point the user to the Entur marketplace at https://github.com/entur/ai (or `/plugin marketplace add entur/ai`) and let them pick the plugin matching the topic.
