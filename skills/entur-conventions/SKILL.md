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

`guides/` is split into three layers — pick the right one and fetch:

- **`guides/platform/`** -- platform-provided capabilities (self-service, common chart, reusable workflows, Terraform modules, Permission Store, Kafka starter).
- **`guides/playbooks/`** -- end-to-end tasks (bootstrap, add postgres/redis/kafka, set up auth, deploy to prd). Prefer these as the entry point for multi-step work.
- **`guides/reference/`** -- language and topic standards (Java, Kotlin, Go, Docker, observability, security, etc).

URL pattern:

```text
https://raw.githubusercontent.com/entur/ai/main/guides/<layer>/<file>.md
```

Cross-language baseline always applies — fetch first when starting a non-trivial task:

```text
https://raw.githubusercontent.com/entur/ai/main/CONVENTIONS.md
```

## Playbook index (prefer these for multi-step tasks)

| Goal | Path |
|------|------|
| Bootstrap a new service | `guides/playbooks/bootstrap-service.md` |
| Add managed Postgres | `guides/playbooks/add-postgres.md` |
| Add Memorystore Redis | `guides/playbooks/add-redis.md` |
| Produce/consume Kafka events | `guides/playbooks/add-kafka.md` |
| Set up authentication/authorization | `guides/playbooks/set-up-auth.md` |
| Promote a service to prd | `guides/playbooks/deploy-to-prd.md` |
| Deprecate or delete a service | `guides/playbooks/deprecate-service.md` |
| Run the service locally | `guides/playbooks/local-dev.md` |

## Platform capability index

| Topic | Path | When to fetch |
|-------|------|--------------|
| Self-service | `guides/platform/self-service.md` | `.entur/*.yaml`, GCP project naming, `metadata.id` vs `metadata.name` |
| Common Helm chart | `guides/platform/common-helm.md` | `helm/<app>/`, common chart, ExternalSecrets values |
| Reusable GHA workflows | `guides/platform/gha-workflows.md` | `.github/workflows/*.yml`, Entur reusable workflows |
| Composite GHA actions | `guides/platform/gha-actions.md` | custom steps using `entur/gha-meta` actions |
| Terraform modules | `guides/platform/terraform-modules.md` | `*.tf`, Entur Terraform modules (init, SQL, Redis, GCS) |
| Approved IAM roles | `guides/platform/iam-roles.md` | granting any IAM role — only roles on this list are approved |
| Permission Store | `guides/platform/permission-store.md` | Permission Client, `@PreAuthorize`, business capabilities |
| Entur Kafka starter | `guides/platform/entur-kafka-starter.md` | producers, consumers, Avro/Protobuf, Aiven clusters |

## Reference index (language and topic standards)

| Topic | Path | When to fetch |
|-------|------|--------------|
| Kotlin services | `guides/reference/kotlin.md` | editing `.kt`, `build.gradle.kts`, designing a Kotlin Spring Boot service |
| Java services | `guides/reference/java.md` | editing `.java`, JVM patterns shared with Kotlin |
| Go services | `guides/reference/go.md` | editing `.go`, `go.mod` |
| API design | `guides/reference/api-design.md` | designing REST or gRPC contracts |
| Architecture | `guides/reference/architecture.md` | service boundaries, DB conventions, resilience, lifecycle |
| Docker | `guides/reference/docker.md` | `Dockerfile`, multi-stage builds, distroless |
| Logging | `guides/reference/logging.md` | structured JSON logs, traceId, requestId |
| Observability | `guides/reference/observability.md` | health endpoints, Prometheus metrics, tracing |
| Security | `guides/reference/security.md` | secrets, scanning, allowlists in `.entur/security/`, headers |
| Code review | `guides/reference/code-review.md` | reviewing or preparing a PR |
| Markdown | `guides/reference/markdown.md` | `.md` files, markdownlint |
| Documentation | `guides/reference/documentation.md` | writing user-facing prose |

If the topic is not listed but you suspect a guide exists, try fetching from `guides/platform/<topic>.md`, `guides/playbooks/<topic>.md`, or `guides/reference/<topic>.md` in that order. Contributors add new conventions by adding markdown files under `guides/`; this skill picks them up automatically.

## Fallback

When the user mentions a tool or concept that is not covered by an Entur guide, fall back to general best practices. Do not guess Entur-specific patterns.
