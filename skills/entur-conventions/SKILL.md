---
name: entur-conventions
description: >
  Entur platform conventions for code, infrastructure, deployment, and operations.
  Activate when editing or asking about: Kotlin/Java/Go/Python services, Spring Boot,
  Gradle, Helm charts, Dockerfiles, Terraform (Entur modules, IAM roles), Kafka,
  Pub/Sub, GCP, GKE, Cloud SQL, Memorystore, OAuth/OIDC, Permission Store,
  observability/logging/metrics, secrets and security scanning, GitHub Actions
  workflows, self-service `.entur/*.yaml` manifests, code review, or markdown style
  for an Entur application. Fetches the matching guide from the entur/ai repo on
  demand instead of bundling content.
---

# Entur Conventions (Router)

This skill is a thin router. It does not bundle convention content — it fetches the matching guide from the `entur/ai` repository when activated.

## Critical rules (always apply, no fetch needed)

These rules are non-negotiable and apply to every Entur application:

1. ALWAYS use Google Secret Manager + ExternalSecrets in Helm. Never hardcode secrets.
2. ALWAYS use IAM roles from the approved list. Never grant roles outside it. Request additions in `#talk-utviklerplattform`.
3. ALWAYS use Entur Terraform modules: `terraform-google-init`, `terraform-google-sql-db`, `terraform-google-memorystore`, `terraform-google-cloud-storage`.
4. ALWAYS use Entur reusable GitHub Actions workflows. Never write custom CI steps.
5. ALWAYS use the Entur `common` Helm chart for K8s deployments.
6. ALWAYS pin dependencies — Terraform `?ref=TAG`, Actions `@vN`, Docker images by specific tag.
7. Every service includes health checks, structured logging, Prometheus metrics.
8. Default region: `europe-west1`.
9. Conventional commits — enables automated semver via release-please.
10. Every PR passes lint, unit tests, security scan (CodeQL + Docker scan), Helm lint.
11. ALWAYS create GCP projects via self-service YAML manifests in `.entur/` (`GoogleCloudApplication`, `GoogleCloudFirebaseApplication`, `GoogleCloudDataProject`). Never use Terraform `google_project` or `gcloud projects create`.

## How to use this skill

Pick the guide that matches the task and fetch it. URL pattern:

```text
https://raw.githubusercontent.com/entur/ai/main/guides/<path>.md
```

Cross-language baseline always applies — fetch first when starting a non-trivial task:

```text
https://raw.githubusercontent.com/entur/ai/main/CONVENTIONS.md
```

## Guide index

| Topic | Path | When to fetch |
|-------|------|--------------|
| Kotlin services | `guides/kotlin.md` | editing `.kt`, `build.gradle.kts`, designing a Kotlin Spring Boot service |
| Java services | `guides/java.md` | editing `.java`, JVM patterns shared with Kotlin |
| Go services | `guides/go.md` | editing `.go`, `go.mod` |
| API design | `guides/api-design.md` | designing REST or gRPC contracts |
| Architecture | `guides/architecture.md` | service boundaries, GCP project structure, resilience, production hardening |
| Authorization | `guides/authorization.md` | Permission Store, Permission Client, `@PreAuthorize` |
| Kafka | `guides/kafka.md` | producers, consumers, Avro/Protobuf schemas, Aiven Kafka |
| Helm | `guides/helm.md` | `helm/<app>/`, common chart, ExternalSecrets values |
| Docker | `guides/docker.md` | `Dockerfile`, multi-stage builds, distroless |
| Terraform modules | `guides/terraform/modules.md` | `*.tf`, Entur Terraform modules (init, SQL, Redis, GCS) |
| IAM roles | `guides/terraform/iam-roles.md` | granting any IAM role — only roles on this list are approved |
| CI/CD workflows | `guides/cicd/workflows.md` | `.github/workflows/*.yml`, Entur reusable workflows |
| CI/CD actions | `guides/cicd/actions.md` | composite GitHub Actions |
| Self-service | `guides/self-service.md` | `.entur/*.yaml`, GCP project naming, `metadata.id` vs `metadata.name` |
| Logging | `guides/logging.md` | structured JSON logs, traceId, requestId |
| Observability | `guides/observability.md` | health endpoints, Prometheus metrics, tracing |
| Security | `guides/security.md` | secrets, scanning, allowlists in `.entur/security/`, headers |
| Code review | `guides/code-review.md` | reviewing or preparing a PR |
| Markdown | `guides/markdown.md` | `.md` files, markdownlint |
| Documentation | `guides/documentation.md` | writing user-facing prose |

If the topic is not listed but matches a filename pattern under `guides/`, try fetching `https://raw.githubusercontent.com/entur/ai/main/guides/<topic>.md` directly. Contributors add new conventions by adding a markdown file under `guides/`; this skill picks them up automatically.

## Fallback

When the user mentions a tool or concept that is not covered by an Entur guide, fall back to general best practices. Do not guess Entur-specific patterns.
