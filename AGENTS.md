# Entur AI Agent Instructions

Entur is a Norwegian public transportation company. All code targets Google Cloud Platform (GKE), follows Entur platform conventions, and uses shared tooling.

## Platform Context

- **Cloud**: GCP, region `europe-west1`
- **Orchestration**: Kubernetes on GKE
- **CI/CD**: GitHub Actions with Entur reusable workflows
- **Registry**: Google Artifact Registry
- **IaC**: Terraform with Entur modules
- **Deployment**: Helm charts using Entur `common` chart
- **Environments**: `dev`, `tst`, `prd`
- **Languages**: Java 21+/Kotlin (majority), Go, Python
- **Frameworks**: Spring Boot (Java/Kotlin), standard library (Go)
- **Build**: Gradle (Java/Kotlin), Go modules, pip/poetry (Python)
- **License**: EUPL v1.2

## Golden Path

Repository name = application name = Docker image name = Kubernetes namespace = Helm release name.

- Terraform: `./terraform/`, env configs in `./terraform/env/`
- Helm: `./helm/<repo-name>/`, env values in `./helm/<repo-name>/env/`
- Dockerfile at repository root
- CI: `.github/workflows/ci.yml`, CD: `.github/workflows/cd.yml`
- Security allowlists: `.entur/security/`
- Documentation: `./docs/`
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
| **Java/Kotlin code** | [java.md](docs/java.md), [kotlin.md](docs/kotlin.md) |
| **Go code** | [go.md](docs/go.md) |
| **API design** | [api-design.md](docs/api-design.md) |
| **Architecture** | [architecture.md](docs/architecture.md) |
| **Kafka** | [kafka.md](docs/kafka.md) |
| **Authorization** | [authorization.md](docs/authorization.md) |
| **Terraform / GCP** | [terraform/modules.md](docs/terraform/modules.md), [terraform/iam-roles.md](docs/terraform/iam-roles.md) |
| **Helm / K8s deploy** | [helm.md](docs/helm.md) |
| **Docker** | [docker.md](docs/docker.md) |
| **CI/CD** | [cicd/workflows.md](docs/cicd/workflows.md), [cicd/actions.md](docs/cicd/actions.md) |
| **Self-service** | [self-service.md](docs/self-service.md) |
| **Firebase** | [cicd/workflows.md](docs/cicd/workflows.md) (gha-firebase section) |
| **Logging** | [logging.md](docs/logging.md) |
| **Observability** | [observability.md](docs/observability.md) |
| **Security** | [security.md](docs/security.md) |
| **Code review** | [code-review.md](docs/code-review.md) |
| **Markdown format** | [markdown.md](docs/markdown.md) |
| **Writing docs** | [documentation.md](docs/documentation.md), [markdown.md](docs/markdown.md) |

## Critical Rules

1. **Never hardcode secrets.** Use Google Secret Manager + ExternalSecrets in Helm.
2. **Never grant IAM roles outside the [allowed list](docs/terraform/iam-roles.md).** Request additions in `#talk-utviklerplattform`.
3. **Always use Entur Terraform modules** (`terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`) -- not raw `google_*` resources.
4. **Always use Entur reusable GitHub Actions workflows** -- not custom CI/CD steps.
5. **Always use the Entur `common` Helm chart** for K8s deployments.
6. **Pin all dependencies** -- Terraform (`?ref=TAG`), Actions (`@vN`), Docker images (specific tag).
7. **All services need** health checks, structured logging, and Prometheus metrics.
8. **Default region**: `europe-west1`.
9. **Conventional commits** -- enables automated semver via release-please.
10. **Every PR must pass**: lint, unit tests, security scan (CodeQL + Docker scan), Helm lint.
