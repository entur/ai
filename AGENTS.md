# Entur AI Agent Instructions

> **Audience:** AI coding agents. Entur employees using or contributing to this repo should start at [README.md](README.md).

Entur is a Norwegian public transportation company. All code targets Google Cloud Platform (GKE), follows Entur platform conventions, and uses shared tooling.

## Platform Context

- **Cloud**: GCP, region `europe-west1`
- **Orchestration**: Kubernetes on GKE
- **CI/CD**: GitHub Actions with Entur reusable workflows
- **Registry**: Google Artifact Registry
- **IaC**: Terraform with Entur modules
- **Deployment**: Helm charts using Entur `common` chart
- **Environments**: `dev`, `tst`, `prd`
- **Languages**: Java 25+/Kotlin (majority), Go, Python
- **Frameworks**: Spring Boot (Java/Kotlin), standard library (Go)
- **Build**: Gradle (Java/Kotlin), Go modules, pip/poetry (Python)
- **License**: EUPL v1.2

## Key Concepts

- **App ID** (`metadata.id` in self-service manifest): 3--10 char alphanumeric identifier, unique across Entur. The Platform Orchestrator creates GCP projects named `ent-{appid}-{env}` (e.g. `metadata.id: products` → `ent-products-dev`, `ent-products-prd`). Data projects use `ent-data-{appid}-{int|ext}-{env}`. Used as Helm `shortname`, Terraform `app_id`, and Terraform state bucket `ent-gcs-tfa-{appid}`. See [self-service.md](guides/platform/self-service.md#gcp-project-naming).
- **App Name** (`metadata.name` in self-service manifest): Becomes the Kubernetes namespace. Typically matches the repository name. Different from App ID.
- **Environments**: `dev`, `tst`, `prd` -- each gets its own GCP project (`ent-{appid}-dev`, `ent-{appid}-tst`, `ent-{appid}-prd`).

## Golden Path

Repository name = application name = Docker image name = Kubernetes namespace = Helm release name.

- Terraform: `./terraform/`, env configs in `./terraform/env/`
- Helm: `./helm/<repo-name>/`, env values in `./helm/<repo-name>/env/`
- Dockerfile at repository root
- CI: `.github/workflows/ci.yml`, CD: `.github/workflows/cd.yml`
- Security allowlists: `.entur/security/`
- Documentation: `./guides/`
- Conventional Commits format (`<type>(<scope>): <description>`, e.g. `feat(api): add endpoint`)

## How to Use This Repository

Centralized AI agent instructions. Teams reference from their `AGENTS.md`:

```markdown
# Project-specific instructions

See https://github.com/entur/ai for Entur-wide standards.

## Overrides and additions
<!-- project-specific instructions here -->
```

## Documentation Map

`guides/` is organised in three layers:

- **[`guides/platform/`](guides/platform/)** -- what the platform provides (self-service orchestrator, common Helm chart, reusable Actions workflows, Terraform modules, Permission Store, Kafka starter).
- **[`guides/playbooks/`](guides/playbooks/)** -- end-to-end tasks (bootstrap, add a database, deploy to prd, etc). Prefer these as your entry point for multi-step work.
- **[`guides/reference/`](guides/reference/)** -- language and topic standards (Java, Kotlin, Go, Docker, observability, logging, tracing, profiler, security, code review, docs).

Always read [CONVENTIONS.md](CONVENTIONS.md) first for cross-cutting standards (naming, repo layout, testing, conventional commits).

### By goal (start here)

| I want to… | Read |
|------------|------|
| Bootstrap a new service | [playbooks/bootstrap-service.md](guides/playbooks/bootstrap-service.md) |
| Set up CI/CD workflows | Load skill: [`skills/setup-cicd-workflows/SKILL.md`](skills/setup-cicd-workflows/SKILL.md) |
| Provision Postgres | [playbooks/add-postgres.md](guides/playbooks/add-postgres.md) |
| Add Redis (caching, locks, dedup) | [playbooks/add-redis.md](guides/playbooks/add-redis.md) |
| Produce or consume Kafka events | [playbooks/add-kafka.md](guides/playbooks/add-kafka.md) |
| Add authentication and authorization | [playbooks/set-up-auth.md](guides/playbooks/set-up-auth.md) |
| Expose service on a custom `*.entur.{no,io,org}` domain | [playbooks/add-custom-domain.md](guides/playbooks/add-custom-domain.md) |
| Promote a service to prd | [playbooks/deploy-to-prd.md](guides/playbooks/deploy-to-prd.md) |
| Deprecate or delete a service | [playbooks/deprecate-service.md](guides/playbooks/deprecate-service.md) |
| Run the service locally | [playbooks/local-dev.md](guides/playbooks/local-dev.md) |

### Platform capabilities (what the platform provides)

| Capability | Doc |
|------------|-----|
| Self-service GCP provisioning | [platform/self-service.md](guides/platform/self-service.md) |
| Common Helm chart | [platform/common-helm.md](guides/platform/common-helm.md) |
| Reusable GitHub Actions workflows | [platform/gha-workflows.md](guides/platform/gha-workflows.md) |
| Composite GHA actions | [platform/gha-actions.md](guides/platform/gha-actions.md) |
| Terraform modules (init, SQL, Redis, GCS) | [platform/terraform-modules.md](guides/platform/terraform-modules.md) |
| Terraform state management | [platform/state-management.md](guides/platform/state-management.md) |
| Allowed IAM roles | [platform/iam-roles.md](guides/platform/iam-roles.md) |
| Permission Store + Permission Client | [platform/permission-store.md](guides/platform/permission-store.md) |
| Entur Kafka Spring starter (Aiven) | [platform/entur-kafka-starter.md](guides/platform/entur-kafka-starter.md) |

### Reference (language and topic standards)

| Topic | Doc |
|-------|-----|
| Java code | [reference/java.md](guides/reference/java.md) |
| Kotlin code | [reference/kotlin.md](guides/reference/kotlin.md) |
| Go code | [reference/go.md](guides/reference/go.md) |
| Docker / containerization | [reference/docker.md](guides/reference/docker.md) |
| API design | [reference/api-design.md](guides/reference/api-design.md) |
| Architecture principles | [reference/architecture.md](guides/reference/architecture.md) |
| SQL / Postgres | [reference/sql.md](guides/reference/sql.md) |
| Logging | [reference/logging.md](guides/reference/logging.md) |
| Observability (metrics, probes) | [reference/observability.md](guides/reference/observability.md) |
| Distributed tracing (OpenTelemetry, Cloud Trace) | [reference/tracing.md](guides/reference/tracing.md) |
| Cloud Profiler (CPU, heap) | [reference/profiler.md](guides/reference/profiler.md) |
| Security | [reference/security.md](guides/reference/security.md) |
| Code review | [reference/code-review.md](guides/reference/code-review.md) |
| Markdown format | [reference/markdown.md](guides/reference/markdown.md) |
| Writing documentation | [reference/documentation.md](guides/reference/documentation.md) |
| UI / Design system / Branding | Load skill: `https://raw.githubusercontent.com/entur/design-system/main/skills/entur-linje/SKILL.md` |

## Critical Rules

1. **ALWAYS use Google Secret Manager** + ExternalSecrets in Helm for all secrets. Never hardcode secrets.
2. **ALWAYS use roles from the [allowed list](guides/platform/iam-roles.md).** Never grant IAM roles outside it. Request additions in `#talk-utviklerplattform`.
3. **ALWAYS use Entur Terraform modules** (`terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`).
4. **ALWAYS use Entur reusable GitHub Actions workflows** for all CI/CD steps.
5. **ALWAYS use the Entur `common` Helm chart** for K8s deployments.
6. **ALWAYS pin all dependencies** -- Terraform (`?ref=TAG`), Actions (`@vN`), Docker images (specific tag).
7. **All services ALWAYS include** health checks, structured logging, and Prometheus metrics.
8. **Default region**: `europe-west1`.
9. **Conventional commits** -- enables automated semver via release-please.
10. **Every PR ALWAYS passes**: lint, unit tests, security scan (CodeQL + Docker scan), Helm lint.
11. **ALWAYS create GCP projects via self-service YAML manifests** in `.entur/` (`GoogleCloudApplication`, `GoogleCloudFirebaseApplication`, `GoogleCloudDataProject`). Never use Terraform `google_project` or `gcloud projects create`. See [self-service.md](guides/platform/self-service.md). For help, ask in `#talk-utviklerplattform`.
