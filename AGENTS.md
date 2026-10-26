# Entur AI Agent Instructions

> **Audience:** AI coding agents. Entur employees using or contributing to this repo should start at [README.md](README.md).

Entur is a Norwegian public transportation company. All code targets Google Cloud Platform (GKE), follows Entur platform conventions, and uses shared tooling.

## Agent Reading Order

1. Read [CONVENTIONS.md](CONVENTIONS.md) for cross-cutting standards.
2. Select the task-specific guide, playbook, reference, or skill from [Documentation Map](#documentation-map).
3. Read the relevant reference for the code or configuration you will change: Java, Kotlin, Go, Docker, SQL, API design, security, observability, or documentation.
4. When work involves introducing, recommending, configuring, documenting, or automating IT systems or software, read [it-systems-policy.md](it-systems-policy.md) before making recommendations.
5. Treat [Critical Rules](#critical-rules) as hard constraints. Project-specific `AGENTS.md` files may add stricter rules, but they must not weaken platform, security, or policy requirements from this repo.

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

| Concept | Source | Use |
|---------|--------|-----|
| **App ID** | `metadata.id` in the self-service manifest. 3--10 char alphanumeric identifier, unique across Entur. | Use for GCP project suffixes (`ent-{appid}-{env}`), Helm `shortname`, Terraform `app_id`, and Terraform state bucket `ent-gcs-tfa-{appid}`. Data projects use `ent-data-{appid}-{int\|ext}-{env}`. Do not use App Name for these values. See [self-service.md](guides/platform/self-service.md#gcp-project-naming). |
| **App Name** | `metadata.name` in the self-service manifest. Typically matches the repository name. | Use for the Kubernetes namespace. Do not use App ID as the Kubernetes namespace unless `metadata.id` and `metadata.name` are intentionally identical. |
| **Environments** | `dev`, `tst`, `prd` | Standard environments. Each environment gets its own GCP project, for example `ent-{appid}-dev`, `ent-{appid}-tst`, and `ent-{appid}-prd`. |

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

Use [Agent Reading Order](#agent-reading-order) first, then use the tables below to find the task-specific document.

### By goal (start here)

| I want to… | Read |
|------------|------|
| Bootstrap a new service | [playbooks/bootstrap-service.md](guides/playbooks/bootstrap-service.md) |
| Set up CI/CD workflows | Load skill: [`skills/setup-cicd-workflows/SKILL.md`](skills/setup-cicd-workflows/SKILL.md) |
| Set up AI agent config for an existing repo | Load skill: [`skills/setup-agent-config/SKILL.md`](skills/setup-agent-config/SKILL.md) |
| Introduce or recommend IT systems and software | [it-systems-policy.md](it-systems-policy.md) |
| Provision Postgres | [playbooks/add-postgres.md](guides/playbooks/add-postgres.md) |
| Add Redis (caching, locks, dedup) | [playbooks/add-redis.md](guides/playbooks/add-redis.md) |
| Produce or consume Kafka events | [playbooks/add-kafka.md](guides/playbooks/add-kafka.md) |
| Add authentication and authorization | [playbooks/set-up-auth.md](guides/playbooks/set-up-auth.md) |
| Expose service on a custom `*.entur.{no,io,org}` domain | [playbooks/add-custom-domain.md](guides/playbooks/add-custom-domain.md) |
| Promote a service to prd | [playbooks/deploy-to-prd.md](guides/playbooks/deploy-to-prd.md) |
| Deprecate or delete a service | [playbooks/deprecate-service.md](guides/playbooks/deprecate-service.md). ALWAYS warn the user first: "For 30 days after an app is deleted, the same App ID cannot be used again." |
| Handle a critical incident | [playbooks/incident-response.md](guides/playbooks/incident-response.md) |
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
| Alerting (Grafana, PromQL, PagerDuty) | [reference/alerting.md](guides/reference/alerting.md) |
| Security | [reference/security.md](guides/reference/security.md) |
| Code review | [reference/code-review.md](guides/reference/code-review.md) |
| Markdown format | [reference/markdown.md](guides/reference/markdown.md) |
| Writing documentation | [reference/documentation.md](guides/reference/documentation.md) |
| UI / Design system / Branding | Use the `entur-linje` skill if the `entur-linje` plugin is installed. Otherwise load `https://raw.githubusercontent.com/entur/design-system/main/skills/entur-linje/SKILL.md` |

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
12. **ALWAYS follow [it-systems-policy.md](it-systems-policy.md)** before introducing, recommending, configuring, documenting, or automating IT systems or software. Never claim a system is registered, classified, or approved unless the repository, user, or DAP documentation confirms it.
13. **ALWAYS create a new branch for infrastructure changes.** Before you create or change any file in `.github/`, `.entur/`, `terraform/`, or `helm/`, ALWAYS create a new branch from the latest default branch (`git fetch origin && git switch -c <type>/<short-description> origin/main`, e.g. `feat/add-postgres`). Never make these changes on `main`.
14. **ALWAYS open a pull request for changes in `.entur/`.** When you create or change any file in `.entur/`, commit with a conventional commit message, push, and open a PR with `gh pr create`. Never commit or push `.entur/` changes directly to `main`. The Platform Orchestrator only plans and applies manifests through a PR. Only comment `entur apply` when the user explicitly allows it for that PR: first show the user the orchestrator's plan, then run `gh pr comment <pr> --body "entur apply"`. Otherwise tell the user to review the plan and comment it. See [self-service.md](guides/platform/self-service.md#how-it-works).
15. **ALWAYS merge `.entur/` before creating `.github/` in a new repository.** When `.entur/` does not exist on `main`, open the `.entur/` PR first and wait until it is applied and merged (`git fetch origin && git ls-tree -d --name-only origin/main .entur` prints `.entur`). Only then create `.github/` on a new branch from the updated `main`. The pipelines authenticate through resources the manifests provision and fail until they exist.
16. **ALWAYS warn before changing App ID or App Name.** When someone tries to change `metadata.id` or `metadata.name` in an existing `.entur/` manifest, stop and show this message word for word before making any change, and do not edit either field until the user confirms they have talked to the Platform team:

    > ⚠️ **Warning:** Do not perform these actions without consulting the Platform team first. They can have serious consequences, including the deletion of critical application resources such as service accounts and databases.
    >
    > Please contact the Platform team in **#talk-utviklingsplattform** before proceeding.
