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
- One YAML document per file (no multi-doc `---`)
- Default naming: `.entur/<metadata.id>.yaml`

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

### GitHubActions Examples

Minimal (defaults to all three environments):

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: my-repo
```

Dev only:

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: my-repo
spec:
  environments:
    - dev
```

Full with trigger:

```yaml
apiVersion: orchestrator.entur.io/github/v1
kind: GitHubActions
metadata:
  id: my-repo
  trigger: 1654089480
spec:
  environments:
    - dev
    - tst
    - prd
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

### Application Full Example

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
      denyEgress: true
      ingress:
        allowedNamespaces:
          - up-k8s
          - helloworld
  network:
    sharedVpcEnabled: true
  terraform:
    createBackend: true
  auth0:
    internal:
      enabled: false
      type: m2m
  appLogBucket:
    enabled: true
    retentionDays: 30
    disableSink: false
    logAnalyticsEnabled: true
  defaultLogBucket:
    logAnalyticsEnabled: true
  appEngine:
    enabled: true
    databaseType: firestore
  secretManager:
    enabled: true
    serviceAccount: application
  serviceAccounts:
    - id: application
      additionalRoles:
        - roles/storage.objectCreator
    - id: "my-custom-account"
      kubernetesEnabled: true
      displayName: "MyCustomAccount"
      description: "A custom account for my app"
      roles:
        - roles/bigquery.admin
  organization: entur
  repositories:
    - my-github-repository
  environments:
    - dev
  quotas:
    bigQuery:
      dailyQuotaPerUser: 1.5
      dailyQuota: 20
```

---

## Testing with Mock Manifests

A mock manifest kind (`orchestrator.entur.io/mock/v1`, kind `MockItem`) exists for testing the workflow without affecting real resources. The flow is identical: create `.entur/*.yaml`, open PR, review plan, `entur apply`, merge.
