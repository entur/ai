# Structured Logging Standards

All services must produce structured JSON logs to stdout. GCP Cloud Logging automatically ingests stdout from Kubernetes pods.

- **Target audience**: developers and AI agents adding or reviewing service logging.
- **Intent**: logs are structured, queryable, correlated with requests/traces, and safe to emit in GCP.
- **Scope**: required log fields, Java/Kotlin and Go logging setup, log levels, sensitive data rules, and operational logging patterns. Metrics and tracing live in [observability.md](observability.md) and [tracing.md](tracing.md).

## Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `timestamp` | ISO 8601 timestamp | `2025-01-15T10:30:00.123Z` |
| `severity` / `level` | Log level | `INFO`, `WARN`, `ERROR` |
| `message` / `msg` | Human-readable message | `"Request processed"` |

## Recommended Fields

| Field | Description | Example |
|-------|-------------|---------|
| `logger` | Logger name (class/package) | `no.entur.myapp.RouteService` |
| `logging.googleapis.com/trace` | Full GCP trace resource | `projects/ent-myapp-dev/traces/4bf92f...` |
| `logging.googleapis.com/spanId` | Current span ID | `00f067aa0ba902b7` |
| `logging.googleapis.com/trace_sampled` | Whether the trace is sampled | `true` |
| `correlationId` | HTTP request correlation ID | `req-abc-123` |
| `application` | Application name | `my-application` |
| `environment` | Runtime environment | `dev`, `tst`, `prd` |

## Implementation

### Java / Kotlin (Spring Boot)

Use [entur/cloud-logging v7.1.0](https://github.com/entur/cloud-logging/tree/v7.1.0) with Spring Boot 4.1.x. It provides structured JSON logging for GCP.
1. Remove any existing logging configuration. Delete application-owned logback.xml, logback-spring.xml, and logbook-test.xml (or equivalent) before adding the library. entur/cloud-logging is plug-and-play - a preexisting Logback config will conflict with it rather than merge with it.
2. Add the BOM, then the Spring Boot starter for how your service receives requests. The BOM only pins versions - every feature still needs its own artifact added explicitly.

```groovy
// build.gradle
ext {
    cloudLoggingVersion = '7.1.0'
}

dependencies {
    implementation platform("no.entur.logging.cloud:bom:${cloudLoggingVersion}")
    testImplementation platform("no.entur.logging.cloud:bom:${cloudLoggingVersion}")

    implementation "no.entur.logging.cloud:spring-boot-starter-gcp-web"
    testImplementation "no.entur.logging.cloud:spring-boot-starter-gcp-web-test"
}
```

For Spring gRPC, use spring-boot-starter-gcp-grpc-spring and spring-boot-starter-gcp-grpc-spring-test instead of the web starters. Both the main and test starter are required: the main starter gives machine-readable JSON for production, the test starter gives human-readable output for local development.

Optional starters:

- request-response-spring-boot-starter-gcp-web (and its -test counterpart) for Logbook HTTP request/response body logging.
- on-demand-spring-boot-starter-gcp-web to buffer detailed logs and flush them only when a request fails.

Use standard SLF4J from here (LoggerFactory.getLogger()) - no custom logger API is required for normal logging.

#### 3. Configure log levels

Configure the root and application log levels with Spring properties:

```properties
# application.properties
logging.level.root=INFO
logging.level.no.entur.myapp=WARN
```

Production runs at `INFO` by default per Entur convention. See [Log Levels](#log-levels) for level guidelines and the data that must never be logged, including secrets, PII, and payment details.

See [java.md](java.md) for full details.

### Go

Use [entur/go-logging](https://github.com/entur/go-logging). Provides JSON output, caller location, and GCP-compatible levels. Use `logging.Info().Str("key", value).Msg("message")` style. Default level from `LOG_LEVEL` env var (defaults to `warning`). See [go.md](go.md) for full details.

### Python

Use standard `logging` with `json_log_formatter.JSONFormatter()` for structured JSON output. Pass structured fields via the `extra` parameter.

## Log Levels

| Level | Use for | Example |
|-------|---------|---------|
| `ERROR` | Unexpected failures requiring attention | Database connection lost, unhandled exception |
| `WARN` | Expected but unusual conditions | Retry attempt, deprecation warning, rate limit approached |
| `INFO` | Normal operational events | Request processed, job completed, startup/shutdown |
| `DEBUG` | Diagnostic detail for troubleshooting | Query parameters, cache hit/miss, detailed flow |

### Guidelines

- Production runs at `INFO` by default
- **Never log** secrets, tokens, passwords, or PII
- **Never log** payment details (PCI-DSS)
- **ALWAYS use DEBUG level** for request/response body logging
- **ALWAYS log at boundaries**: entering/exiting the system (HTTP requests, message consumption)
- **ALWAYS include context**: enough to trace back to a specific request or operation
- **ALWAYS choose one**: either log the error or propagate it -- handling at both levels creates noise
- **ALWAYS encode user-supplied data** in logs to prevent log injection/forging
- Session tokens in logs only in irreversible hashed form

### Security Events to Log

- Successful and failed authentication attempts
- Access control failures
- Input validation failures
- Deserialization failures
- Application startup and shutdown

## Correlation

### Distributed Tracing

Once distributed tracing and cloud logging is set up for your application, logs and traces are correlated automatically and no manual field extraction is needed. Cloud Logging joins logs to traces when a structured log entry contains the full `logging.googleapis.com/trace` resource and `logging.googleapis.com/spanId`. Include `logging.googleapis.com/trace_sampled` when sampling information is available. 
Version 7.1.0 supports both Java tracing setups documented in [tracing.md](tracing.md).

Do not copy trace fields into MDC manually. Use standard SLF4J inside the traced request; cloud-logging selects the correct mapping automatically. Its legacy correlation-ID fallback remains active when no OpenTelemetry trace context exists.

See [tracing.md](tracing.md) and for instrumentation, propagation, sampling, API enablement, and IAM setup.

## 5. View logs

Open **GCP Console → Logging → Logs Explorer** in the cluster host project (`ent-kub-<env>`), not the application project. Kubernetes logs are written to the cluster host project by the kubelet.

Filter logs using fields such as:

```text
resource.labels.namespace_name="my-application"
resource.labels.pod_name="my-application-..."
jsonPayload.logger="no.entur.myapp.RouteService"
```

This project routing differs from Cloud Trace: Kubernetes logs are stored in `ent-kub-<env>`, while traces are stored in `ent-<app>-<env>`.
