# Distributed Tracing

Distributed tracing lets you follow a single request as it moves through a service and out to its dependencies, so you can see where time is spent and where failures occur. At Entur, trace spans are exported to Google Cloud Trace in a shared host project (ent-kub-<env>), rather than each team's own project — this is what makes it possible to correlate a trace with the logs from the same request.

## How to instrument an Entur service with OpenTelemetry tracing and export spans to Google Cloud Trace.

- **Target audience**: developers adding tracing to a new service or fixing tracing in an existing service.
- **Intent**: inbound requests produce traces in Cloud Trace Explorer, with logs correlated to their trace and span.
- **Scope**: Java Agent setup, manual instrumentation for non-Java services, sampling, log correlation, and trace viewing. Metrics and probes live in [observability.md](observability.md), structured logging in [logging.md](logging.md), and profiling in [profiler.md](profiler.md).
- **Prerequisites**: a `GoogleCloudApplication` manifest and a workload deployed on Kubernetes. IAM roles and the required Google Cloud APIs for tracing are provisioned automatically through the common Helm chart — no manual setup is needed.

For Java services, the Java Agent is the Golden Path. Instrument Java manually only when there is a specific reason to do so.

## 1. OpenTelemetry Instrumentation

### 1.1 Java Agent

The OpenTelemetry Java Agent instruments Spring Boot services at startup without code changes. It uses bytecode injection to capture inbound requests, outbound HTTP calls, and database calls automatically.

The agent is a JAR attached to the JVM at startup through the `-javaagent` flag. A multi-stage build is the recommended approach — it keeps the final image clean by using a temporary Alpine stage to download the JARs before copying them into the distroless runtime. If your team does not use multi-stage builds, you can download the JARs another way and copy them in, but the Dockerfile below shows the recommended pattern.

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

Check the [opentelemetry-java-instrumentation releases](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases) for the latest versions.

The two JARs serve different purposes and must both be present:

- `opentelemetry-javaagent.jar` is the OpenTelemetry Java agent, a single JAR maintained by the OpenTelemetry project that contains the full instrumentation library. It attaches to the JVM at startup and uses bytecode injection to automatically wrap HTTP servers, HTTP clients, database drivers, messaging systems, and more with tracing code. You never call it directly; the `-javaagent` JVM flag is what activates it.
- `gcp-auth-extension-*-shadow.jar` is the GCP auth extension, a separate JAR from the `opentelemetry-java-contrib` project. It does one thing: it intercepts outbound calls from the agent's OpenTelemetry Protocol (OTLP) exporter and attaches a valid GCP access token obtained via Application Default Credentials (ADC). Without it, requests to `telemetry.googleapis.com` would be rejected as unauthenticated. It contains all its dependencies shaded inside the shadow JAR, so it does not conflict with anything else on the classpath.

If you are also using the Google Cloud Profiler, see [profiler.md](profiler.md). The Profiler agent also runs as a JVM flag, so you need to merge the flags from both guides into a single `CMD` — do not have two separate `CMD` instructions, as only the last one takes effect.

### 1.2 Manuel Instrumentation

For manual instrumentation look at OpenTelemetry’s documentation [Java] (https://opentelemetry.io/docs/languages/java/)

## 2. Sampling

Sampling must be set explicitly per environment, do not rely on the default everywhere.

Our recommendation for Kubernetes:

```yaml
# values-kub-ent-dev.yaml
OTEL_TRACES_SAMPLER: "parentbased_always_on" # sample everything, easiest for debugging.
```

```yaml
# values-kub-ent-prd.yaml/values-kub-ent-tst.yaml
OTEL_TRACES_SAMPLER: "parentbased_traceidratio" # sample a fixed ratio of requests
OTEL_TRACES_SAMPLER_ARG: "0.1" # 10% of requests
```

These are our recommended defaults, not hard requirements - adjust `OTEL_TRACES_SAMPLER_ARG` if a service needs a different ratio.

| Calls per minute | Sampling ratio |
|-------------------|----------------|
| Fewer than 50      | 100%           |
| 50 to 100          | 50%            |
| 100 to 1000        | 10%            |
| More than 1,000    | 1%             |
| More than 5,000    | 0.01%          |

Recommended sampling ratio matrix.

If you want to apply more filters to your traces than just the sampler ratio, you can read more about the [Java Agent Sampler Extension](https://opentelemetry.io/docs/zero-code/java/agent/extensions/).

## 3. Correlate logs with traces

Cloud Logging joins log entries to traces in Trace Explorer when each log line includes a trace field and a spanId. Include these fields on every log line emitted inside a traced HTTP request - otherwise, logs and traces appear separately and cannot be correlated.

Each log entry must also include the required fields `timestamp` (ISO 8601), `severity`/`level`, and `message`/`msg`. For trace correlation, include these three fields on every log line emitted inside a traced request:

- `logging.googleapis.com/trace` - the full trace identifier as `traceId`, used by Cloud Logging to find the corresponding trace in Cloud Trace.
- `logging.googleapis.com/spanId` - identifies which specific span within the trace this log line belongs to.
- `logging.googleapis.com/trace_sampled` - whether this trace was sampled; without this, Cloud Logging may not surface the correlation in the UI even if the other fields are correct.

For more details, see [logging.md](logging.md).

**For Java and Kotlin Spring Boot services, no manual work is needed**. If [entur/cloud-logging](https://github.com/entur/cloud-logging) (v7.1.0) together with `spring-boot-starter-gcp-web` is set up, those two handle all of this automatically.

### Go

**For Go services, start by installing [entur/go-logging](https://github.com/entur/go-logging)**: `go get github.com/entur/go-logging`. Use `entur/go-logging` and extract the span context from the request context at the start of each handler. Pass the trace and spanId fields on every log call inside that handler:

```go
sc := trace.SpanFromContext(r.Context()).SpanContext()
logging.Info().
    Str("logging.googleapis.com/trace", sc.TraceID().String()).
    Str("logging.googleapis.com/spanId", sc.SpanID().String()).
    Bool("logging.googleapis.com/trace_sampled", sc.IsSampled()).
    Msg("handling /home request")
```

Apply the same pattern to error and warning log lines within the same handler. Health probe endpoints filtered from tracing do not need these fields.

**For other services see** [Correlate log entries](https://docs.cloud.google.com/logging/docs/view/correlate-logs).

## 4. View traces

1. Go to [GCP Console](https://console.cloud.google.com/welcome).
2. Select the host project in the top left corner - `ent-kub-<env>`.
3. Open the navigation menu on the left, select **Monitoring → Trace → Trace Explorer**.

Trace storage (the `_Trace` bucket) should provision automatically the first time a trace span is successfully written to the project. It's not instant, so don't expect it to appear in the Trace Explorer console within seconds - give it a few minutes.

Since traces from multiple applications land in the shared host project, filter Trace Explorer down to your own service before reading anything into it: use the filter bar and search by `service.name` to scope the view to just your application's spans.

***We are also currently working on a solution to view traces in Grafana.***

## Further reading

- [Migrate from the Trace exporter to the OTLP endpoint](https://docs.cloud.google.com/stackdriver/docs/instrumentation/migrate-to-otlp-endpoints)
- [Structured logging](logging.md)
- [Observability](observability.md)
- [Google Cloud Profiler](profiler.md)
- [Common Helm chart](../platform/common-helm.md)
