# Entur AI Agent Instructions

You are helping develop software at Entur, a Norwegian public transportation company. All code must follow Entur's platform conventions, use shared tooling, and target Google Cloud Platform (GKE).

## Platform Context

- **Cloud**: Google Cloud Platform (GCP)
- **Orchestration**: Kubernetes on GKE (Google Kubernetes Engine), region `europe-west1`
- **CI/CD**: GitHub Actions with Entur reusable workflows
- **Container Registry**: Google Artifact Registry
- **Infrastructure as Code**: Terraform with Entur modules
- **Deployment**: Helm charts using Entur's `common` chart
- **Environments**: `dev`, `tst`, `prd`
- **Primary languages**: Java 21+ / Kotlin (majority), Go, Python
- **Frameworks**: Spring Boot (Java/Kotlin), standard library (Go)
- **Build tools**: Gradle (Java/Kotlin), Go modules, pip/poetry (Python)
- **License**: EUPL v1.2

## Golden Path Conventions

Entur follows a "golden path" (convention-over-configuration) approach:

- **Repository name = application name = Docker image name = Kubernetes namespace = Helm release name**
- Terraform files live in `./terraform/` with env-specific configs in `./terraform/env/`
- Helm chart lives in `./helm/<repo-name>/` with env-specific values in `./helm/<repo-name>/env/`
- Dockerfile lives at repository root
- CI workflow is `.github/workflows/ci.yml`, CD workflow is `.github/workflows/cd.yml`
- Security scan allowlists live in `.entur/security/`
- Documentation lives in `./docs/`
- Use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages

## How to Use This Repository

This repository contains centralized instructions for AI coding agents. Teams reference this repository (or specific files) from their own `AGENTS.md` to inherit shared standards.

**In your project's AGENTS.md**, reference like this:

```markdown
# Project-specific instructions

See https://github.com/entur/ai for Entur-wide standards.

## Overrides and additions
<!-- project-specific instructions here -->
```

## Documentation Map

Read the relevant documents below based on the task at hand. Always read `CONVENTIONS.md` first for cross-cutting standards.

### Always Read

- [CONVENTIONS.md](CONVENTIONS.md) -- Cross-language coding conventions, naming, error handling, testing

### By Task Type

| Task | Read these documents |
|------|---------------------|
| **Java/Kotlin application code** | [docs/java.md](docs/java.md), [docs/kotlin.md](docs/kotlin.md) |
| **Go application code** | [docs/go.md](docs/go.md) |
| **API design** | [docs/api-design.md](docs/api-design.md) |
| **Architecture decisions** | [docs/architecture.md](docs/architecture.md) |
| **Kafka / event streaming** | [docs/kafka.md](docs/kafka.md) |
| **Authorization / permissions** | [docs/authorization.md](docs/authorization.md) |
| **Terraform / GCP infrastructure** | [docs/terraform/modules.md](docs/terraform/modules.md), [docs/terraform/iam-roles.md](docs/terraform/iam-roles.md) |
| **Helm charts / Kubernetes deploy** | [docs/helm.md](docs/helm.md) |
| **Docker / containerization** | [docs/docker.md](docs/docker.md) |
| **CI/CD pipelines** | [docs/cicd/workflows.md](docs/cicd/workflows.md), [docs/cicd/actions.md](docs/cicd/actions.md) |
| **Self-service provisioning** | [docs/self-service.md](docs/self-service.md) |
| **Firebase Hosting** | [docs/cicd/workflows.md](docs/cicd/workflows.md) (gha-firebase section) |
| **Logging** | [docs/logging.md](docs/logging.md) |
| **Observability** | [docs/observability.md](docs/observability.md) |
| **Security** | [docs/security.md](docs/security.md) |
| **Code review** | [docs/code-review.md](docs/code-review.md) |
| **Markdown / documentation format** | [docs/markdown.md](docs/markdown.md) |
| **Writing documentation** | [docs/documentation.md](docs/documentation.md), [docs/markdown.md](docs/markdown.md) |

## Critical Rules

1. **Never hardcode secrets.** Use Google Secret Manager and reference via ExternalSecrets in Helm.
2. **Never grant IAM roles not in the [allowed list](docs/terraform/iam-roles.md).** Request additions in `#talk-utviklerplattform` on Slack.
3. **Always use Entur shared Terraform modules** (`terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`) instead of raw `google_*` resources for managed services.
4. **Always use Entur reusable GitHub Actions workflows** instead of writing custom CI/CD steps for build, test, scan, deploy.
5. **Always use the Entur `common` Helm chart** as a dependency for Kubernetes deployments.
6. **Pin all external dependencies** -- Terraform modules (`?ref=TAG`), GitHub Actions (`@vN`), Docker base images (digest or specific tag).
7. **All services must have health checks**, structured logging, and Prometheus metrics.
8. **Use `europe-west1`** as the default GCP region.
9. **Follow conventional commits** -- this enables automated semantic versioning via release-please.
10. **Every PR must pass**: lint, unit tests, security scan (CodeQL + Docker scan), and Helm lint before merge.
