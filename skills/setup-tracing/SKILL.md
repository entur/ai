---
name: setup-tracing
description: >
  Instrument an Entur application with OpenTelemetry and ship traces to Google
  Cloud Trace, following the Entur tracing golden path. Covers Terraform
  (Cloud Trace API + IAM), the OpenTelemetry Java Agent for Spring Boot
  (Kotlin/Java), manual OpenTelemetry SDK wiring for Go, runtime env vars, and
  log-trace correlation. Use this skill when the user says "add tracing",
  "set up OpenTelemetry", "instrument for Cloud Trace", or asks to wire
  distributed tracing into a Kotlin/Java or Go service. Confirms the repo
  language and that Cloud Trace storage has already been enabled in the
  console before making any changes.
---

# Set Up Tracing (Golden Path)

Wire distributed tracing into an Entur service so every inbound request produces a span in Cloud Trace, correlated with structured logs. This golden path covers **Kotlin/Java (Spring Boot, via the OpenTelemetry Java Agent)** and **Go (manual OpenTelemetry SDK)** only.

## Step 0: Confirm prerequisites -- do not guess any of these

Ask the user directly; do not infer silently and do not proceed until all four are answered.

1. **Language.** Detect from the repo: `build.gradle.kts` → Kotlin/Java, `go.mod` → Go. State the detected language and ask the user to confirm it.
   - If the repo is neither (Python, Node, etc.), **stop**. Tell the user this golden path only documents Java Agent instrumentation (Java/Kotlin) and manual SDK wiring (Go) -- do not improvise OpenTelemetry setup for another language.
2. **Step 1 -- Cloud Trace storage.** Ask: *"Have you already enabled Cloud Trace storage in the GCP Console (Monitoring → Trace Explorer → Enable trace storage) for every environment you're setting up, in the application project `ent-<app>-<env>` -- not the cluster host project `ent-kub-<env>`?"*
   - If no, or unsure, **stop** and tell them to do this first, once per environment (dev, tst, prd). This is a console-only action -- there is no Terraform resource for it yet. Do not attempt to script it.
   - Only proceed with the environments the user confirms are done. If they've only enabled it for `dev`, scope the rest of this skill to `dev`.
3. **App ID.** Needed to build `ent-<appId>-<env>`. Read `metadata.id` from the self-service manifest under `.entur/*.yaml` if present; otherwise ask.
4. **Runtime.** Kubernetes (`helm/<app>/env/values-kub-ent-<env>.yaml`) or Cloud Run (`cloudrun.yaml`)? Determines where env vars in Step 4 are set.

## Step 1: Enable the Cloud Trace API in Terraform

Add `cloudtrace.googleapis.com` to the per-environment service-activation block:

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

If the repo has no `terraform/` directory yet, do not create one ad hoc -- tell the user to bootstrap Terraform with the Entur `terraform-google-init` module first.

## Step 2: Grant the workload service account `roles/cloudtrace.agent`

```hcl
# terraform/main.tf
resource "google_project_iam_member" "runtime_cloudtrace_agent" {
  project    = module.init.app.project_id
  role       = "roles/cloudtrace.agent"
  member     = "serviceAccount:${module.init.service_accounts.runtime.email}"
  depends_on = [google_project_service.services]
}
```

Bind exactly this role -- it is write-only. Never bind `roles/cloudtrace.user`, `.admin`, or `.editor` to the workload; read access is granted separately at the folder level.

## Step 3: Instrument the application

### Kotlin/Java -- OpenTelemetry Java Agent (default; do not hand-instrument without a specific reason)

Attach the agent via `-javaagent` in a multi-stage Dockerfile -- a temporary Alpine stage downloads the JARs, the distroless final stage only carries the JARs themselves:

```dockerfile
# Dockerfile
FROM alpine:3.24 AS otel
RUN mkdir /otel && \
    wget -q -O /otel/opentelemetry-javaagent.jar \
      https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v2.29.0/opentelemetry-javaagent.jar && \
    wget -q -O /otel/gcp-auth-extension.jar \
      https://repo1.maven.org/maven2/io/opentelemetry/contrib/opentelemetry-gcp-auth-extension/1.58.0-alpha/opentelemetry-gcp-auth-extension-1.58.0-alpha-shadow.jar

FROM gcr.io/distroless/java25-debian13:nonroot
WORKDIR /app
COPY --from=otel /otel /otel
COPY build/libs/app.jar app.jar
CMD ["-javaagent:/otel/opentelemetry-javaagent.jar", \
     "-Dotel.javaagent.extensions=/otel/gcp-auth-extension.jar", \
     "-Dotel.javaagent.logging=application", \
     "-jar", "/app/app.jar"]
```

- Both JARs are required: `opentelemetry-javaagent.jar` does the bytecode instrumentation; `gcp-auth-extension` attaches a Workload Identity token to outbound OTLP calls -- without it, `telemetry.googleapis.com` rejects the export as unauthenticated.
- Check the OpenTelemetry Java agent's own release notes for a newer version before pinning; `v2.29.0` (agent) / `1.58.0-alpha` (gcp-auth-extension) are current as of this guide.
- If the project also uses Cloud Profiler, merge its `-javaagent`/flags into this **same** `CMD` -- a Dockerfile only honors its last `CMD`, a second one silently disables the first.

### Go -- manual OpenTelemetry SDK wiring

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
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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
        // sdktrace.WithSampler(sdktrace.AlwaysSample()), use this instead if the service runs on Cloud Run
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return tp.Shutdown, nil
}
```

- `googlegrpc.NewDefaultCredentials()` bundles TLS + Application Default Credentials -- no API keys or credential files in the container; auth resolves through the Workload Identity binding from Step 2.
- Wrap the outermost handler so every inbound request gets a span, and filter probe/metrics paths so they don't drown the trace stream:

  ```go
  // main.go
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

- Call `tracing.Init` at the top of `main()`, guarded by `TRACING_ENABLED`, and always `defer shutdown(ctx)` so buffered spans flush before exit:

  ```go
  // main.go
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

- Any manual span from `tracer.Start(ctx, ...)` must `defer span.End()` -- an unended span never exports.

## Step 4: Set runtime environment variables

Set on Helm values (`values-kub-ent-<env>.yaml`, under `common.container.env`) for Kubernetes, or on the container spec (`cloudrun.yaml`) for Cloud Run. `GCP_PROJECT_ID` is always the application's per-env project (`ent-<app>-<env>`) -- never the cluster host project.

**Kotlin/Java (Java Agent auto-configures from env vars):**

| Variable | Kubernetes | Cloud Run |
|---|---|---|
| `OTEL_TRACES_EXPORTER` | `otlp` | `otlp` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `https://telemetry.googleapis.com` | `https://telemetry.googleapis.com` |
| `OTEL_SERVICE_NAME` | `<service-name>` | `<service-name>` |
| `TRACING_ENABLED` | `true` | `true` |
| `GCP_PROJECT_ID` | `ent-<app>-<env>` | `ent-<app>-<env>` |
| `OTEL_TRACES_SAMPLER` | `parentbased_always_on` | **omit** |

Do not set `OTEL_TRACES_SAMPLER` on Cloud Run: the Cloud Run load balancer injects its own (thin) sampling decision upstream, and `parentbased_always_on` would defer to it, sampling far less than intended. Omitting the variable lets the agent default to `always_on`.

**Go (sampler/exporter already hardcoded in `tracing.Init`; only these two are needed):**

| Variable | Kubernetes | Cloud Run |
|---|---|---|
| `TRACING_ENABLED` | `true` | `true` |
| `GCP_PROJECT_ID` | `ent-<app>-<env>` | `ent-<app>-<env>` |

Set the Go sampler in code to match the runtime, mirroring the Java table above: `ParentBased(AlwaysSample())` on Kubernetes, bare `AlwaysSample()` on Cloud Run.

## Step 5: Correlate logs with traces

Every log line inside a traced request needs `logging.googleapis.com/trace` (`projects/<project>/traces/<id>`), `logging.googleapis.com/spanId`, and `logging.googleapis.com/trace_sampled` -- otherwise Cloud Logging can't join it to the trace in Trace Explorer.

- **Kotlin/Java:** use `entur/cloud-logging` with `spring-boot-starter-gcp-web` and plain SLF4J logging. The fields are injected automatically via Micrometer Tracing -- no manual extraction.
- **Go:** `go get github.com/entur/go-logging`, then extract the span context per handler and attach the fields to every log call inside that handler (info/warn/error alike):

  ```go
  sc := trace.SpanFromContext(r.Context()).SpanContext()
  projectID := os.Getenv("GCP_PROJECT_ID")
  logging.Info().
      Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, sc.TraceID())).
      Str("logging.googleapis.com/spanId", sc.SpanID().String()).
      Bool("logging.googleapis.com/trace_sampled", sc.IsSampled()).
      Msg("handling request")
  ```

  Health-probe endpoints filtered out of tracing in Step 3 don't need these fields.

## Step 6: Verify

- GCP Console: **Trace → Trace Explorer** under `ent-<app>-<env>` (not `ent-kub-<env>`, even for Kubernetes workloads -- traces always land in the application project).
- Programmatically: if the entur-kompass MCP server is available, use `list_cloud_traces` / `get_cloud_trace` against the app's per-env project.

## Critical Rules

- **Step 1 (trace storage) is a manual console action, once per environment.** Never attempt to Terraform it; never proceed with Steps 1-5 for an environment the user hasn't confirmed is enabled.
- **Grant exactly `roles/cloudtrace.agent`** to the workload -- never a broader trace role.
- **`GCP_PROJECT_ID` is always the application project (`ent-<app>-<env>`)**, never the cluster host project, even on Kubernetes.
- **Java/Kotlin defaults to the Java Agent.** Only hand-roll manual OpenTelemetry instrumentation if the user gives a specific reason.
- **One `CMD` per Dockerfile.** Merge Cloud Profiler flags into the same `CMD` as the tracing agent flags.
- **Kubernetes vs Cloud Run sampler differs** -- see Step 4. Getting this backwards silently under-samples on Cloud Run or double-samples on Kubernetes.
- **Do not improvise instrumentation for languages outside Kotlin/Java and Go** -- this golden path doesn't cover them; say so and stop.
