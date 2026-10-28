---
name: entur-project-bootstrap
description: >
  Bootstrap a new Entur application from scratch. Generates self-service manifests,
  Helm chart, Terraform scaffolding, Dockerfile, and CI/CD workflows -- all with
  correct cross-file coordination (metadata.id = Helm shortname = Terraform app_id).
  Use this skill when the user says "new project", "bootstrap", "scaffold",
  "create a new app", "set up a new service", or needs to generate the full
  infrastructure setup for a new application at Entur.
---

# Entur Project Bootstrap

Generate the complete infrastructure scaffolding for a new Entur application. This skill coordinates across self-service, Helm, Terraform, Docker, and CI/CD to ensure all identifiers are consistent.

## Asking Questions

ALWAYS ask the user through your interactive question tool (for example `AskUserQuestion` in Claude Code) whenever the answer can be offered as options. Fall back to a plain-text question only when the tool is not available, or when the answer is free text with no sensible options.

- **Choices and yes/no** -- language, environments, ingress, database, Redis, Kafka, Auth0 M2M, confirming the identity chain, permission to run `entur apply`, adding required reviewers. Put the default first and mark it "(Recommended)".
- **Several inputs at once** -- batch related questions into one call (up to four), for example database, Redis, Kafka, and Auth0 M2M together.
- **Multi-select** -- use it when the user may pick more than one option, such as environments.
- **Free-text values** -- repository name, App ID, team, display name, and ingress hosts. When you can suggest a value (for example the current directory name, or a team found in existing manifests), offer it as an option; the user types a different value through the tool's free-text "Other" option. Never offer a guessed ingress host as an option -- show the pattern as a hint and let the user type the value.

## Step 1: Gather Inputs

Ask the user for the following, following [Asking Questions](#asking-questions). Skip inputs the user already gave. If not provided, suggest sensible defaults:

| Input | Required | Constraints | Default |
|-------|----------|-------------|---------|
| **Repository name** | yes | lowercase kebab-case, max 63 chars, `[a-z0-9-]` | -- |
| **App ID** (`metadata.id`) | yes | 3-10 lowercase alphanumeric `^[a-z0-9]+$`, no `ent-` prefix, no env suffixes | -- |
| **Team** | yes | must start with `team-` | -- |
| **Language** | yes | `kotlin`, `java`, `go`, `python` | `kotlin` |
| **Environments** | no | subset of `dev`, `tst`, `prd` | `[dev, tst, prd]` |
| **Needs ingress** | no | boolean | `true` |
| **Ingress host per environment** | yes, unless ingress is disabled | under `entur.no`, `entur.io`, or `entur.org`, e.g. `myapp.dev.entur.io` | -- ALWAYS ask, never derive |
| **Needs database** | no | boolean | `false` |
| **Needs Redis** | no | boolean | `false` |
| **Needs Kafka** | no | boolean | `false` |
| **Needs Auth0 M2M** | no | boolean | `false` |
| **Display name** | no | human-readable name | derived from repo name |

When you ask for **Environments**, ALWAYS ask a multi-select question with `dev`, `tst`, and `prd` as separate options, so the user can pick one, two, or all three. Do not offer fixed bundles such as "dev, tst, prd" or "dev only".

### Validation Rules

Before proceeding, validate:

- `metadata.id` is 3-10 chars, `^[a-z0-9]+$`
- `metadata.id` ALWAYS uses bare identifiers (the platform adds the `ent-` prefix)
- `metadata.id` ALWAYS uses base identifiers (the platform adds environment suffixes)
- Repository name is valid kebab-case
- Team starts with `team-`
- At least one environment is selected
- Unless ingress is disabled (`ingress.enabled: false`), an ingress host is given for every environment. ALWAYS ask the user for each `ingress.host` -- NEVER invent one or derive it from the repo name. You may show the pattern from `guides/playbooks/add-custom-domain.md` as a suggestion, but wait for the user to confirm the exact value. When ingress is disabled, do not ask for a host and do not write `ingress.host`

## Step 2: Print the Identity Chain

Before generating files, print the identity chain for the user to confirm:

```text
Identity Chain:
  metadata.id:        {appId}
  metadata.name:      {repoName}       (K8s namespace)
  GCP projects:       ent-{appId}-dev, ent-{appId}-tst, ent-{appId}-prd
  Helm app:           {repoName}
  Helm shortname:     {appId}
  Terraform app_id:   {appId}
  TF state bucket:    ent-gcs-tfa-{appId}
  Docker image:       {repoName}
  Secret Manager:     ent-{appId}-{env}
  Ingress hosts:      dev={devHost}, tst={tstHost}, prd={prdHost}   (or "disabled")
```

Ask the user to confirm before generating files, with an interactive question whose options are to confirm or to change a value.

After the user confirms, ALWAYS create a new branch before writing any file -- this skill writes to `.entur/`, `helm/`, `terraform/`, and `.github/`. Never generate these files on `main`.

The skill runs in two phases, each on its own branch:

1. **Phase 1 -- self-service manifests (Step 3).** Branch `feat/self-service-manifests` from `origin/main`. This branch contains **only** `.entur/` files.
2. **Phase 2 -- everything else (Steps 4--8).** Starts only after the Phase 1 PR is applied and merged, and `.entur/` exists on `main`. Branch `feat/bootstrap-{repoName}` from the updated `origin/main`.

Create each branch with `git fetch origin && git switch -c <branch> origin/main`.

**Why:** the GitHub Actions pipelines in `.github/` authenticate through the Workload Identity and GitHub environments that the `.entur/` manifests provision. Workflows pushed before the manifests are applied and merged fail.

## Step 3: Generate Self-Service Manifests

If `.entur/` already contains manifests, edit them in place -- never rename an existing manifest to match the filenames below.

If an existing manifest has a different `metadata.id` or `metadata.name` than the user's input, do NOT change it. Stop and show this message word for word:

> ⚠️ **Warning:** Do not perform these actions without consulting the Platform team first. They can have serious consequences, including the deletion of critical application resources such as service accounts and databases.
>
> Please contact the Platform team in **#talk-utviklingsplattform** before proceeding.

Continue with the existing values unless the user confirms they have talked to the Platform team.

ALWAYS open a pull request for the `.entur/` manifests: commit on the Phase 1 branch, push, and run `gh pr create`. Never commit them directly to `main`. Ask the user whether you may apply. Only when they explicitly allow it, show them the orchestrator's plan on the PR and run `gh pr comment <pr> --body "entur apply"`. Otherwise link the PR and tell the user to review the plan and comment `entur apply`.

Never write fields set to their platform default (for example `terraform.createBackend: true`, `secretManager.enabled: true`, `kubernetes.enabled: true`, `auth0.internal.type: m2m`). See the **Default** column in `guides/platform/self-service.md`.

### `.entur/github-{appId}.yaml`

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: {repoName}
```

Add `spec.environments: [{environments}]` only when the environments are not `[dev, tst, prd]`.

### `.entur/app-{appId}.yaml`

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudApplication
metadata:
  id: {appId}
  displayName: "{displayName}"
  name: {repoName}
  owner: {team}
  trigger: {currentUnixTimestamp}
spec:
  environments: [{environments}]
  repositories: [{repoName}]
```

The Terraform state bucket and Secret Manager are enabled by default -- do not add `terraform` or `secretManager` blocks.

If `auth0 M2M` is needed, add under `spec`:

```yaml
  auth0:
    internal:
      enabled: true
```

### Wait for the manifests to reach `main`

When `.entur/` does not exist on `main` (a new repository), STOP after opening the Phase 1 PR. Do not create `.github/` -- or any Phase 2 file -- until the PR is applied and merged. Tell the user to come back once it is merged, then confirm that `git fetch origin && git ls-tree -d --name-only origin/main .entur` prints `.entur` before starting Phase 2.

## Step 4: Generate Helm Chart

Steps 4--8 are Phase 2: run them on a new branch from the updated `origin/main` (see Step 2).

Read `guides/platform/common-helm.md` in the entur/ai repository for the full common chart reference.

### `helm/{repoName}/Chart.yaml`

```yaml
apiVersion: v2
name: {repoName}
version: 0.1.0
dependencies:
  - name: common
    version: "1.22.0"
    repository: "https://entur.github.io/helm-charts"
    alias: common
```

### `helm/{repoName}/values.yaml`

```yaml
common:
  app: {repoName}
  shortname: {appId}
  team: {teamWithoutPrefix}
  container:
    image: {repoName}
    cpu: {cpuByLanguage}
    memory: {memoryByLanguage}
    memoryLimit: {memoryByLanguage}
    probes:                          # Go/Python only -- omit for Kotlin/Java
      liveness:
        path: {livenessByLanguage}
      readiness:
        path: {readinessByLanguage}
    prometheus:
      enabled: true
      path: {metricsByLanguage}      # Go/Python only -- omit for Kotlin/Java
  ingress:
    trafficType: api   # when ingress is disabled, replace with `enabled: false`
  configmap:
    enabled: true
```

Never write values equal to the `common` chart default (for example `ingress.enabled: true`, `service.internalPort: 8080`, the Spring Boot probe and metrics paths, `pdb.minAvailable: 50%`).

**Language defaults:**

| Setting | Kotlin/Java | Go | Python |
|---------|------------|-----|--------|
| cpu | 0.5 | 0.1 | 0.2 |
| memory | 512 | 64 | 128 |
| liveness path | /actuator/health/liveness | /health/liveness | /health/liveness |
| readiness path | /actuator/health/readiness | /health/readiness | /health/readiness |
| metrics path | /actuator/prometheus | /metrics | /metrics |

If database is needed, add:

```yaml
  postgres:
    enabled: true
    connectionConfig: {repoName}
```

If secrets are needed (do NOT add `PGUSER`/`PGPASSWORD` here — those are already provided by `postgres.enabled`), add:

```yaml
  secrets:
    my-app-secrets: [MY_APP_API_KEY, EXTERNAL_SERVICE_KEY]
```

### `helm/{repoName}/env/values-kub-ent-dev.yaml`

```yaml
common:
  env: dev
  ingress:
    host: {devHost}   # value given by the user; omit the block when ingress is disabled
  configmap:
    data:
      SPRING_PROFILES_ACTIVE: "dev"
      SPRING_CLOUD_GCP_SECRETMANAGER_PROJECTID: "ent-{appId}-dev"
```

Generate one `env/values-kub-ent-{env}.yaml` file per environment selected in Step 1 — not unconditionally all three. `values-kub-ent-tst.yaml` and `values-kub-ent-prd.yaml` follow the same pattern, adjusting the environment, project ID, and the user-provided `ingress.host` (when ingress is enabled). This exact filename is required: the CI/CD skill's `helm-lint` job and `cd.yml`'s deploy jobs request `values-kub-ent-${{ matrix.environment }}.yaml` — a differently named file will not be found.

For prd, also set:

```yaml
  hpa:
    spec:
      maxReplicas: 5
```

## Step 5: Generate Terraform Scaffolding

Read `guides/platform/terraform-modules.md` in the entur/ai repository for the full module reference.

### `terraform/main.tf`

```hcl
module "init" {
  source      = "github.com/entur/terraform-google-init//modules/init?ref=v1"
  app_id      = var.app_id
  environment = var.environment
}
```

If database is needed, add:

```hcl
module "postgresql" {
  source    = "github.com/entur/terraform-google-sql-db//modules/postgresql?ref=v1"
  init      = module.init
  databases = ["{repoName}"]
}
```

If Redis is needed, add:

```hcl
module "redis" {
  source = "github.com/entur/terraform-google-memorystore//modules/redis?ref=v2"
  init   = module.init
}
```

If Kafka is needed, this skill does not generate Kafka scaffolding directly -- Aiven cluster access, topic naming, and Avro/Protobuf schema registration are not a simple Terraform module addition. Read `guides/playbooks/add-kafka.md` and `guides/platform/entur-kafka-starter.md` in the entur/ai repository and report Kafka setup as a required follow-up step in the Step 9 summary (do not silently drop the request).

### `terraform/variables.tf`

```hcl
variable "app_id" {
  description = "Application ID (must match self-service metadata.id)"
  type        = string
}

variable "environment" {
  description = "Environment: dev, tst, or prd"
  type        = string
}
```

### `terraform/providers.tf`

```hcl
terraform {
  backend "gcs" {
    bucket = "ent-gcs-tfa-{appId}"
  }

  required_providers {
    google = {
      source = "hashicorp/google"
    }
  }
}
```

### `terraform/env/{env}.tfvars`

Generate one `.tfvars` file per environment selected in Step 1 (e.g. `dev.tfvars`, `tst.tfvars`, `prd.tfvars` for the full set) — not unconditionally all three:

```hcl
app_id      = "{appId}"
environment = "{env}"
```

## Step 6: Generate Dockerfile

Read `guides/reference/docker.md` in the entur/ai repository for the full Docker reference.

Select the Dockerfile template based on language:

### Kotlin/Java

Use multi-stage with a layered JAR and distroless runtime:

- Build stage: `gradle:9.6.1-jdk25-alpine`
- Layers stage: `eclipse-temurin:25-jre-alpine`
- Runtime stage: `gcr.io/distroless/java25-debian13:nonroot`
- Non-root runtime, port 8080, `-XX:MaxRAMPercentage=75.0`

### Go

- Build stage: `golang:1.26-alpine` with `CGO_ENABLED=0`
- Runtime stage: `gcr.io/distroless/static-debian13:nonroot`
- Port 8080

### Python

- Build stage: `python:3.13-slim` with pip install
- Runtime stage: `python:3.13-slim`
- Non-root user, port 8080

## Step 7: Generate CI/CD Workflows

Delegate to the **setup-cicd-workflows** skill (`skills/setup-cicd-workflows/SKILL.md` in the entur/ai repository) -- it is the canonical source for workflow file names, structure, and reusable-workflow `@vN` pins. Do not generate these files from this skill: keeping two sources of truth has already caused drift.

Do not restate the workflow file list in this skill. Workflow files use `.yml`; the setup-cicd skill owns the exact set.

For background on each workflow's role, read `guides/platform/gha-workflows.md` in the entur/ai repository.

## Step 8: Generate Supporting Files

### `.mise.toml`

```toml
[tools]
java        = 'temurin-25'            # Kotlin/Java
# go        = '1.26'                   # Go
terraform   = '1.15.8'

[settings]
experimental = true

[hooks]
enter = 'mise install'
```

### `AGENTS.md`

```markdown
# {displayName}

{language} application that {brief description}.

## Entur Standards

Read and follow the Entur platform standards at:
https://github.com/entur/ai/blob/main/AGENTS.md

## Project-Specific

- App ID: {appId}
- GCP Projects: ent-{appId}-dev, ent-{appId}-tst, ent-{appId}-prd
```

## Step 9: Print Summary

After generating all files, print:

1. Files created (list with paths)
2. The identity chain (repeated for reference)
3. If Kafka was requested: an explicit "Kafka setup not yet complete" line pointing to `guides/playbooks/add-kafka.md` and `guides/platform/entur-kafka-starter.md` -- do not let the request disappear silently.
4. Next steps:
   - Run `helm dependency update helm/{repoName}/`
   - After Phase 1: link the pull request opened for the `.entur/` manifests. The user reviews the plan and comments `entur apply` (or allows you to comment it), then merges. Phase 2 starts only after the merge
   - After GCP projects are created: set up Terraform workspaces and apply
   - Configure Helm deploy via CI/CD
   - Turn on **Deployment protection rules** for each GitHub environment. Ask whether to add a team (`entur/<team-name>`, e.g. `entur/team-plattform`) as required reviewers, following Step 13 of the setup-cicd-workflows skill

## Critical Rules

- **Never** create GCP projects via Terraform or gcloud -- only via self-service manifests
- **Always** use Entur Terraform modules, not raw `google_*` resources
- **Always** use the Entur `common` Helm chart
- **Always** use Entur reusable GitHub Actions workflows
- **Pin all dependencies** -- Terraform `?ref=TAG`, Actions `@vN`, Docker image tags
- **metadata.id is immutable** -- changing it deletes and recreates GCP projects
- **Always open a pull request** for changes in `.entur/` -- never commit them directly to `main`
- **In a new repository, `.entur/` is merged before `.github/` is created** -- the pipelines fail until the manifests are provisioned
