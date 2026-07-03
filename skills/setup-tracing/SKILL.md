---
name: setup-tracing
description: >
    Wire distributed tracing into an Entur service following the golden path.
    Use when the user says "add tracing", "set up OpenTelemetry", "instrument
    for Cloud Trace", or asks to add distributed tracing to a Kotlin/Java or
    Go service.
---

# Set Up Tracing (Golden Path)

Wire distributed tracing into an Entur service so every inbound request produces a span in Cloud Trace, correlated with structured logs. This golden path covers **Kotlin/Java (Spring Boot, via the OpenTelemetry Java Agent)** and **Go (manual OpenTelemetry SDK)** only.

## Step 0: Confirm prerequisites -- do not guess any of these

Ask the user directly; do not infer silently and do not proceed until all four are answered.

1. **Language.** Detect from the repo: `build.gradle.kts` → Kotlin/Java, `go.mod` → Go. State the detected language and ask the user to confirm it.
   - If the repo is neither (Python, Node, etc.), **stop**. Tell the user this golden path only documents Java Agent instrumentation (Java/Kotlin) and manual SDK wiring (Go) -- do not improvise OpenTelemetry setup for another language.
2. **Cloud Trace storage (manual, console-only).** Ask: *"Have you already enabled Cloud Trace storage in the GCP Console (Monitoring → Trace Explorer → Enable trace storage) for every environment you're setting up, in the application project `ent-<app>-<env>` -- not the cluster host project `ent-kub-<env>`?"*
   - If no, or unsure, **stop** and tell them to do this first, once per environment (dev, tst, prd). This is a console-only action -- there is no Terraform resource for it yet. Do not attempt to script it.
   - Only proceed with the environments the user confirms are done. If they've only enabled it for `dev`, scope the rest of this skill to `dev`.
3. **App ID.** Needed to build `ent-<app>-<env>`. Read `metadata.id` from the self-service manifest under `.entur/*.yaml` if present. If the manifest exists but `metadata.id` is missing or empty, ask the user for it instead of guessing -- do not derive it from the repo name or any other field. If no manifest exists at all, ask the user directly.
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
  project            = module.init.app.project_id # "init" assumes that's this repo's terraform-google-init alias
  service            = each.value
  disable_on_destroy = false
}
```

If the repo has no `terraform/` directory yet, do not create one ad hoc -- tell the user to bootstrap Terraform with the Entur `terraform-google-init` module first.

## Step 2: Grant the workload service account `roles/cloudtrace.agent`

```hcl
# terraform/main.tf
resource "google_project_iam_member" "runtime_cloudtrace_agent" {
  project    = module.init.app.project_id # adapt "init" to this repo's module alias
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
- Use the pinned versions above as-is -- `v2.29.0` (agent) / `1.58.0-alpha` (gcp-auth-extension). Do not spend time checking upstream for a newer release. In Step 6's summary, tell the user which versions were used and point to the `otel` build stage in the Dockerfile as where to bump them later.
- `java25-debian13` is an example tag, not a guarantee -- match the distroless image version to the project's own Java toolchain (`build.gradle.kts`/`.tool-versions`), and confirm the resulting tag actually exists before using it. Flag the tag used in the Step 6 summary as something the user may need to change.
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

Run `go mod tidy` after pasting.

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

**If `helm/<app>/env/values-kub-ent-<env>.yaml` doesn't exist yet for a confirmed environment**, ask the user: *"No Helm values file exists yet for `<env>` -- do you want me to create `helm/<app>/env/values-kub-ent-<env>.yaml` with the tracing env vars?"* Only create it if they say yes, and only for the environment(s) confirmed in Step 0. Use this shape:

```yaml
# helm/<app>/env/values-kub-ent-<env>.yaml
common:
  env: <env>
  container:
    env:
      - name: OTEL_TRACES_EXPORTER
        value: otlp
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: https://telemetry.googleapis.com
      - name: OTEL_SERVICE_NAME
        value: <service-name>
      - name: TRACING_ENABLED
        value: "true"
      - name: GCP_PROJECT_ID
        value: ent-<app>-<env>
      - name: OTEL_TRACES_SAMPLER
        value: parentbased_always_on
      - name: OTEL_METRICS_EXPORTER
        value: none
      - name: OTEL_LOGS_EXPORTER
        value: none
```

If the file already exists with its own `common.container.env` entries, append these to the existing list instead of overwriting the file -- check for entries with the same `name` first and update their `value` in place rather than duplicating. Omit `OTEL_TRACES_SAMPLER` on Cloud Run, per the table below.

Only create the single env file this way -- do not scaffold `helm/<app>/Chart.yaml` or `helm/<app>/values.yaml` ad hoc. If those don't exist either, Helm hasn't been bootstrapped for this service at all, which is out of scope for this skill (same rule as Step 1's Terraform bootstrap).

**Kotlin/Java (Java Agent auto-configures from env vars):**

| Variable | Kubernetes | Cloud Run |
|---|---|---|
| `OTEL_TRACES_EXPORTER` | `otlp` | `otlp` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `https://telemetry.googleapis.com` | `https://telemetry.googleapis.com` |
| `OTEL_SERVICE_NAME` | `<service-name>` | `<service-name>` |
| `TRACING_ENABLED` | `true` | `true` |
| `GCP_PROJECT_ID` | `ent-<app>-<env>` | `ent-<app>-<env>` |
| `OTEL_TRACES_SAMPLER` | `parentbased_always_on` | **omit** |
| `OTEL_METRICS_EXPORTER` | `none` | `none` |
| `OTEL_LOGS_EXPORTER` | `none` | `none` |

Always set `OTEL_METRICS_EXPORTER=none` and `OTEL_LOGS_EXPORTER=none` on both runtimes -- the agent defaults both to `otlp` like traces, but Step 2's IAM grant doesn't authorize either.

Do not set `OTEL_TRACES_SAMPLER` on Cloud Run: the Cloud Run load balancer injects its own (thin) sampling decision upstream, and `parentbased_always_on` would defer to it, sampling far less than intended. Omitting the variable lets the agent default to `always_on`.

**Go (sampler/exporter already hardcoded in `tracing.Init`; only these two are needed):** set `TRACING_ENABLED` to `true` and `GCP_PROJECT_ID` to `ent-<app>-<env>`, the same on both Kubernetes and Cloud Run.

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

## Step 6: Tell the user what's left

Steps 0-5 are everything this skill can do by editing the repo. Verifying traces actually arrive requires a live deploy and real traffic, which happens outside this skill -- tell the user to:

1. Commit and merge the changes, and let the normal CD pipeline deploy to the confirmed environment(s) from Step 0.
2. Send a few requests to the deployed service to generate spans.
3. Check **Trace → Trace Explorer** in the GCP Console under `ent-<app>-<env>` (not `ent-kub-<env>`, even for Kubernetes workloads -- traces always land in the application project).
4. If no spans show up: re-check that Step 0's trace storage was enabled for the *deployed* environment, and that `TRACING_ENABLED`/`GCP_PROJECT_ID` reached the running container (a common miss is setting them in the wrong Helm values file for the target environment).

For Kotlin/Java, also state in the summary: the OpenTelemetry Java agent (`v2.29.0`) and gcp-auth-extension (`1.58.0-alpha`) versions were pinned as-is, not checked against upstream for something newer, and are set in the Dockerfile's `otel` build stage if the user wants to bump them later. Also state the distroless base image tag used and that it may need to change to match the project's Java toolchain.

## Critical Rules

- **The Step 0 trace-storage check is a manual console action, once per environment.** Never attempt to Terraform it; never proceed with Steps 1-5 for an environment the user hasn't confirmed is enabled.
- **Grant exactly `roles/cloudtrace.agent`** to the workload -- never a broader trace role.
- **`GCP_PROJECT_ID` is always the application project (`ent-<app>-<env>`)**, never the cluster host project, even on Kubernetes.
- **Java/Kotlin defaults to the Java Agent.** Only hand-roll manual OpenTelemetry instrumentation if the user gives a specific reason.
- **One `CMD` per Dockerfile.** Merge Cloud Profiler flags into the same `CMD` as the tracing agent flags if present.
- **Kubernetes vs Cloud Run sampler differs** -- see Step 4. Getting this backwards silently under-samples on Cloud Run or double-samples on Kubernetes.
- **Kotlin/Java also disables metrics/logs export** -- see Step 4. They default to `otlp` too if left unset, but Step 2 doesn't authorize either.
