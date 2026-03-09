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

Use [entur/cloud-logging](https://github.com/entur/cloud-logging) -- plug-and-play structured JSON logging for GCP (no manual `logback.xml` needed):

```kotlin
// build.gradle.kts
dependencies {
    implementation(platform("no.entur.logging.cloud:bom:$cloudLoggingVersion"))
    implementation("no.entur.logging.cloud:spring-boot-starter-gcp-web")
    testImplementation("no.entur.logging.cloud:spring-boot-starter-gcp-web-test")
}
```

```java
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

private static final Logger LOG = LoggerFactory.getLogger(RouteService.class);

LOG.info("Route found for id {}", routeId);
LOG.error("Failed to fetch route {}", routeId, exception);
```

Optional starters: `request-response-spring-boot-starter-gcp-web` (Logbook HTTP body logging), `on-demand-spring-boot-starter-gcp-web` (buffer and flush only on failure). See [java.md](java.md) for full details.

### Go

Use [entur/go-logging](https://github.com/entur/go-logging):

```go
import "github.com/entur/go-logging"

logging.Info().Str("routeId", routeId).Str("origin", origin).Msg("route found")
logging.Error().Err(err).Str("routeId", routeId).Msg("failed to fetch route")
```

Handles JSON output, caller location, and GCP-compatible levels. Default level from `LOG_LEVEL` env var (defaults to `warning`). See [go.md](go.md) for full details.

### Python

Use standard `logging` with JSON formatting:

```python
import logging
import json_log_formatter

formatter = json_log_formatter.JSONFormatter()
handler = logging.StreamHandler()
handler.setFormatter(formatter)

logger = logging.getLogger("my_application")
logger.addHandler(handler)
logger.setLevel(logging.INFO)

logger.info("Route found", extra={"route_id": route_id, "origin": origin})
```

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
- **Never log** request/response bodies at INFO -- use DEBUG
- **Log at boundaries**: entering/exiting the system (HTTP requests, message consumption), not inside every method
- **Include context**: enough to trace back to a specific request or operation
- **Don't log and throw**: either log or propagate, not both
- **Prevent log injection**: encode user-supplied data to prevent log forging
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
