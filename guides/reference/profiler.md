# Cloud Profiler

How to enable Google Cloud Profiler on an Entur service so the runtime uploads CPU and heap profiles for production analysis.

- **Target audience**: developers adding profiling to a new service, or debugging CPU or memory pressure in an existing one.
- **Intent**: the workload uploads CPU and heap profiles to Cloud Profiler under its own service account; profiles are inspectable in the Cloud Console and via the read-side MCP tools.
- **Scope**: Terraform wiring (API + IAM), Java and Go agent setup, the `SERVICE_VERSION = K_REVISION` pattern for clean growth analysis, project routing. Health probes, Prometheus metrics, and tracing live in [observability.md](observability.md) and [tracing.md](tracing.md).
- **Prerequisites**: the application is provisioned via self-service YAML in `.entur/` (see [self-service.md](../platform/self-service.md)) and emits structured JSON logs (see [logging.md](logging.md)).

## Enable the Cloud Profiler API

Add `cloudprofiler.googleapis.com` to the per-env service activation block. Do **not** rely on the API being enabled implicitly by another resource -- `google_project_service` is the only source of truth.

```hcl
# terraform/main.tf
resource "google_project_service" "services" {
  for_each = toset([
    "cloudprofiler.googleapis.com",
    # ... other APIs your service needs
  ])

  project            = module.init.app.project_id
  service            = each.value
  disable_on_destroy = false
}
```

`module.init.app.project_id` resolves to the application's per-env project ID -- see [terraform-modules.md](../platform/terraform-modules.md#init-outputs) for the full module outputs.

## Grant the workload service account `roles/cloudprofiler.agent`

The Profiler agent uploads profiles under the workload's service account via Application Default Credentials. Bind exactly `roles/cloudprofiler.agent` -- it grants `cloudprofiler.profiles.{create,update}` (write only) and nothing else.

```hcl
# terraform/main.tf
resource "google_project_iam_member" "runtime_cloudprofiler_agent" {
  project = module.init.app.project_id
  role    = "roles/cloudprofiler.agent"
  member  = "serviceAccount:${google_service_account.runtime.email}"
}
```

Do **not** bind `roles/cloudprofiler.user` or `roles/cloudprofiler.admin` to the workload -- the agent only writes. Read access is granted separately to engineers and CD pipelines via folder-level bindings.

## Wire the Profiler agent in Spring Boot

Cloud Profiler ships as a JVM-loaded native agent (`profiler_java_agent.so`). Install it in the runtime image, then enable it through `JAVA_TOOL_OPTIONS` so the same image runs locally without profiling:

```dockerfile
# Dockerfile (additional layer; see docker.md for the base layout)
RUN curl -fsSL https://storage.googleapis.com/cloud-profiler/java/latest/profiler_java_agent.tar.gz \
    | tar -xzC /opt && chmod 0644 /opt/profiler_java_agent.so
```

Set the agent flags via `JAVA_TOOL_OPTIONS` on the Helm chart values or Cloud Run env:

```yaml
# helm/<service>/env/prd.yaml
env:
  PROFILING_ENABLED: "true"
  JAVA_TOOL_OPTIONS: >-
    -agentpath:/opt/profiler_java_agent.so=-cprof_service=<service>,-cprof_service_version=${K_REVISION}
```

Do **not** hard-code `cprof_service_version` to a release tag. Use the Cloud Run revision name (`K_REVISION`) or the Helm release revision so growth analysis can pin a single deploy and avoid mixing pre- and post-deploy heap state.

The agent reads the target project from the workload's Application Default Credentials -- no `-cprof_project_id` flag is needed when the workload runs in the application project. Pass `-cprof_project_id ent-<app>-<env>` only if the workload runs under credentials that resolve to a different project (e.g. a cross-project debug session).

## Wire the Profiler agent in Go

Use `cloud.google.com/go/profiler` directly. The agent uploads under Application Default Credentials, so the runtime SA's `roles/cloudprofiler.agent` binding is what admits writes.

```go
// cmd/<service>/main.go
import "cloud.google.com/go/profiler"

if cfg.ProfilingEnabled {
    if err := profiler.Start(profiler.Config{
        Service:        "<service>",
        ServiceVersion: cfg.ServiceVersion, // see SERVICE_VERSION below
        ProjectID:      cfg.ProjectID,
    }); err != nil {
        logging.Warn().Err(err).Msg("cloud profiler agent start failed; continuing")
    }
}
```

Do **not** fail startup if `profiler.Start` returns an error -- profiling is observability, not a runtime dependency. Log and continue.

For a working end-to-end example see the [entur/ai-portal-mcp](https://github.com/entur/ai-portal-mcp) `cmd/ai-portal-mcp/main.go` profiler block.

## Set runtime environment variables

Both Java and Go services read enablement and service version from the environment.

For Cloud Run, set the variables on the container spec:

```yaml
# cloudrun.yaml
env:
  - name: PROFILING_ENABLED
    value: "true"
  - name: GCP_PROJECT_ID
    value: ent-<app>-<env>   # the application's per-env project, never ent-kub-<env>
  # SERVICE_VERSION unset -> the agent picks up K_REVISION (Cloud Run injects it per revision)
```

For Kubernetes services, set the same variables on the Helm chart values -- see [common-helm.md](../platform/common-helm.md) for the env-var syntax. Use the Helm release revision (or a Helm-injected label carrying the image tag) as `SERVICE_VERSION`; do **not** leave it empty under Kubernetes since `K_REVISION` is a Cloud Run-only variable.

## Default to enabled in prd only

Keep `PROFILING_ENABLED=true` only in `prd`. Leave it unset (off) in `dev` and `tst`:

- Profiling traffic on a low-load dev environment skews the statistical sampling -- the resulting profile is mostly idle.
- The agent dials out to `cloudprofiler.googleapis.com` from local runs, which is noise during integration tests.

Do **not** enable profiling for short-lived workloads (cron jobs, batch tasks). The agent uploads on a ~minute cadence; processes that exit faster than that contribute no usable data.

## Profiler project routing

Cloud Profiler profiles land in the **application's** per-env project (`ent-<app>-<env>`) for every runtime, including Kubernetes. The agent stamps each upload with the project ID from `cfg.ProjectID` (Go), the `-cprof_project_id` flag (Java explicit form), or Application Default Credentials (Java + Go default) -- always the workload's own project.

This is **not** symmetric with Cloud Logging:

| Signal                  | Kubernetes workloads | Cloud Run / Firebase / DataProject |
|-------------------------|----------------------|------------------------------------|
| Traces (Cloud Trace)    | `ent-<app>-<env>`    | `ent-<app>-<env>`                  |
| Profiles (Cloud Profiler) | `ent-<app>-<env>`  | `ent-<app>-<env>`                  |
| Logs (Cloud Logging)    | `ent-kub-<env>`      | `ent-<app>-<env>`                  |

Cloud Logging writes Kubernetes logs under the kubelet's identity in the cluster host project, while the profiler and trace agents embedded in the workload report under the workload's SA in the application project. Configure profiler dashboards and the read-side MCP tools against `ent-<app>-<env>`.

## View profiles

- **Cloud Console**: Profiler in `ent-<app>-<env>`. The caller needs a Cloud Profiler read role bound at the folder level (typically `roles/cloudprofiler.user`).
- **Programmatic / agentic**: the [entur/ai-portal-mcp](https://github.com/entur/ai-portal-mcp) MCP server exposes `list_cloud_profiles`, `get_cloud_profile_topn`, and `detect_cloud_profile_growth`, callable from any MCP client under the caller's Google IAM. Use these when iterating from a coding agent rather than the console -- `detect_cloud_profile_growth` is the cheapest way to confirm whether a heap leak is real before opening the full profile in the UI.

## Further reading

- [observability.md](observability.md) -- health probes and Prometheus metrics.
- [tracing.md](tracing.md) -- OpenTelemetry and Cloud Trace setup; shares the same project-routing rule and IAM pattern as profiler.
- [logging.md](logging.md) -- structured log fields.
- [self-service.md](../platform/self-service.md) -- per-env project layout.
- [iam-roles.md](../platform/iam-roles.md) -- the assignable-roles allowlist that `roles/cloudprofiler.agent` is part of.
