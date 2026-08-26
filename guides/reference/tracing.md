# Distributed Tracing

How to instrument an Entur service with OpenTelemetry and export spans to Google Cloud Trace.

- **Target audience**: developers adding tracing to a new service or fixing tracing in an existing service.
- **Intent**: inbound requests produce traces in Cloud Trace Explorer, with logs correlated to their trace and span.
- **Scope**: workload IAM, required APIs, Java Agent setup, manual instrumentation for non-Java services, runtime configuration, sampling, log correlation, and trace viewing. Metrics and probes live in [observability.md](observability.md), structured logging in [logging.md](logging.md), and profiling in [profiler.md](profiler.md).
- **Prerequisites**: an application project (`ent-<app>-<env>`), a `GoogleCloudApplication` manifest, and a deployed workload service account.

## 1. Grant the workload service account `roles/cloudtrace.agent`

The OpenTelemetry exporter needs permission to write trace data to Google Cloud Trace. `roles/cloudtrace.agent` provides write access to traces only, in line with least privilege.

Add the role to the default application service account in the `GoogleCloudApplication` manifest at `.entur/app-<service_name>.yaml`. Merge this fragment under `spec`:

```yaml
serviceAccounts:
  - id: application
    additionalRoles:
      - roles/cloudtrace.agent
```

Do not bind `roles/cloudtrace.user`, `roles/cloudtrace.admin`, or `roles/cloudtrace.editor` to the workload. The exporter only needs to write traces. Engineers and CD pipelines receive read access separately through folder-level bindings.

The application also needs these APIs enabled in every application project:

- `cloudtrace.googleapis.com`
- `telemetry.googleapis.com`

Team Plattform is working to enable them by default for all projects. Until then, initialize them in Terraform:

```hcl
# terraform/main.tf
resource "google_project_service" "tracing" {
  for_each = toset([
    "cloudtrace.googleapis.com",
    "telemetry.googleapis.com",
  ])

  project            = module.init.app.project_id
  service            = each.value
  disable_on_destroy = false
}
```

If the repository does not have Terraform yet, enable both APIs manually from **Google Cloud Console → APIs & Services → API Library**. Select the application project in the project selector before enabling them.

If the service is not written in Java, skip to [3. Instrument OpenTelemetry manually](#3-instrument-opentelemetry-manually). For Java services, the Java Agent is the Golden Path. Instrument Java manually only when there is a specific reason to do so.

## 2. Connect the Java Agent

The OpenTelemetry Java Agent instruments Spring Boot services at startup without code changes. It uses bytecode injection to capture inbound requests, outbound HTTP calls, database calls, messaging, and other supported operations automatically.

The agent is a JAR attached to the JVM at startup through the `-javaagent` flag. Use a multi-stage build so a temporary Alpine stage downloads the JARs and the final distroless image contains only the runtime files. If the service does not use multi-stage builds, download the JARs another way and copy them into the runtime image.

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
# ENTRYPOINT in distroless is ["java"] - CMD entries are arguments to that process
CMD ["-javaagent:/otel/opentelemetry-javaagent.jar", \
     "-Dotel.javaagent.extensions=/otel/gcp-auth-extension.jar", \
     "-Dotel.javaagent.logging=application", \
     "-jar", "/app/app.jar"]
```

Check the OpenTelemetry Java instrumentation releases before updating the pinned versions. Both JARs must be present:

- `opentelemetry-javaagent.jar` contains the instrumentation library maintained by the OpenTelemetry project. The `-javaagent` JVM flag activates it; application code does not call it directly.
- `gcp-auth-extension-*-shadow.jar` comes from `opentelemetry-java-contrib`. It intercepts calls from the agent's OTLP exporter and attaches a GCP access token obtained through Workload Identity. Without it, `telemetry.googleapis.com` rejects requests as unauthenticated. Its dependencies are shaded into the JAR to avoid classpath conflicts.

If the application also uses Google Cloud Profiler, merge the flags from [profiler.md](profiler.md) into the same `CMD`. Do not add a second `CMD`; Docker uses only the last one.

### Configure the Java Agent with environment variables

Keep JVM agent flags in the Dockerfile `CMD`. Configure the exporter and service identity through runtime environment variables.

#### Kubernetes

Put environment-independent settings in the common values file:

```yaml
# values.yaml
common:
  configmap:
    enabled: true
    data:
      OTEL_SERVICE_NAME: "<my-service>" # replace with the actual service name
      OTEL_TRACES_EXPORTER: "otlp" # export traces using OTLP
      OTEL_EXPORTER_OTLP_ENDPOINT: "https://telemetry.googleapis.com"
      OTEL_EXPORTER_OTLP_PROTOCOL: "grpc"
      OTEL_METRICS_EXPORTER: "none" # metrics are handled separately
      OTEL_LOGS_EXPORTER: "none" # logs are handled through Cloud Logging
      TRACING_ENABLED: "true" # Entur convention
```

Put project-specific settings in each environment values file:

```yaml
# values-kub-ent-<env>.yaml
common:
  configmap:
    data:
      OTEL_RESOURCE_ATTRIBUTES: "gcp.project_id=ent-<app>-<env>"
      GOOGLE_CLOUD_PROJECT: "ent-<app>-<env>"
      GCP_PROJECT_ID: "ent-<app>-<env>"
```

`GOOGLE_CLOUD_PROJECT` identifies the target project to Google libraries. `GCP_PROJECT_ID` is available to application code and logging configuration. Always use the application's per-environment project, never the Kubernetes cluster host project (`ent-kub-<env>`).

#### Cloud Run

Set the variables on the container specification:

```yaml
# cloudrun.yaml
container:
  env:
    - name: OTEL_TRACES_EXPORTER
      value: "otlp"
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "https://telemetry.googleapis.com"
    - name: OTEL_EXPORTER_OTLP_PROTOCOL
      value: "grpc"
    - name: OTEL_SERVICE_NAME
      value: "<my-service>"
    - name: TRACING_ENABLED
      value: "true"
    - name: GCP_PROJECT_ID
      value: "ent-<app>-<env>"
    - name: GOOGLE_CLOUD_PROJECT
      value: "ent-<app>-<env>"
```

`OTEL_TRACES_EXPORTER=otlp` sends traces to Google's managed OTLP endpoint. The GCP auth extension authenticates through Workload Identity, so the workload does not need API keys or credential files.

### Configure sampling

Set sampling explicitly for each Kubernetes environment instead of relying on the default everywhere.

Use this recommended development configuration:

```yaml
# values-kub-ent-dev.yaml
common:
  configmap:
    data:
      OTEL_TRACES_SAMPLER: "parentbased_always_on" # sample everything for debugging
```

Use ratio-based sampling in test and production:

```yaml
# values-kub-ent-tst.yaml or values-kub-ent-prd.yaml
common:
  configmap:
    data:
      OTEL_TRACES_SAMPLER: "parentbased_traceidratio"
      OTEL_TRACES_SAMPLER_ARG: "0.1" # sample 10% of requests
```

These are recommended defaults, not hard requirements. Adjust `OTEL_TRACES_SAMPLER_ARG` when a service needs a different ratio. Use the following matrix as a starting point:

| Calls per minute | Sampling ratio |
|------------------|----------------|
| Fewer than 50    | 100%           |
| 50 to 100        | 50%            |
| More than 1,000  | 1%             |
| More than 5,000  | 0.01%          |

For Cloud Run, do not set `OTEL_TRACES_SAMPLER` in any environment. The Cloud Run load balancer makes an upstream sampling decision and injects it into the trace context. Setting `parentbased_always_on` makes the agent honor that upstream decision, so the service can trace far fewer requests than intended. Omitting the variable lets the agent retain its `always_on` default at the agent level.

For filters beyond a sampler ratio, use the OpenTelemetry Java Agent Sampler Extension.

After configuring the Java Agent, skip to [4. Correlate logs with traces](#4-correlate-logs-with-traces). The next section is for non-Java services.

## 3. Instrument OpenTelemetry manually

Manual instrumentation is not Go-specific; the following example happens to use Go. Apply the same approach to any non-Java service.

The Go OTLP exporter sends traces over gRPC to `telemetry.googleapis.com`. Google's managed ingestion endpoint routes them into Cloud Trace in the application project. The `roles/cloudtrace.agent` binding from step 1 authorizes the workload service account to write traces, so apply it before deploying.

### 3a. Create the tracer provider and instrument the HTTP handler

Create the tracer provider and configure the OTLP exporter:

```go
// internal/tracing/tracing.go
package tracing

import (
    "context"
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
    "google.golang.org/grpc"
    googlegrpc "google.golang.org/grpc/credentials/google"
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

    res, err := resource.New(ctx, resource.WithAttributes(
        semconv.ServiceName(serviceName),
        semconv.ServiceVersion(serviceVersion),
        attribute.String("gcp.project_id", projectID),
    ))
    if err != nil {
        return nil, fmt.Errorf("create trace resource: %w", err)
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
        // Use sdktrace.AlwaysSample() instead if the application runs on Cloud Run.
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return tp.Shutdown, nil
}
```

`googlegrpc.NewDefaultCredentials()` combines TLS for the connection with Application Default Credentials for authentication. On GCP, authentication resolves to the workload service account through Workload Identity, so the container does not need API keys or credential files.

Wrap the outermost HTTP handler with `otelhttp.NewHandler` so every inbound request creates a span. Filter probe and metrics paths so they do not overwhelm the trace stream:

```go
// main.go
import (
    "net/http"
    "strings"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

handler := otelhttp.NewHandler(mux, "<service-name>", // replace with the service name
    otelhttp.WithFilter(func(r *http.Request) bool {
        return r.URL.Path != "/metrics" && !strings.HasPrefix(r.URL.Path, "/health/")
    }),
)
http.ListenAndServe(":8080", handler)
```

Always call `defer span.End()` for a manual span created with `tracer.Start(ctx, ...)`. An unended span is never exported.

### 3b. Initialize tracing in `main()`

Creating `internal/tracing/tracing.go` defines the setup logic, but does not run it. Call it near the start of `main()`, before registering routes or handlers:

```go
// main.go
import (
    "context"
    "os"

    "my-module-name/internal/tracing" // match the module name in go.mod
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

Replace `my-service` with the service name and `1.0.0` with the service version. `defer shutdown(ctx)` flushes buffered spans before the process exits; without it, spans created near shutdown may be lost.

### 3c. Set runtime environment variables

Configure the variables for every runtime. Both Spring Boot and Go services use the application project ID, and `TRACING_ENABLED` follows the Entur convention for toggling tracing.

#### Kubernetes

Use the common Helm chart environment-variable syntax:

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
        value: "ent-<app>-<env>" # application project, never ent-kub-<env>
```

#### Cloud Run

Set the same variables on the container specification:

```yaml
# cloudrun.yaml
common:
  env: <env>
  ingress:
    host: <host>
  container:
    env:
      - name: TRACING_ENABLED
        value: "true"
      - name: GCP_PROJECT_ID
        value: "ent-<app>-<env>" # application project, never ent-kub-<env>
```

## 4. Correlate logs with traces

Cloud Logging joins log entries to traces in Trace Explorer when each log line includes the trace and span fields. Include them on every log line emitted inside a traced request; otherwise, logs and traces appear separately.

Each entry must also include the required `timestamp` (ISO 8601), `severity`/`level`, and `message`/`msg` fields. For trace correlation, include these fields:

- `logging.googleapis.com/trace` contains the full trace identifier in the format `projects/<project>/traces/<id>`.
- `logging.googleapis.com/spanId` identifies the span within the trace.
- `logging.googleapis.com/trace_sampled` states whether the trace was sampled. Without it, Cloud Logging might not surface the correlation in the UI even when the other fields are correct.

See [logging.md](logging.md) for the complete structured logging requirements.

### Java / Kotlin (Spring Boot)

No manual correlation code is required. When [entur/cloud-logging v7.1.0](https://github.com/entur/cloud-logging/tree/v7.1.0) and `spring-boot-starter-gcp-web` are configured correctly, the library adds the Google trace fields through Micrometer Tracing. Use standard SLF4J within a traced request.

### Python

Use the standard `logging` module with `json_log_formatter.JSONFormatter()`. Extract the OpenTelemetry span context and pass the fields through `extra`:

```python
from opentelemetry import trace

span = trace.get_current_span()
ctx = span.get_span_context()
project_id = os.environ["GCP_PROJECT_ID"]

logger.info("handling request", extra={
    "logging.googleapis.com/trace": f"projects/{project_id}/traces/{format(ctx.trace_id, '032x')}",
    "logging.googleapis.com/spanId": format(ctx.span_id, "016x"),
    "logging.googleapis.com/trace_sampled": ctx.trace_flags.sampled,
})
```

### Go

Install [entur/go-logging](https://github.com/entur/go-logging) with `go get github.com/entur/go-logging`. Extract the span context from the request context and add the fields to every log call inside the handler:

```go
sc := trace.SpanFromContext(r.Context()).SpanContext()
projectID := os.Getenv("GCP_PROJECT_ID")

logging.Info().
    Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, sc.TraceID())).
    Str("logging.googleapis.com/spanId", sc.SpanID().String()).
    Bool("logging.googleapis.com/trace_sampled", sc.IsSampled()).
    Msg("handling /home request")
```

Apply the same pattern to error and warning logs within the handler. Health probe endpoints excluded from tracing do not need these fields.

Set `GCP_PROJECT_ID` to `ent-<app>-<env>` in every environment. Never use the cluster host project (`ent-kub-<env>`), because traces are stored in the application project.

## View traces

Open **Google Cloud Console**, select the application project (`ent-<app>-<env>`) in the project selector, and go to **Monitoring → Trace → Trace Explorer**.

The `_Trace` bucket provisions automatically when a trace span is first written successfully. Provisioning and ingestion can take a few minutes, so the trace might not appear immediately.

Cloud Trace spans always go to the application's per-environment project, whether the workload runs on Kubernetes or Cloud Run. The exporter reports the project through its resource configuration. Always use the workload's application project, never the Kubernetes cluster host project.

To view logs together with traces, create a trace scope in the application project that includes both the host project and the application project, then set that scope as the default. More detailed instructions will be added when the setup is finalized.

We is also working on a solution for viewing traces in Grafana.

## Further reading

- [Structured logging](logging.md)
- [Observability](observability.md)
- [Google Cloud Profiler](profiler.md)
- [Common Helm chart](../platform/common-helm.md)
