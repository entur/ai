# Security Standards

Security conventions for all Entur services. Entur uses OWASP ASVS three-tier model:

- **Level 1**: Mandatory for all services
- **Level 2**: Required for services handling money or personal information
- **Level 3**: Nice-to-have improvements

## Secret Management

### Rules

- **Never hardcode secrets** in source code, config files, Dockerfiles, or CI workflows
- **Never commit secrets** to Git -- not even in "test" configurations
- **Use Google Secret Manager** for all secrets (passwords, API keys, tokens)
- **Reference secrets via ExternalSecrets** in Kubernetes (configured in Helm values)

### Creating Secrets (Terraform)

Entur Terraform modules create secrets automatically (e.g. `terraform-google-sql-db` creates `PG_USER`/`PG_PASSWORD`). For custom secrets:

```hcl
resource "google_secret_manager_secret" "api_key" {
  secret_id = "${module.init.app.id}-API_KEY"
  project   = module.init.app.project_id

  replication {
    auto {}
  }

  labels = module.init.labels
}
```

### Consuming Secrets (Helm)

```yaml
# helm/<app>/env/dev.yaml
common:
  secrets:
    my-secret:               # ExternalSecret name -> creates K8s Secret
      - API_KEY              # Secret Manager secret name -> mounted as env var
      - EXTERNAL_SERVICE_KEY
```

### Consuming Secrets (Application Code)

```java
// Java/Kotlin - via Spring property
@Value("${API_KEY}")
private String apiKey;
```

```go
// Go
apiKey := os.Getenv("API_KEY")
```

### Local Development

Access secrets locally via `gcloud`:

```bash
APP_PASSWORD=$(gcloud secrets versions access latest --secret <secret-name> --project <project-id>)
export APP_PASSWORD
```

Do not store secrets locally in files.

### Secret Rotation

Rotate secrets at least every 90 days (excludes OAuth tokens which should have shorter lifetimes). Secret sync to K8s happens hourly; redeployment typically required for updates.

## Authentication Patterns

### Service-to-Service (Internal)

- Use GCP service account identity with Workload Identity
- Terraform `init` module provides necessary service account outputs

### External API Authentication

- Use OAuth 2.0 / OpenID Connect for user-facing APIs
- Validate JWT tokens at API gateway or in application
- Never build custom authentication -- use established libraries
- Use Entur OIDC libraries: `oidc-auth-resource-server` (validation), `oidc-auth-client` (token acquisition)

### Authorization

Centralized system based on **Permission Store** and **Permission Client**. See [authorization.md](authorization.md) for details.

- **Business Capabilities**: operation + access level (LES, OPPRETT, ENDRE, SLETT) for endpoint-level access
- **Responsibility Sets**: operation + responsibility type + object key for data-level access
- **Agreements**: link responsibility sets to organisations
- Use `@PreAuthorize("hasPermission('operation', 'access')")` in Spring controllers
- Use `LOCAL_TEST_CACHE` with test users for testing
- Use `IN_MEMORY` cache with WebSocket push notifications in production

## HTTP Security Headers

All HTTP responses must include:

- **`Content-Type`** with safe charset (UTF-8 or ISO-8859-1) for text types
- **`X-Content-Type-Options: nosniff`**
- **`Strict-Transport-Security: max-age=15724800; includeSubdomains`**

For web applications (not REST APIs), also set:

- **`Content-Security-Policy`** to mitigate XSS
- **`X-Frame-Options`** or `frame-ancestors` to prevent clickjacking

## Input Validation

- Validate all input at API boundaries using allow lists when possible
- Use **DTOs** to prevent mass parameter assignment -- never bind input directly to domain objects
- Only accept HTTP methods actually used; log unexpected methods
- Do not use `Origin` header for auth or access control
- Validate URL redirects using an allow list
- Enforce strongly typed schemas: allowed characters, length, pattern

## Dependency Scanning

### Dependabot

Automatically enabled for all Entur repositories. Use Dependabot (not Renovate). Configure for: `docker`, `helm`, `github-actions`, `terraform`, `npm`, `gradle`, `gomod`, `pip`.

Vulnerabilities must be triaged and fixed within **30 days**. Manage in repository **Security** tab.

### Code Scanning (CodeQL)

Every repository must have `.github/workflows/codeql.yml`:

```yaml
name: CodeQL

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 6 * * 1'    # Weekly on Monday

jobs:
  code-scan:
    uses: entur/gha-security/.github/workflows/code-scan.yml@v2
    secrets: inherit
```

The workflow name **must** be `codeql.yml`.

### Docker Image Scanning

```yaml
jobs:
  docker-build:
    uses: entur/gha-docker/.github/workflows/build.yml@v1
  docker-scan:
    needs: [docker-build]
    uses: entur/gha-security/.github/workflows/docker-scan.yml@v2
    secrets: inherit
    with:
      image_artifact: ${{ needs.docker-build.outputs.image_artifact }}
```

### Allowlisting False Positives

#### Code Scan Allowlist (`.entur/security/codescan.yml`)

```yaml
apiVersion: entur.io/securitytools/v1
kind: CodeScanConfig
metadata:
  id: my-application
spec:
  allowlist:
    - cwe: "cwe-080"
      comment: "False positive: XSS not possible in this context"
      reason: "false_positive"        # false_positive | wont_fix | test
  notifications:
    severityThreshold: "high"         # low | medium | high | critical
    outputs:
      slack:
        enabled: true
        channelId: "C01ABCDEFGH"
      pullRequest:
        enabled: true
```

#### Docker Scan Allowlist (`.entur/security/dockerscan.yml`)

```yaml
apiVersion: entur.io/securitytools/v1
kind: DockerScanConfig
metadata:
  id: my-application
spec:
  allowlist:
    - cve: "CVE-2024-1234"
      comment: "Not exploitable in our configuration"
      reason: "false_positive"
  notifications:
    severityThreshold: "high"
    outputs:
      slack:
        enabled: true
        channelId: "C01ABCDEFGH"
      pullRequest:
        enabled: false
```

## Container Security

For Dockerfile conventions and examples, see [Docker guide](docker.md).

- Use minimal base images (distroless/slim), run as non-root, multi-stage builds
- Pin base image versions, no secrets in images, no unnecessary packages

Kubernetes security:

- Pods run as non-root by default (enforced by common Helm chart)
- Use network policies for pod-to-pod restriction where appropriate
- Use Workload Identity -- never mount service account keys
- Never mount default K8s service account token unless required

## IAM and Permissions

Only use roles from the [approved list](terraform/iam-roles.md). Request additions in `#talk-utviklerplattform`.

| Role | Use case |
|------|----------|
| `roles/secretmanager.secretAccessor` | Read secrets from Secret Manager |
| `roles/cloudsql.client` | Connect to Cloud SQL via proxy |
| `roles/pubsub.publisher` | Publish to Pub/Sub |
| `roles/pubsub.subscriber` | Subscribe to Pub/Sub |
| `roles/storage.objectViewer` | Read from Cloud Storage |
| `roles/bigquery.dataViewer` | Query BigQuery datasets |

## Vulnerability Management

Use the GitHub **Security** tab to review code vulnerabilities, dependency vulnerabilities, and committed secrets. Organization overview: `https://github.com/orgs/entur/security/overview`.

When dismissing alerts, provide a reason: **Fix already started** (link to PR/ticket) or **False positive** (explain why).

## Access Control

- Enforce server-side access control (zero trust: "never trust, always verify")
- Access control attributes must not be manipulable by end users
- Principle of least privilege for all users and services
- All accounts must be single-purpose
- Access controls must fail securely, including on exceptions
- Use attribute/feature-based authorization, not role-based
- Each team controls access to its own data
- Grant IAM permissions at **group level** over individual users
- Manage data access through Terraform in CD pipeline for auditability

## Security Checklist

Before submitting a PR:

- [ ] No secrets in source code or config files
- [ ] Secrets in Google Secret Manager, referenced via ExternalSecrets
- [ ] Dockerfile runs as non-root
- [ ] Base images pinned to specific versions
- [ ] Dependencies scanned (CodeQL + Docker scan in CI)
- [ ] IAM roles from approved list only
- [ ] Input validation at API boundaries
- [ ] Error responses don't leak internals (stack traces, DB errors)
- [ ] Auth and authorization properly configured
- [ ] CORS configured restrictively (not `*` in production)
