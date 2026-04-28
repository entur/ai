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

- **App ID** (`metadata.id` in self-service manifest): 3--10 char alphanumeric identifier, unique across Entur. The Platform Orchestrator creates GCP projects named `ent-{appid}-{env}` (e.g. `metadata.id: products` → `ent-products-dev`, `ent-products-prd`). Data projects use `ent-data-{appid}-{int|ext}-{env}`. Used as Helm `shortname`, Terraform `app_id`, and Terraform state bucket `ent-gcs-tfa-{appid}`. See [self-service.md](guides/self-service.md#gcp-project-naming).
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
- [Conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)

## How to Use This Repository

Centralized AI agent instructions. Teams reference from their `AGENTS.md`:

```markdown
# Project-specific instructions

See https://github.com/entur/ai for Entur-wide standards.

## Overrides and additions
<!-- project-specific instructions here -->
```

## Documentation Map

Always read `CONVENTIONS.md` first for cross-cutting standards.

### Always Read

- [CONVENTIONS.md](CONVENTIONS.md) -- Cross-language conventions, naming, error handling, testing

### By Task Type

| Task | Documents |
|------|-----------|
| **Java/Kotlin code** | [java.md](guides/java.md), [kotlin.md](guides/kotlin.md) |
| **Go code** | [go.md](guides/go.md) |
| **API design** | [api-design.md](guides/api-design.md) |
| **Architecture** | [architecture.md](guides/architecture.md) |
| **Kafka** | [kafka.md](guides/kafka.md) |
| **Authorization** | [authorization.md](guides/authorization.md) |
| **Terraform / GCP** | [terraform/modules.md](guides/terraform/modules.md), [terraform/iam-roles.md](guides/terraform/iam-roles.md) |
| **Helm / K8s deploy** | [helm.md](guides/helm.md) |
| **Docker** | [docker.md](guides/docker.md) |
| **CI/CD** | [cicd/workflows.md](guides/cicd/workflows.md), [cicd/actions.md](guides/cicd/actions.md) |
| **Self-service** | [self-service.md](guides/self-service.md) |
| **Firebase** | [cicd/workflows.md](guides/cicd/workflows.md) (gha-firebase section) |
| **Logging** | [logging.md](guides/logging.md) |
| **Observability** | [observability.md](guides/observability.md) |
| **Security** | [security.md](guides/security.md) |
| **Code review** | [code-review.md](guides/code-review.md) |
| **Markdown format** | [markdown.md](guides/markdown.md) |
| **Writing docs** | [documentation.md](guides/documentation.md), [markdown.md](guides/markdown.md) |

## Critical Rules

1. **ALWAYS use Google Secret Manager** + ExternalSecrets in Helm for all secrets. Never hardcode secrets.
2. **ALWAYS use roles from the [allowed list](guides/terraform/iam-roles.md).** Never grant IAM roles outside it. Request additions in `#talk-utviklerplattform`.
3. **ALWAYS use Entur Terraform modules** (`terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`).
4. **ALWAYS use Entur reusable GitHub Actions workflows** for all CI/CD steps.
5. **ALWAYS use the Entur `common` Helm chart** for K8s deployments.
6. **ALWAYS pin all dependencies** -- Terraform (`?ref=TAG`), Actions (`@vN`), Docker images (specific tag).
7. **All services ALWAYS include** health checks, structured logging, and Prometheus metrics.
8. **Default region**: `europe-west1`.
9. **Conventional commits** -- enables automated semver via release-please.
10. **Every PR ALWAYS passes**: lint, unit tests, security scan (CodeQL + Docker scan), Helm lint.
11. **ALWAYS create GCP projects via self-service YAML manifests** in `.entur/` (`GoogleCloudApplication`, `GoogleCloudFirebaseApplication`, `GoogleCloudDataProject`). Never use Terraform `google_project` or `gcloud projects create`. See [self-service.md](guides/self-service.md). For help, ask in `#talk-utviklerplattform`.
