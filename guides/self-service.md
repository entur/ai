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
- ALWAYS use bare identifiers -- the platform adds the `ent-` prefix automatically
- ALWAYS use base identifiers -- the platform adds the environment suffix automatically
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
| `apiVersion` | yes | | Must be `orchestrator.entur.io/github/v1` |
| `kind` | yes | | Must be `GitHubActions` |
| `metadata.id` | yes (immutable) | string | 1--63 chars, `^[A-Za-z0-9_.-]+$`. **Must match the GitHub repository name.** If repo is renamed, delete and recreate this manifest. |
| `metadata.trigger` | no | integer | Unix timestamp (1--9999999999). Change to force re-apply without other manifest changes. |
| `spec.environments` | no | array | Values from `dev`, `tst`, `prd` (unique). Default: all three. Must match linked Application manifest. |

### GitHubActions Example

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: my-repo              # Must match GitHub repository name exactly
  trigger: 1654089480       # Optional: unix timestamp to force re-apply
spec:
  environments: [dev, tst, prd]  # Optional: defaults to all three
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
| `id` | string | yes | 3--10 chars, `^[a-z0-9]+$`, must NOT end with `sbx\|dev\|tst\|test\|prd\|prod`, must NOT start with `ent-`. **Immutable -- changing deletes and recreates GCP projects.** |
| `displayName` | string | yes | Human-friendly name |
| `name` | string | yes | 3--30 chars, `^[a-z0-9-]+$`. Becomes K8s namespace. **Changing is disruptive** (namespace rename, pod restarts). |
| `owner` | string | yes | Must start with `team-` |
| `trigger` | integer | no | Unix timestamp to force re-apply |
| `domain` | string | no | Deprecated -- avoid |

#### `spec.environments` (required)

One or more of: `dev`, `tst`, `prd`.

#### `spec.repositories` (recommended)

GitHub repository names that can deploy to this application. Required for CI/CD permissions.

### Optional spec Blocks

- `kubernetes.enabled`, `kubernetes.clusterGroup` (`entur` or `journeyPlanner`), `kubernetes.securityPolicy.level`, `kubernetes.networkPolicies.enabled|denyInternal|denyPublic|denyEgress|ingress.allowedNamespaces`
- `terraform.createBackend`
- `auth0.internal.enabled`, `auth0.internal.type` (only `m2m`)
- `appLogBucket.enabled|retentionDays|disableSink|logAnalyticsEnabled`
- `defaultLogBucket.logAnalyticsEnabled|location`
- `appEngine.enabled|databaseType`
- `secretManager.enabled|serviceAccount`
- `serviceAccounts[]` with `id`, `additionalRoles`, `kubernetesEnabled`, `displayName`, `description`, `roles`
- `quotas.enabled`, `quotas.bigQuery.dailyQuotaPerUser`, `quotas.bigQuery.dailyQuota`
- `network.sharedVpcEnabled`

For `GoogleCloudFirebaseApplication` only:

- `firebase.region` with enum: `europe-west`, `europe-west1`, `europe-west2`, `europe-west3`, `europe-west4`, `global`

For `GoogleCloudDataProject` only:

- `dataAccess.external` (boolean)

Both `metadata` and `spec` have `additionalProperties: false` -- no unlisted fields allowed.

### YAML Conventions

- 2-space indentation, no tabs
- Use explicit booleans (`true`/`false`)

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
  organization: entur
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
  organization: entur
  repositories:
    - my-github-repository
  environments:
    - dev
    - tst
    - prd
```

---

## Testing with Mock Manifests

A mock manifest kind (`orchestrator.entur.io/mock/v1`, kind `MockItem`) exists for testing the workflow without affecting real resources. The flow is identical: create `.entur/*.yaml`, open PR, review plan, `entur apply`, merge.
