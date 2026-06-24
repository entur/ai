# Self-Service Platform Provisioning

Define YAML manifests in `.entur/` and apply through a GitOps PR workflow.

## How It Works

1. Create/modify YAML manifests in `.entur/`.
2. Open a PR. The orchestrator validates and presents a **plan**.
3. Comment `entur apply` on the PR.
4. Wait for apply to succeed, then merge.

### Aborting or Rolling Back

- **Not yet applied**: Close the PR.
- **Already applied**: Revert the manifest, run `entur apply` again, then merge/close.

### GitHub Repository Requirements

If you use Repository Rulesets, add a bypass for the **Platform Orchestrator** GitHub application on each rule.

## Manifest Kinds

The **GitHub manifest** must be applied before any **Application manifest**.

| Kind | apiVersion | Purpose |
|------|-----------|---------|
| `GitHubActions` | `orchestrator.entur.io/github/v1` | GCP Workload Identity + GitHub environments for CI/CD |
| `GoogleCloudApplication` | `orchestrator.entur.io/apps/v1` | GCP projects for containerized K8s apps |
| `GoogleCloudFirebaseApplication` | `orchestrator.entur.io/apps/v1` | GCP projects for Firebase apps |
| `GoogleCloudDataProject` | `orchestrator.entur.io/apps/v1` | GCP projects for data workloads |

### File Conventions

- Manifests live in `.entur/` at repository root
- ALWAYS use one YAML document per file (single document, no `---` separator)
- Default naming: `.entur/<metadata.id>.yaml`

## GCP Project Naming

The Platform Orchestrator ALWAYS creates GCP projects automatically from your `metadata.id`.

| Kind | Project ID Pattern | Example (`metadata.id: myapp`) |
|------|-------------------|-------------------------------|
| `GoogleCloudApplication` | `ent-{metadata.id}-{env}` | `ent-myapp-dev`, `ent-myapp-tst`, `ent-myapp-prd` |
| `GoogleCloudFirebaseApplication` | `ent-{metadata.id}-{env}` | `ent-myapp-dev`, `ent-myapp-prd` |
| `GoogleCloudDataProject` | `ent-data-{metadata.id}-{int\|ext}-{env}` | `ent-data-myapp-int-dev`, `ent-data-myapp-ext-prd` |

Data projects use a different prefix (`ent-data-`) and include an `int`/`ext` suffix indicating whether the project is for internal or external data sharing. This is controlled by `spec.dataAccess.external` (`true` → `ext`, `false` → `int`).

**This project ID is used everywhere:**

- **Terraform** `app_id` variable → `module.init.app.project_id` resolves to `ent-{metadata.id}-{env}`
- **Terraform state bucket**: `ent-gcs-tfa-{metadata.id}`
- **Helm** `shortname` should match `metadata.id`
- **Secret Manager**: secrets are stored in the application's GCP project (`ent-{metadata.id}-{env}`)

**Constraints on `metadata.id`:**

- 3--10 characters, lowercase alphanumeric only (`^[a-z0-9]+$`)
- MUST NOT end with `sbx`, `dev`, `tst`, `prd`, or `prod` (these clash with the auto-appended environment suffix)
- MUST NOT start with `ent-` -- the platform adds this prefix automatically
- **Immutable** -- changing it deletes and recreates all GCP projects

**Example identity chains:**

```text
# Application (GoogleCloudApplication)
metadata.id: products        → GCP projects: ent-products-dev, ent-products-tst, ent-products-prd
metadata.name: products-api  → K8s namespace: products-api
                             → Helm shortname: products
                             → Terraform app_id: products
                             → Terraform state: ent-gcs-tfa-products

# Data project (GoogleCloudDataProject, dataAccess.external: true)
metadata.id: akt             → GCP projects: ent-data-akt-ext-dev, ent-data-akt-ext-prd
```

## Environments

The platform recognizes four GCP environments: `dev`, `tst`, and `prd`.

- An environment can only be used if it has been provisioned for the owning team. Check the team's folder structure in the GCP Console; reach out to Team Plattform if a folder is missing.
- `spec.environments` on Application manifests defaults to `[dev, tst, prd]` -- but it is **required** to be set explicitly.
- `spec.environments` on `GitHubActions` defaults to `[dev, tst, prd]` and can be omitted entirely.
- The environment list MUST match between the `GitHubActions` manifest and its linked Application manifest.

## Getting Started

Common setup: containerized application on Kubernetes in Google Cloud.

### Prerequisites

- Read the [DevOps Handbook](https://enturnett.atlassian.net/wiki/spaces/ESP/overview) Plan section
- Onboard to GitHub and create a repository
- Build a web application listening on port `8080` with:
  - `GET /actuator/health/liveness` → HTTP 200
  - `GET /actuator/health/readiness` → HTTP 200

### Step 1: Create the GitHub Manifest

Create `.entur/cicd.yaml`:

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: myuniquerepo  # must match your repository name exactly
spec:
  environments: [dev, tst, prd]
```

### Step 2: Create the Application Manifest

Create `.entur/application.yaml`:

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudApplication
metadata:
  id: myappid  # 3-10 lowercase alphanumeric chars, unique in Entur org ("App ID")
  displayName: My Application
  name: my-unique-app  # 3-30 chars, lowercase alphanumeric + hyphens, becomes your K8s namespace
  owner: team-excellence
  trigger: 1747398600  # current unix timestamp, see https://unixtime.org/
spec:
  environments: [dev, tst, prd]
  repositories: [myuniquerepo]  # repos that can deploy to this application
```

> **Important:** Remember `metadata.id` -- you need it for Helm configuration.

### Step 3: Apply via PR

1. Commit both files to a new branch.
2. Push and open a PR targeting main.
3. Review the plan output, then comment `entur apply`.
4. Wait for successful apply, then merge.

### Next Steps

1. **Document your API** -- See [API design](api-design.md)
2. **Create a container image** -- See [Docker guide](docker.md)
3. **Set up CI/CD pipelines** -- See [CI/CD workflows](cicd/workflows.md)
4. **Configure Helm deployment** -- See [Helm guide](helm.md)

---

## GitHubActions Manifest Reference

Configures GCP Workload Identity and GitHub environments for CI/CD.

### GitHubActions Fields

| Field | Required | Type | Constraints |
|-------|----------|------|-------------|
| `apiVersion` | yes | string | Must be `orchestrator.entur.io/github/v1` |
| `kind` | yes | string | Must be `GitHubActions` |
| `metadata.id` | yes (immutable) | string | 1--63 chars, `^[A-Za-z0-9_.-]+$`. **Must match the GitHub repository name exactly.** If the repo is renamed, delete and recreate this manifest. |
| `metadata.trigger` | no | integer | Unix timestamp (1--9999999999). Change to force re-apply without other manifest changes. |
| `spec` | no | object | The entire `spec` block is optional; omit it to accept defaults. |
| `spec.environments` | no | array | Unique values from `dev`, `tst`, `prd`. Default: `[dev, tst, prd]`. Must match the linked Application manifest. |

### GitHubActions Example

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: my-repo              # Must match GitHub repository name exactly
  trigger: 1654089480      # Optional: unix timestamp to force re-apply
spec:
  environments: [dev, tst, prd]  # Optional: defaults to [dev, tst, prd]
```

---

## Application Manifest Reference

Provisions GCP projects and related resources. Three kinds: `GoogleCloudApplication`, `GoogleCloudFirebaseApplication`, `GoogleCloudDataProject`.

### Application Fields

- **`apiVersion`** (required): `orchestrator.entur.io/apps/v1`
- **`kind`** (required): One of the three kinds above

#### `metadata` (required)

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `id` | string | yes | 3--10 chars, `^[a-z0-9]+$`. Must NOT end with `sbx\|dev\|tst\|prd\|prod`. Must NOT start with `ent-`. **Immutable -- changing deletes and recreates GCP projects.** |
| `displayName` | string | yes | Human-friendly name (short description of the application). |
| `name` | string | yes | 3--30 chars, `^[a-z0-9-]+$`. Becomes K8s namespace. **Changing is disruptive** (namespace rename, pod restarts). |
| `owner` | string | yes | Must start with `team-` and match an existing team in the GCP `teams/` folder structure. |
| `trigger` | integer | no | Unix timestamp to force re-apply. |
| `domain` | string | no | **Deprecated** -- avoid. |

#### `spec.environments` (required)

Array with unique values from `dev`, `tst`, `prd`. The owning team's GCP folder must already include the chosen environments.

#### `spec.repositories` (recommended)

GitHub repository names (`^[A-Za-z0-9_.-]+$`) that can deploy to this application. Required for CI/CD permissions.

### Optional spec Blocks

Defaults are shown where the platform applies one when the field is omitted.

#### `spec.kubernetes`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `enabled` | bool | no | `true` | Create K8s namespace + workload-identity service accounts. |
| `clusterGroup` | string | no | `entur` | One of `entur`, `journeyPlanner`. |
| `securityPolicy.level` | string | no | -- | Optional security policy level. |
| `networkPolicies.enabled` | bool | no | `true` | Enforce platform-managed network policies. |
| `networkPolicies.denyInternal` | bool | no | `false` | Deny ingress from other namespaces. |
| `networkPolicies.denyPublic` | bool | no | `false` | Deny ingress from outside the cluster. |
| `networkPolicies.denyEgress` | bool | no | `false` | Deny all egress traffic. |
| `networkPolicies.ingress.allowedNamespaces` | string[] | no | `[]` | Each value `^[a-z0-9-]+$`. |

#### `spec.terraform`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `createBackend` | bool | no | `true` | Provision the Terraform state bucket `ent-gcs-tfa-{metadata.id}`. |

#### `spec.auth0`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `internal.enabled` | bool | **yes** (when block present) | `false` | Provision Auth0 client credentials in the internal tenant. |
| `internal.type` | string | no | `m2m` | Only `m2m` is supported. |

#### `spec.appLogBucket`

For long-term log retention beyond the shared Stackdriver default (30 days). Pods must carry the label `customLogRetention=enabled`.

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `enabled` | bool | no | `false` | Create a GCS bucket + log sink in the application project. |
| `retentionDays` | int | no | `30` | Days to retain forwarded logs. |
| `disableSink` | bool | no | `false` | Disable the sink without removing the bucket. |
| `logAnalyticsEnabled` | bool | no | `false` | Enable Log Analytics for SQL queries. |

#### `spec.defaultLogBucket`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `logAnalyticsEnabled` | bool | no | `false` | Enable Log Analytics on the `_Default` log bucket. |
| `location` | string | no | -- | Override the bucket location. |

#### `spec.appEngine`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `enabled` | bool | **yes** (when block present) | `false` | Bootstrap App Engine in the project. |
| `databaseType` | string | no | `firestore` | `firestore` or `datastore`. **Cannot be changed after creation.** |

#### `spec.secretManager`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `enabled` | bool | **yes** (when block present) | `true` | Create a `secretStore` resource in the K8s namespace. |
| `serviceAccount` | string | no | `application` | Must match one of `spec.serviceAccounts[].id`. Do NOT add `secretmanager.secretAccessor` to that account's roles -- the platform manages it. |

#### `spec.serviceAccounts[]`

Array of additional service accounts. The platform always provisions a default `application@` SA with `roles/cloudsql.client`, `roles/storage.objectAdmin`, `roles/cloudprofiler.agent`. Setting `id: application` modifies the default SA.

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `id` | string | **yes** | -- | Lowercase letters, numbers, hyphens. Keep it short. |
| `displayName` | string | no | `${id}` | -- |
| `description` | string | no | autogenerated | -- |
| `kubernetesEnabled` | bool | no | `true` | Bind to a K8s service account via workload identity. |
| `additionalRoles` | string[] | no | `[]` | Adds to the default `application@` role set. Format `roles/<role>`. |
| `roles` | string[] | no | -- | Authoritative role list. Use with care on `application@` -- it replaces the defaults. |

**Blacklisted IAM roles** (pattern `^(roles/owner\|roles/editor\|roles/resourcemanager.*\|roles/iam.*)$`): contact Team Plattform if you need elevated permissions.

#### `spec.quotas`

Default BigQuery quotas are always set; use this block to override.

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `enabled` | bool | no | `true` | Set `false` to disable quota provisioning. |
| `bigQuery.dailyQuotaPerUser` | number | **yes** (when `bigQuery` present) | `10` | TiB per user per day. Min `0.000001` (1 MiB), max `200`. |
| `bigQuery.dailyQuota` | int | **yes** (when `bigQuery` present) | `20` | TiB per project per day. Min `0.000001`, max `200`. |

#### `spec.network`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `sharedVpcEnabled` | bool | **yes** (when block present) | `true` | Prepare the GCP project for Shared VPC. Omit the block to accept the default. |

#### `spec.firebase` (`GoogleCloudFirebaseApplication` only)

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `region` | string | **yes** (when block present) | -- | One of `europe-west`, `europe-west1`, `europe-west2`, `europe-west3`, `europe-west4`, `global`. |

> The Firebase schema accepts only `region`. Any other Firebase database settings (e.g. database name, type, delete protection) are managed by the platform.

#### `spec.dataAccess` (`GoogleCloudDataProject` only)

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `external` | bool | **yes** (when block present) | `false` | `true` → `ent-data-{id}-ext-{env}`, `false` → `ent-data-{id}-int-{env}`. **Immutable** -- baked into the GCP project name. |

#### Deprecated

- `metadata.domain` -- avoid.
- `spec.organization` -- only value `entur`; do not set.

Both `metadata` and `spec` enforce `additionalProperties: false` -- unknown fields are rejected.

### YAML Conventions

- 2-space indentation, no tabs
- Use explicit booleans (`true`/`false`)
- camelCase for all manifest fields (`displayName`, not `display_name`)

### Application Minimal Templates

GoogleCloudApplication:

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudApplication
metadata:
  id: myappid
  displayName: "My Application"
  name: my-application
  owner: team-myteam
spec:
  environments:
    - dev
```

GoogleCloudFirebaseApplication:

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudFirebaseApplication
metadata:
  id: mywebapp
  displayName: "My Firebase Web App"
  name: my-firebase-app
  owner: team-myteam
spec:
  environments:
    - dev
```

GoogleCloudDataProject:

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudDataProject
metadata:
  id: mydataprj
  displayName: "My Data Project"
  name: my-data-project
  owner: team-myteam
spec:
  dataAccess:
    external: true
  environments:
    - dev
```

### Application Example with Common Options

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudApplication
metadata:
  id: exampleapp
  displayName: "This is an example app"
  name: my-example-app
  owner: team-myteam
  trigger: 1654089480
spec:
  environments:
    - dev
    - tst
    - prd
  repositories:
    - my-github-repository
  kubernetes:
    enabled: true
    networkPolicies:
      enabled: true
      denyInternal: true
      denyPublic: true
  terraform:
    createBackend: true
  auth0:
    internal:
      enabled: true
      type: m2m
  secretManager:
    enabled: true
    serviceAccount: application
  serviceAccounts:
    - id: application
      additionalRoles:
        - roles/storage.objectCreator
```

---

## Testing with Mock Manifests

A mock manifest kind (`orchestrator.entur.io/mock/v1`, kind `MockItem`) exists for testing the workflow without affecting real resources. The flow is identical: create `.entur/*.yaml`, open PR, review plan, `entur apply`, merge.
