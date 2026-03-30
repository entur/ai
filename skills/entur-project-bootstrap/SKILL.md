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

## Step 1: Gather Inputs

Ask the user for the following. If not provided, suggest sensible defaults:

| Input | Required | Constraints | Default |
|-------|----------|-------------|---------|
| **Repository name** | yes | lowercase kebab-case, max 63 chars, `[a-z0-9-]` | -- |
| **App ID** (`metadata.id`) | yes | 3-10 lowercase alphanumeric `^[a-z0-9]+$`, no `ent-` prefix, no env suffixes | -- |
| **Team** | yes | must start with `team-` | -- |
| **Language** | yes | `kotlin`, `java`, `go`, `python` | `kotlin` |
| **Environments** | no | subset of `dev`, `tst`, `prd` | `[dev, tst, prd]` |
| **Needs database** | no | boolean | `false` |
| **Needs Redis** | no | boolean | `false` |
| **Needs Kafka** | no | boolean | `false` |
| **Needs Auth0 M2M** | no | boolean | `false` |
| **Display name** | no | human-readable name | derived from repo name |

### Validation Rules

Before proceeding, validate:

- `metadata.id` is 3-10 chars, `^[a-z0-9]+$`
- `metadata.id` does NOT start with `ent-`
- `metadata.id` does NOT end with `sbx`, `dev`, `tst`, `test`, `prd`, `prod`
- Repository name is valid kebab-case
- Team starts with `team-`

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
```

Ask the user to confirm before generating files.

## Step 3: Generate Self-Service Manifests

### `.entur/cicd.yaml`

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: {repoName}
spec:
  environments: [{environments}]
```

### `.entur/{appId}.yaml`

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
  terraform:
    createBackend: true
  secretManager:
    enabled: true
    serviceAccount: application
```

If `auth0 M2M` is needed, add under `spec`:

```yaml
  auth0:
    internal:
      enabled: true
      type: m2m
```

## Step 4: Generate Helm Chart

Read `guides/helm.md` in the entur/ai repository for the full common chart reference.

### `helm/{repoName}/Chart.yaml`

```yaml
apiVersion: v2
name: {repoName}
version: 0.1.0
dependencies:
  - name: common
    version: "1.21.1"
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
    probes:
      liveness:
        path: {livenessByLanguage}
      readiness:
        path: {readinessByLanguage}
    prometheus:
      enabled: true
      path: {metricsByLanguage}
  service:
    internalPort: 8080
  ingress:
    enabled: true
    trafficType: api
  configmap:
    enabled: true
```

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

If secrets are needed, add:

```yaml
  secrets:
    psql-credentials: [PG_USER, PG_PASSWORD]
```

### `helm/{repoName}/env/dev.yaml`

```yaml
common:
  env: dev
  configmap:
    data:
      SPRING_PROFILES_ACTIVE: "dev"
      SPRING_CLOUD_GCP_SECRETMANAGER_PROJECTID: "ent-{appId}-dev"
```

Generate `tst.yaml` and `prd.yaml` with the same pattern, adjusting the environment and project ID.

For prd, also set:

```yaml
  hpa:
    spec:
      maxReplicas: 5
  pdb:
    minAvailable: 50%
```

## Step 5: Generate Terraform Scaffolding

Read `guides/terraform/modules.md` in the entur/ai repository for the full module reference.

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

### `terraform/env/dev.tfvars`, `tst.tfvars`, `prd.tfvars`

```hcl
app_id      = "{appId}"
environment = "{env}"
```

## Step 6: Generate Dockerfile

Read `guides/docker.md` in the entur/ai repository for the full Docker reference.

Select the Dockerfile template based on language:

### Kotlin/Java

Use multi-stage with layered JAR and CDS:

- Build stage: `gradle:9.3.1-jdk25-alpine`
- Layers stage: `bellsoft/liberica-runtime-container:jre-25-cds-slim-musl`
- Runtime stage: `bellsoft/liberica-runtime-container:jre-25-cds-slim-musl`
- Non-root user, port 8080, `-XX:MaxRAMPercentage=75.0`

### Go

- Build stage: `golang:1.25-alpine` with `CGO_ENABLED=0`
- Runtime stage: `gcr.io/distroless/static-debian12:nonroot`
- Port 8080

### Python

- Build stage: `python:3.12-slim` with pip install
- Runtime stage: `python:3.12-slim`
- Non-root user, port 8080

## Step 7: Generate CI/CD Workflows

Delegate to the **setup-cicd-workflows** skill if available, or generate:

- `.github/workflows/ci.yml` (reusable build)
- `.github/workflows/ci-pr.yml` (PR verification)
- `.github/workflows/deploy.yml` (dev → tst → prd)
- `.github/workflows/codeql.yml` (security scanning)
- `.github/dependabot.yml`

Read `guides/cicd/workflows.md` in the entur/ai repository for the full workflow reference.

## Step 8: Generate Supporting Files

### `.mise.toml`

```toml
[tools]
java        = 'liberica-25.0.2+12'    # Kotlin/Java
# go        = '1.25'                   # Go
terraform   = '1.9.8'

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
3. Next steps:
   - Run `helm dependency update helm/{repoName}/`
   - Commit self-service manifests, open PR, comment `entur apply`
   - After GCP projects are created: set up Terraform workspaces and apply
   - Configure Helm deploy via CI/CD

## Critical Rules

- **Never** create GCP projects via Terraform or gcloud -- only via self-service manifests
- **Always** use Entur Terraform modules, not raw `google_*` resources
- **Always** use the Entur `common` Helm chart
- **Always** use Entur reusable GitHub Actions workflows
- **Pin all dependencies** -- Terraform `?ref=TAG`, Actions `@vN`, Docker image tags
- **metadata.id is immutable** -- changing it deletes and recreates GCP projects
