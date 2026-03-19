# Entur Common Helm Chart

Reference: [entur/helm-charts](https://github.com/entur/helm-charts)

The Entur `common` Helm chart is the standard base chart for deploying applications to Kubernetes. It provides sensible defaults for Spring Boot and can be configured for Go, Python, or any containerized service.

## Naming Convention

Application name = Git repository name = backend URL (`yourapp.entur.io`). Must be unique across Entur, max 63 characters, only `[a-z0-9-]`.

## Setup

### Chart.yaml

```yaml
apiVersion: v2
name: my-application
version: 0.1.0
dependencies:
  - name: common
    version: "1.21.1"
    repository: "https://entur.github.io/helm-charts"
    alias: common
```

Run `helm dependency update` after creating or modifying `Chart.yaml`.

### Directory Structure

```text
helm/
  my-application/
    Chart.yaml
    Chart.lock
    values.yaml           # Default values
    env/
      dev.yaml            # Dev overrides
      tst.yaml            # Test overrides
      prd.yaml            # Production overrides
    tests/                # Helm unit tests (optional)
      deployment_test.yaml
```

## Required Values

Every deployment must set these:

```yaml
# values.yaml
common:
  app: my-application
  shortname: myapp          # Max 10 characters, used for GCP 2.0 app ID
  team: my-team             # Team name without "team-" prefix
  container:
    image: my-application   # Docker image name (without registry/tag)
```

```yaml
# env/dev.yaml
common:
  env: dev
```

```yaml
# env/prd.yaml
common:
  env: prd
```

## Container Configuration

### Resources

```yaml
common:
  container:
    cpu: 0.2          # CPU request in cores (200m)
    memory: 256        # Memory request in Mi
    memoryLimit: 256   # Set equal to memory request
```

- **CPU limit**: Do not set. CPU is compressible -- pods throttled, not killed. Allow bursting.
- **Memory limit**: Set equal to request. Memory is incompressible -- exceeding limit causes OOM kills.
- Start small, let VPA recommend optimal settings.

### Replicas and Scaling

```yaml
common:
  container:
    replicas: 2              # Desired replicas (set to 1 for Recreate strategy)

  deployment:
    maxReplicas: 10          # HPA maximum (default: 10)
    minAvailable: "50%"      # PDB minimum available (default: 50%)
```

- In `prd`, HPA and PDB are enabled automatically when replicas > 1
- `replicas: 1` uses Recreate strategy (no PDB, no HPA)
- HPA scales on CPU utilization (80% threshold by default)
- **PDB percentage gotcha**: Kubernetes rounds `minAvailable` up. With 3 replicas and 80%, ceil(2.4) = 3, preventing all disruption. Use `50%` or ensure enough replicas.
- In dev/tst with a single pod, set `minAvailable: 0`

### Health Probes

Default paths (Spring Boot):

```yaml
common:
  container:
    probes:
      enabled: true
      liveness:
        path: /actuator/health/liveness
      readiness:
        path: /actuator/health/readiness
```

**Probe rules:**

- **Liveness**: Verify app responds within reasonable time. Do NOT check external deps -- failing liveness causes restarts.
- **Readiness**: Check only **private resources** (own DB, internal cache). Never check shared/external services -- all pods removed from routing simultaneously if shared service is down.

For non-Spring-Boot (Go, Python):

```yaml
common:
  container:
    probes:
      liveness:
        path: /health/liveness
      readiness:
        path: /health/readiness
```

Custom probe spec:

```yaml
common:
  container:
    probes:
      spec:
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Prometheus Metrics

```yaml
common:
  container:
    prometheus:
      enabled: true
      path: /actuator/prometheus    # Spring Boot
      # path: /metrics              # Go / Python
```

### Ports

```yaml
common:
  service:
    internalPort: 8080    # Container port (default)
    externalPort: 80      # Service port (default)
```

## Networking

### Ingress

```yaml
common:
  ingress:
    enabled: true                    # Default: true
    trafficType: api                 # Required: "api", "public", or "http2"
```

| Traffic Type | Description |
|-------------|-------------|
| `api` | Internal API traffic (default for backend services) |
| `public` | Public-facing traffic (internet-accessible) |
| `http2` | gRPC / HTTP/2 traffic |

### gRPC

```yaml
common:
  grpc: true
  ingress:
    trafficType: http2
```

When `grpc: true`, the chart automatically sets gRPC annotations and configures gRPC health checking probes.

## Database (Cloud SQL Proxy)

```yaml
common:
  postgres:
    enabled: true
```

Injects a Cloud SQL proxy sidecar. Application connects to `localhost:5432`. Credentials provided as env vars from K8s secrets (created by `terraform-google-sql-db` module): `PG_USER`, `PG_PASSWORD`. Database name configured via application config.

## Secrets (ExternalSecrets)

Sync secrets from Google Secret Manager to Kubernetes:

```yaml
common:
  secrets:
    my-app-secrets:           # K8s Secret name
      - API_KEY               # Secret Manager secret -> env var name
      - EXTERNAL_SERVICE_KEY
    database-credentials:     # Another K8s Secret
      - PG_USER
      - PG_PASSWORD
```

Each entry creates an ExternalSecret syncing named secrets from Secret Manager into a K8s Secret mounted as env vars.

## CronJobs

```yaml
common:
  cron:
    enabled: true
    schedule: "0 2 * * *"    # Daily at 02:00 UTC
```

## Environment Variables

```yaml
common:
  container:
    envFrom:
      - configMapRef:
          name: my-config
    env:
      - name: SPRING_PROFILES_ACTIVE
        value: "cloud"
      - name: JAVA_TOOL_OPTIONS
        value: "-Xmx512m"
```

## Custom HPA Spec

```yaml
common:
  hpa:
    spec:
      minReplicas: 2
      maxReplicas: 20
      metrics:
        - type: Resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 70
```

## Resource Sizing Best Practices

| Resource | Recommendation |
|----------|---------------|
| CPU request | Set for normal load with overhead. Compressible -- pods throttled, not killed. |
| CPU limit | **Do not set.** Allow bursting. |
| Memory request | Set close to real usage including burst. |
| Memory limit | **Set equal to request.** Incompressible -- exceeding causes OOM kills. |

Use VPA recommendations (enabled on all clusters) to tune over time. See [observability](observability.md) for Grafana dashboards.

## Local Debugging

```bash
# Lint (check for errors):
helm lint helm/my-application/ -f helm/my-application/env/dev.yaml

# Template (render K8s YAML locally):
helm template my-application helm/my-application/ -f helm/my-application/env/dev.yaml
```

## ConfigMap with Environment Variables

```yaml
# values.yaml
common:
  configmap:
    enabled: true

# env/dev.yaml
common:
  configmap:
    data:
      SPRING_PROFILES_ACTIVE: "dev"
      SPRING_CLOUD_GCP_SECRETMANAGER_PROJECTID: "ent-myapp-dev"
      ENTUR_PERMISSION_PERMISSIONCACHE_URL: "http://permission-store.dev.entur.internal"
```

## Multi-Namespace Deployments

For deploying to multiple namespaces (e.g., different data partitions), use separate values files per namespace:

```text
helm/my-app/
  values.yaml                     # Shared defaults
  env/
    values-kub-ent-dev.yaml       # Primary dev namespace
    values-kub-ent-dev-ep.yaml    # Secondary dev namespace
    values-kub-ent-tst.yaml
    values-kub-ent-tst-ep.yaml
    values-kub-ent-prd.yaml
```

Override `app` and `shortname` in secondary namespace values:

```yaml
# env/values-kub-ent-dev-ep.yaml
common:
  app: my-app-ep
  shortname: myappep
  ingress:
    host: my-app-ep.dev.entur.io
  postgres:
    connectionConfig: my-app-ep
```

Deploy using matrix strategy in GitHub Actions (see [CI/CD workflows](cicd/workflows.md)).

## Helm Values Naming Convention

Pattern: `values-kub-ent-{environment}.yaml` or `values-kub-ent-{environment}-{variant}.yaml`

Examples: `values-kub-ent-dev.yaml`, `values-kub-ent-tst-ep.yaml`, `values-kub-ent-prd.yaml`

## Complete Example (Spring Boot with Kotlin)

```yaml
# values.yaml
common:
  app: products-api
  shortname: products
  team: produkt
  ingress:
    trafficType: api
  service:
    internalPort: 8086
  configmap:
    enabled: true
  container:
    image: <+artifacts.primary.image>
    cpu: 0.5
    memory: 1000
    memoryLimit: 1000
    prometheus:
      enabled: true
  postgres:
    enabled: true
    connectionConfig: products-api
  secrets:
    psql-credentials: [PGINSTANCES, PGHOST, PGPORT, PGPASSWORD, PGUSER]
```

```yaml
# env/values-kub-ent-dev.yaml
common:
  ingress:
    host: products-api.dev.entur.io
  hpa:
    spec:
      maxReplicas: 2
  pdb:
    minAvailable: 40%
  env: dev
  configmap:
    data:
      SPRING_PROFILES_ACTIVE: "dev"
      SPRING_CLOUD_GCP_SECRETMANAGER_PROJECTID: "ent-products-dev"
```

```yaml
# env/values-kub-ent-prd.yaml
common:
  ingress:
    host: products-api.entur.io
  hpa:
    spec:
      maxReplicas: 5
  pdb:
    minAvailable: 50%
  env: prd
  configmap:
    data:
      SPRING_PROFILES_ACTIVE: "prd"
      LOG_LEVEL: "INFO"
      SPRING_CLOUD_GCP_SECRETMANAGER_PROJECTID: "ent-products-prd"
```

## Complete Example (Go Service)

```yaml
# values.yaml
common:
  app: stop-lookup
  shortname: stoplkup
  team: data-platform
  container:
    image: stop-lookup
    cpu: 0.1
    memory: 64
    replicas: 2
    probes:
      liveness:
        path: /health/liveness
      readiness:
        path: /health/readiness
    prometheus:
      enabled: true
      path: /metrics
  service:
    internalPort: 8080
  ingress:
    enabled: true
    trafficType: api
```

## Helm Unit Testing

Use [helm-unittest](https://github.com/helm-unittest/helm-unittest):

```yaml
# tests/deployment_test.yaml
suite: deployment tests
templates:
  - templates/deployment.yaml
tests:
  - it: should set the correct image
    set:
      common.container.image: my-app
    asserts:
      - contains:
          path: spec.template.spec.containers
          content:
            image: my-app
```

Run in CI:

```yaml
jobs:
  helm-unittest:
    uses: entur/gha-helm/.github/workflows/unittest.yml@v1
```
