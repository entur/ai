# Distributed Tracing

How to instrument an Entur service with OpenTelemetry and ship spans to Google Cloud Trace.

- **Target audience**: developers adding a new service or fixing tracing on an existing one.
- **Intent**: every inbound request produces a trace visible in Cloud Trace Explorer, with `traceId` and `spanId` correlated to structured logs.
- **Scope**: per-project trace storage initialization, Terraform wiring (APIs + IAM), instrumentation strategy per language (Java Agent vs. manual OpenTelemetry SDK), runtime configuration, propagation, sampling, and project routing. Health probes, Prometheus metrics, and Cloud Profiler live in [observability.md](observability.md). Structured log fields live in [logging.md](logging.md).

This guide follows the instructions from [How-To Initialise Distributed Tracing](https://entur.atlassian.net/wiki/spaces/ESP/pages/6607110165/How-To+Initialise+Distributed+Tracing) in Confluence.

## Initialize Cloud Trace storage for the project

Cloud Trace Explorer requires per-project trace storage initialization before it retains spans. Until this is done the OpenTelemetry exporter ships spans successfully, Cloud Trace discards them, and the UI shows: *"Trace storage is not initialized for this project. Enable trace storage to begin collecting trace data."*

Initialize once per environment:

1. Open the Google Cloud Console for the **application's per-env project** (`ent-<app>-<env>`).
2. Open the navigation menu, then go to **Monitoring → Trace Explorer** under **Explore**.
3. Click **Enable trace storage**.
4. Repeat for every environment: `dev`, `tst`, and `prd`.

Do **not** initialize trace storage in the cluster host project (`ent-kub-<env>`). Cloud Trace stores spans in the project the workload reports under -- always the application project, even for Kubernetes workloads.

This step is currently a console click. The Cloud Trace API does not yet expose a Terraform-manageable resource for trace storage initialization.

## Enable the Cloud Trace and Telemetry APIs

Add both `cloudtrace.googleapis.com` and `telemetry.googleapis.com` to the per-env service activation block -- `cloudtrace.googleapis.com` covers the Cloud Trace backend, while `telemetry.googleapis.com` enables the OTLP ingestion endpoint every exporter (Java Agent or manual SDK) sends to. Some Google APIs are enabled implicitly as a side effect of creating other resources, but do **not** rely on that -- enabling them explicitly keeps the dependency clear and avoids breakage if the implicit behavior changes later.

```hcl
# terraform/main.tf
resource "google_project_service" "services" {
  for_each = toset([
    "cloudtrace.googleapis.com",
    "telemetry.googleapis.com",
    # ... other APIs your service needs
  ])

  project            = module.init.app.project_id
  service            = each.value
  disable_on_destroy = false
}
```

`module.init.app.project_id` resolves to the application's per-env project ID -- see [terraform-modules.md](../platform/terraform-modules.md#init-outputs) for the full module outputs.

If the repository has no Terraform yet, either bootstrap it with `terraform-google-init` first, or enable both APIs manually in the console (**APIs & Services → Library**, search "Cloud Trace API" and "Telemetry API", click **Enable** on each) as a stopgap. Manually enabled APIs are not tracked as code and can drift if someone disables them later -- prefer Terraform as soon as it exists.

## Grant the workload service account `roles/cloudtrace.agent`

The OpenTelemetry exporter writes spans under the workload's service account via Application Default Credentials -- no API keys required, credentials resolve from the runtime environment through Workload Identity. Bind exactly `roles/cloudtrace.agent`; it grants `cloudtrace.traces.patch` (write only) and nothing else, in line with least privilege.

```hcl
# terraform/main.tf
resource "google_project_iam_member" "runtime_cloudtrace_agent" {
  project    = module.init.app.project_id
  role       = "roles/cloudtrace.agent"
  member     = "serviceAccount:${module.init.service_accounts.runtime.email}" # "init" assumes that's this repo's terraform-google-init alias
  depends_on = [google_project_service.services]
}
```

Do **not** bind `roles/cloudtrace.user`, `roles/cloudtrace.admin`, or `roles/cloudtrace.editor` to the workload -- the exporter only writes. Read access is granted separately to engineers and CD pipelines via folder-level bindings. For now, you will need to ask for roles/cloudtrace.agent permission in #talk-utviklerplattform, but we are currently working on a better solution for this.

## Choose an instrumentation strategy per language

**For Java/Kotlin (Spring Boot), the OpenTelemetry Java Agent is the golden path.** Only hand-instrument manually if you have a specific reason to. For every other language (Go, Python, ...), instrument the OpenTelemetry SDK manually -- there is no equivalent bytecode-injection agent maintained for those runtimes.

### OpenTelemetry Java Agent (Spring Boot)

The Java Agent instruments Spring Boot services at startup without code changes. It uses bytecode injection to capture inbound requests, outbound HTTP calls, and database calls automatically, and is configured entirely through environment variables.

The agent is a JAR attached to the JVM at startup via `-javaagent`. Use a multi-stage Dockerfile: a temporary Alpine stage downloads the JARs, and the distroless final stage carries over only the JARs themselves, keeping the runtime image free of a shell or package manager.

```dockerfile
# Dockerfile
# temporary stage - only used to download JARs, never shipped
FROM alpine:3.24 AS otel
RUN mkdir /otel && \
    wget -q -O /otel/opentelemetry-javaagent.jar \
      https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v2.29.0/opentelemetry-javaagent.jar && \
    wget -q -O /otel/gcp-auth-extension.jar \
      https://repo1.maven.org/maven2/io/opentelemetry/contrib/opentelemetry-gcp-auth-extension/1.58.0-alpha/opentelemetry-gcp-auth-extension-1.58.0-alpha-shadow.jar

# final image - no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/java25-debian13:nonroot
WORKDIR /app
# only the JARs are carried over from the Alpine stage
COPY --from=otel /otel /otel
COPY build/libs/app.jar app.jar
# ENTRYPOINT in distroless is ["java"] - CMD entries are the arguments passed to that java process
CMD ["-javaagent:/otel/opentelemetry-javaagent.jar", \
     "-Dotel.javaagent.extensions=/otel/gcp-auth-extension.jar", \
     "-Dotel.javaagent.logging=application", \
     "-jar", "/app/app.jar"]
```

Check [opentelemetry-java-instrumentation releases](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases) for the latest versions before pinning.

Both JARs are required and serve different purposes:

- `opentelemetry-javaagent.jar` is the OpenTelemetry Java agent -- a single JAR maintained by the OpenTelemetry project containing the full instrumentation library. It attaches to the JVM at startup and wraps HTTP servers, HTTP clients, database drivers, and messaging systems with tracing code automatically. The `-javaagent` JVM flag is what activates it; nothing in application code calls it directly.
- `gcp-auth-extension-*-shadow.jar` is the GCP auth extension from `opentelemetry-java-contrib`. It intercepts outbound calls from the agent's OTLP exporter and attaches a GCP access token obtained from Workload Identity. Without it, requests to `telemetry.googleapis.com` are rejected as unauthenticated. Its dependencies are shaded inside the shadow JAR, so it does not conflict with anything else on the classpath.

If the service also runs Google Cloud Profiler, merge its `-javaagent` flags into this **same** `CMD` -- see [profiler.md](profiler.md). A Dockerfile only honors its last `CMD`; a second one silently disables the first instead of adding to it.

### Configure environment variables

``` yaml
# values-kub-ent-<env>.yaml
common:
  container:
    env:
      - name: OTEL_TRACES_EXPORTER # export traces using OTLP protocol
        value: "otlp"
      - name: OTEL_EXPORTER_OTLP_ENDPOINT # destination for exported traces
        value: "https://telemetry.googleapis.com"
      - name: OTEL_SERVICE_NAME # replace 'my-service' with your actual service name
        value: "my-service"
      - name: TRACING_ENABLED # Entur convention
        value: "true"
      - name: GCP_PROJECT_ID # read by your application to identify the target GCP project, replace with your service name
        value: "ent-<app>-<env>"
      - name: GOOGLE_CLOUD_PROJECT # read by Google libraries to identify the target GCP project, replace with your service name
        value: "ent-<app>-<env>"
      - name: OTEL_TRACES_SAMPLER # sampling flows from the upstream
        value: "parentbased_always_on"
      - name: OTEL_METRICS_EXPORTER # disable metrics export, metrics are handled separately
        value: "none"
      - name: OTEL_LOGS_EXPORTER # disable log export, logs are handled via Cloud Logging
        value: "none"
```

### Sampling

Sampling must be set explicitly per environment, do not rely on the default everywhere.
Our recommendation for Kubernetes:

```yaml
# values-kub-ent-dev.yaml
- name: OTEL_TRACES_SAMPLER
  value: "parentbased_always_on" # sample everything, easiest for debugging.
 ```

```yaml
# values-kub-ent-prd.yaml/values-kub-ent-tst.yaml
- name: OTEL_TRACES_SAMPLER
  value: "parentbased_traceidratio" # sample a fixed ratio of requests
- name: OTEL_TRACES_SAMPLER_ARG
  value: "0.1" # 10% of requests
```

These are our recommended defaults, not hard requirements - adjust OTEL_TRACES_SAMPLER_ARG if a service needs a different ratio.
If you want to filter traces beyond the sampler ratio, see the [Java Agent Sampler Extension](https://opentelemetry.io/docs/zero-code/java/agent/extensions/).

### Manual OpenTelemetry SDK instrumentation (Go and other non-Java services)

Manual instrumentation is not Go-specific -- the example below just happens to use Go. The same approach applies to any non-Java service (Python, Node, ...): initialize an OTLP exporter, wrap the outermost HTTP handler, and export via the OTLP gRPC exporter to `telemetry.googleapis.com`, Google's managed ingestion endpoint. The `roles/cloudtrace.agent` binding from the previous section is what admits writes, so confirm it is in place before deploying.

```go
// internal/tracing/tracing.go
package tracing

import (
    "context"
    "fmt"

    googlegrpc "google.golang.org/grpc/credentials/google"
    "google.golang.org/grpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

func Init(ctx context.Context, projectID, serviceName, serviceVersion string) (func(context.Context) error, error) {
    exp, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("telemetry.googleapis.com:443"),
        otlptracegrpc.WithDialOption(
            grpc.WithTransportCredentials(googlegrpc.NewDefaultCredentials()),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("otlp trace exporter: %w", err)
    }

    res, _ := resource.New(ctx, resource.WithAttributes(
        semconv.ServiceName(serviceName),
        semconv.ServiceVersion(serviceVersion),
    ))

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
        // sdktrace.WithSampler(sdktrace.AlwaysSample()), use this if your app runs on Cloud Run
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return tp.Shutdown, nil
}
```

`googlegrpc.NewDefaultCredentials()` bundles TLS encryption for the connection and Application Default Credentials for authentication into a single gRPC credential. On GCP, authentication resolves to the workload service account via Workload Identity automatically, so no API keys or credential files are needed in the container.

Wrap the outermost HTTP handler with `otelhttp.NewHandler` so every inbound request creates a span. Filter out probe and metrics paths so they do not drown the trace stream:

```go
// main.go
import (
    "net/http"
    "strings"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

handler := otelhttp.NewHandler(mux, "<service-name>", // replace with your service name
    otelhttp.WithFilter(func(r *http.Request) bool {
        return r.URL.Path != "/metrics" && !strings.HasPrefix(r.URL.Path, "/health/")
    }),
)
http.ListenAndServe(":8080", handler)
```

Always `defer span.End()` for any manual span created with `tracer.Start(ctx, ...)`. A span that is not ended never exports.

**Initialise tracing in main**
Creating `internal/tracing/tracing.go` defines the setup logic, but it does nothing until called. Add this at the top of `main()`, before registering any routes or handlers:

```go
// main.go
import (
  "context"
  "my-module-name/internal/tracing" // matches the module name in your go.mod
)

ctx := context.Background()
if os.Getenv("TRACING_ENABLED") == "true" {
    projectID := os.Getenv("GCP_PROJECT_ID")
    shutdown, err := tracing.Init(ctx, projectID, "my-service", "1.0.0")
    if err != nil {
        logging.Fatal().Err(err).Msg("failed to initialize tracing")
    }
    defer shutdown(ctx)
}
```

### Set runtime environment variables

Configure these for every runtime. The Java Agent is configured entirely through environment variables (no code changes); the Go SDK setup above hardcodes its exporter and sampler in code, so it only needs `TRACING_ENABLED` and `GCP_PROJECT_ID`.

```yaml
# values-kub-ent-<env>.yaml
common:
  env: <env>
  ingress:
    host: <host>
  container:
    env:
      - name: TRACING_ENABLED
        value: "true"
      - name: GCP_PROJECT_ID
        value: "ent-<app>-<env>" # the application's per-env project, never ent-kub-<env>
```

## Correlate logs with traces

Cloud Logging can join log entries to traces in Trace Explorer when each log line includes both a trace field in the format `projects/<project>/traces/<id>` and a `spanId`. Include these fields on every log line emitted inside a traced HTTP request; otherwise logs and traces appear separately and cannot be correlated.

Each log entry must also include the required fields `timestamp` (ISO 8601), `severity`/`level`, and `message`/`msg` -- see [logging.md](logging.md#required-fields). For trace correlation specifically, include these three fields on every log line emitted inside a traced request:

- `logging.googleapis.com/trace` -- the full trace identifier in the format `projects/<project>/traces/<id>`, used by Cloud Logging to find the corresponding trace in Cloud Trace.
- `logging.googleapis.com/spanId` -- identifies which specific span within the trace this log line belongs to.
- `logging.googleapis.com/trace_sampled` -- whether this trace was sampled; without it, Cloud Logging may not surface the correlation in the UI even if the other fields are correct.

### Java / Kotlin (Spring Boot)

Use [entur/cloud-logging](https://github.com/entur/cloud-logging) together with `spring-boot-starter-gcp-web`. The library injects `logging.googleapis.com/trace` and `logging.googleapis.com/spanId` automatically through Micrometer Tracing, so no manual field extraction is needed. Use standard SLF4J logging; the fields are included on log lines emitted within a traced request. ***We are currently making some changes to the cloud-logging repo, because the current setup will not work with OpenTelemetry.***

### Python

Use the standard `logging` module with `json_log_formatter.JSONFormatter()`. Extract the trace context from the OTel span and pass the fields through the `extra` parameter:

```python
from opentelemetry import trace

span = trace.get_current_span()
ctx = span.get_span_context()
project_id = os.environ["GCP_PROJECT_ID"]

logger.info("handling request", extra={
    "logging.googleapis.com/trace": f"projects/{project_id}/traces/{format(ctx.trace_id, '032x')}",
    "logging.googleapis.com/spanId": format(ctx.span_id, '016x'),
    "logging.googleapis.com/trace_sampled": ctx.trace_flags.sampled,
})
```

### Go

Install [entur/go-logging](https://github.com/entur/go-logging) (`go get github.com/entur/go-logging`), then extract the span context from the request context at the start of each handler and pass the trace and `spanId` fields on every log call inside that handler:

```go
sc := trace.SpanFromContext(r.Context()).SpanContext()
projectID := os.Getenv("GCP_PROJECT_ID")

logging.Info().
    Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, sc.TraceID())).
    Str("logging.googleapis.com/spanId", sc.SpanID().String()).
    Bool("logging.googleapis.com/trace_sampled", sc.IsSampled()).
    Msg("handling request")
```

Apply the same pattern to error and warning log lines within the same handler. Health probe endpoints filtered from tracing do not need these fields.

Set `GCP_PROJECT_ID` per environment (`ent-<app>-<env>`). Never use the cluster host project (`ent-kub-<env>`), because traces are stored in the application project.

## Trace project routing

Cloud Trace spans always land in the application's per-env project (`ent-<app>-<env>`), whether the workload runs on Kubernetes or Cloud Run. The exporter stamps each batch of spans with a projectID passed to the exporter (Go), or by reading the `GCP_PROJECT_ID` environment variable (Spring Boot). In both cases, this must always be the workload's application project, never the cluster host project.
  
## View traces

- **Cloud Console**: Trace → Trace Explorer in `ent-<app>-<env>`. The caller needs a Cloud Trace read role bound at the folder level (typically `roles/cloudtrace.user`).
