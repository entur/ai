# Structured Logging Standards

All services must produce structured JSON logs to stdout. GCP Cloud Logging automatically ingests stdout from Kubernetes pods.

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
| `traceId` | Distributed trace ID | `abc123def456` |
| `spanId` | Current span ID | `789ghi` |
| `requestId` | HTTP request correlation ID | `req-abc-123` |
| `application` | Application name | `my-application` |
| `environment` | Runtime environment | `dev`, `tst`, `prd` |

## Implementation

### Java / Kotlin (Spring Boot)

Use [entur/cloud-logging](https://github.com/entur/cloud-logging) -- plug-and-play structured JSON logging for GCP (no manual `logback.xml` needed). Add the BOM and `spring-boot-starter-gcp-web` dependency, then use standard SLF4J (`LoggerFactory.getLogger()`). Optional starters: `request-response-spring-boot-starter-gcp-web` (Logbook HTTP body logging), `on-demand-spring-boot-starter-gcp-web` (buffer and flush only on failure). See [java.md](java.md) for full details.

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

- Spring Boot: use Micrometer Tracing for trace context propagation
- Go: propagate `X-Cloud-Trace-Context` header
- Include `traceId` and `spanId` in all log entries

### Request IDs

- Generate unique request ID at ingress if not in `X-Request-ID` header
- Propagate through all downstream calls
- Include in all log entries and error responses
