# Distributed Tracing

How to instrument an Entur service with OpenTelemetry and ship spans to Google Cloud Trace.

- **Target audience**: developers adding a new service or fixing tracing on an existing one.
- **Intent**: every inbound request produces a trace visible in Cloud Trace Explorer, with `traceId` and `spanId` correlated to structured logs.
- **Scope**: per-project trace storage initialization, Terraform wiring (API + IAM), application setup (Spring Boot and Go), propagation, sampling, project routing. Health probes, Prometheus metrics, and Cloud Profiler live in [observability.md](observability.md). Structured log fields live in [logging.md](logging.md).
- **Prerequisites**: the application is provisioned via self-service YAML in `.entur/` (see [self-service.md](../platform/self-service.md)) and emits structured JSON logs (see [logging.md](logging.md)).

## Initialize Cloud Trace storage for the project

Cloud Trace Explorer requires per-project trace storage initialization before it retains spans. Until this is done the OpenTelemetry exporter ships spans successfully, Cloud Trace discards them, and the UI shows: *"Trace storage is not initialized for this project. Enable trace storage to begin collecting trace data."*

Initialize once per environment:

1. Open the Google Cloud Console **Trace → Trace Explorer** view in the **application's per-env project** (`ent-<app>-<env>`).
2. Click **Enable trace storage**.

Do **not** initialize trace storage in the cluster host project (`ent-kub-<env>`). Cloud Trace stores spans in the project the workload reports under -- always the application project, even for Kubernetes workloads. See [Trace project routing](#trace-project-routing).

This step is currently a console click. The Cloud Trace API does not yet expose a Terraform-manageable resource for trace storage initialization; track the gap in `#talk-utviklerplattform` and re-evaluate when GCP exposes a `google_trace_*` resource.

## Enable the Cloud Trace API

Add `cloudtrace.googleapis.com` to the per-env service activation block. Do **not** rely on the API being enabled implicitly by another resource -- `google_project_service` is the only source of truth.

```hcl
# terraform/main.tf
resource "google_project_service" "services" {
  for_each = toset([
    "cloudtrace.googleapis.com",
    # ... other APIs your service needs
  ])

  project            = module.init.app.project_id
  service            = each.value
  disable_on_destroy = false
}
```

`module.init.app.project_id` resolves to the application's per-env project ID -- see [terraform-modules.md](../platform/terraform-modules.md#init-outputs) for the full module outputs.

## Grant the workload service account `roles/cloudtrace.agent`

The OpenTelemetry exporter writes spans under the workload's service account via Application Default Credentials. Bind exactly `roles/cloudtrace.agent` -- it grants `cloudtrace.traces.patch` (write only) and nothing else.

```hcl
# terraform/main.tf
resource "google_project_iam_member" "runtime_cloudtrace_agent" {
  project = module.init.app.project_id
  role    = "roles/cloudtrace.agent"
  member  = "serviceAccount:${google_service_account.runtime.email}"
}
```

Do **not** bind `roles/cloudtrace.user`, `roles/cloudtrace.admin`, or `roles/cloudtrace.editor` to the workload -- the exporter only writes. Read access is granted separately to engineers and CD pipelines via folder-level bindings.

`roles/cloudtrace.agent` is not yet on the [iam-roles.md](../platform/iam-roles.md) allowlist; request its addition via `#talk-utviklerplattform` before the platform policy guard rejects the binding.

## Wire OpenTelemetry in Spring Boot

Add the Micrometer Tracing bridge and the OTLP exporter to `build.gradle.kts`:

```kotlin
// build.gradle.kts
dependencies {
    implementation("io.micrometer:micrometer-tracing-bridge-otel")
    implementation("io.opentelemetry:opentelemetry-exporter-otlp")
}
```

Set sampling probability in `application.yml`:

```yaml
# application.yml
management:
  tracing:
    sampling:
      probability: 1.0   # 100% in dev/tst; lower in prd only if span volume exceeds Cloud Trace's free tier
```

Do **not** add a separate trace-context filter -- Micrometer Tracing wires propagation through Spring Boot's auto-configuration.

## Wire OpenTelemetry in Go

Use the Google Cloud OpenTelemetry exporter directly. It targets Cloud Trace under Application Default Credentials, so the runtime SA's `roles/cloudtrace.agent` binding is what admits writes.

```go
// internal/tracing/tracing.go
package tracing

import (
    "context"
    "fmt"

    texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
    gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
    "google.golang.org/api/option"
)

func Init(ctx context.Context, projectID, serviceName, serviceVersion string) (func(context.Context) error, error) {
    exp, err := texporter.New(
        texporter.WithProjectID(projectID),
        texporter.WithTraceClientOptions([]option.ClientOption{option.WithTelemetryDisabled()}),
    )
    if err != nil {
        return nil, fmt.Errorf("cloud trace exporter: %w", err)
    }

    res, _ := resource.New(ctx, resource.WithAttributes(
        semconv.ServiceName(serviceName),
        semconv.ServiceVersion(serviceVersion),
    ))

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        gcppropagator.CloudTraceOneWayPropagator{},
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return tp.Shutdown, nil
}
```

Wrap the outermost HTTP handler with `otelhttp.NewHandler` so every inbound request produces a span. Filter probe and metrics paths so they do not drown the trace stream:

```go
// cmd/<service>/main.go
import (
    "net/http"
    "strings"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

handler := otelhttp.NewHandler(mux, "<service-name>",
    otelhttp.WithFilter(func(r *http.Request) bool {
        return r.URL.Path != "/metrics" && !strings.HasPrefix(r.URL.Path, "/health/")
    }),
)
http.ListenAndServe(":8080", handler)
```

Always `defer span.End()` on any manual span created via `tracer.Start(ctx, ...)`. An unended span never exports.

`option.WithTelemetryDisabled()` on the exporter's client options is mandatory. Without it, the `google.golang.org/api/transport/grpc` package auto-attaches `otelgrpc.NewClientHandler()` to the exporter's own gRPC dial. Every `BatchWriteSpans` export then records itself as a root span and ships on the next exporter tick -- a self-feeding loop that fills `list_cloud_traces` results with exporter noise (one root span per ~5s tick) and burns the 2.5M-spans/month free-tier budget. The option scopes only to this client and does not disable any other instrumentation in the app.

For a working end-to-end example see the [entur/ai-portal-mcp](https://github.com/entur/ai-portal-mcp) `internal/tracing` package.

## Set runtime environment variables

Both Spring Boot and Go services read the target project ID from the environment.

For Cloud Run, set the variables on the container spec:

```yaml
# cloudrun.yaml
env:
  - name: TRACING_ENABLED
    value: "true"
  - name: GCP_PROJECT_ID
    value: ent-<app>-<env>   # the application's per-env project, never ent-kub-<env>
```

For Kubernetes services, set the same variables on the Helm chart values -- see [common-helm.md](../platform/common-helm.md) for the env-var syntax.

## Propagate trace context

Use a composite TextMapPropagator with three propagators, in this order on extract (first match wins):

- **CloudTraceOneWayPropagator** -- reads the `X-Cloud-Trace-Context` header that Cloud Run injects on inbound requests, but does not write it on outbound calls.
- **TraceContext** (W3C) -- standard `traceparent` and `tracestate` headers.
- **Baggage** -- standard `baggage` header.

This lets inbound Cloud Run traffic continue an upstream trace while outbound HTTP calls speak modern W3C. Do **not** drop W3C in favor of `X-Cloud-Trace-Context` alone -- the GCP header is non-standard and breaks federation with any non-GCP callee.

## Pick the sampler to match your runtime

Use bare `AlwaysSample()` on Cloud Run services that include `CloudTraceOneWayPropagator` (see [Propagate trace context](#propagate-trace-context)). Use `ParentBased(AlwaysSample())` everywhere else.

The Cloud Run load balancer injects an `X-Cloud-Trace-Context` header carrying its own sample decision -- ~0.1 % of inbound requests. `CloudTraceOneWayPropagator` reads that header, and `ParentBased` then honors the LB's bit, so your app traces almost nothing despite asking for 100 %. Bare `AlwaysSample` makes a local decision per request. Trace continuity from the inbound headers (`trace_id`, `parent_span_id`) is preserved independently of the sampler, so cross-service traces remain linked.

On Kubernetes / GKE, or any service without Cloud Run's LB upstream, there is no synthetic sample bit to inherit and `ParentBased` extends existing sampled traces from real upstream W3C callers (other Entur services) into yours without splitting them.

Switch the inner sampler to `TraceIDRatioBased(0.1)` only when sustained span volume exceeds Cloud Trace's free tier (2.5M spans/month) -- keep the same wrapper (bare vs `ParentBased`) you picked above.

## Correlate logs with traces

Cloud Logging joins logs to traces in Trace Explorer when each log entry carries both `trace` (`projects/<project>/traces/<id>`) and `spanId`. Emit them from every log line inside a traced request. See [logging.md](logging.md#recommended-fields) for the field schema and the Spring Boot / Go integration points.

## Trace project routing

Cloud Trace spans land in the **application's** per-env project (`ent-<app>-<env>`) for every runtime, including Kubernetes. The exporter stamps each batch with the project ID passed to `texporter.WithProjectID` (Go) or the `GOOGLE_CLOUD_PROJECT` / `GCP_PROJECT_ID` environment variable (Spring Boot) -- always the workload's own project.

This is **not** symmetric with Cloud Logging:

| Signal               | Kubernetes workloads | Cloud Run / Firebase / DataProject |
|----------------------|----------------------|------------------------------------|
| Traces (Cloud Trace) | `ent-<app>-<env>`    | `ent-<app>-<env>`                  |
| Profiles (Profiler)  | `ent-<app>-<env>`    | `ent-<app>-<env>`                  |
| Logs (Cloud Logging) | `ent-kub-<env>`      | `ent-<app>-<env>`                  |

Cloud Logging writes Kubernetes logs under the kubelet's identity in the cluster host project, while the trace and profiler agents embedded in the workload report under the workload's SA in the application project. Configure dashboards and trace links against `ent-<app>-<env>`.

## View traces

- **Cloud Console**: Trace → Trace Explorer in `ent-<app>-<env>`. The caller needs a Cloud Trace read role bound at the folder level (typically `roles/cloudtrace.user`).
- **Programmatic / agentic**: the [entur/ai-portal-mcp](https://github.com/entur/ai-portal-mcp) MCP server exposes `list_cloud_traces` and `get_cloud_trace`, callable from any MCP client under the caller's Google IAM. Use these when iterating from a coding agent rather than the console.

## Further reading

- [observability.md](observability.md) -- health probes, Prometheus metrics, Cloud Profiler.
- [logging.md](logging.md) -- structured log fields and trace-correlation field names.
- [self-service.md](../platform/self-service.md) -- per-env project layout and orchestrator manifest fields.
- [iam-roles.md](../platform/iam-roles.md) -- the assignable-roles allowlist that `roles/cloudtrace.agent` needs to join.
