# Observability Standards

All Entur services must expose health checks, Prometheus metrics, and distributed tracing.

## Health Checks

Every service must expose liveness and readiness probes for Kubernetes pod lifecycle management.

### Liveness Probe

Answers "is the process running and not deadlocked?"

- Return `200 OK` if alive
- Do NOT check external dependencies -- a slow database should not cause restarts
- Default path: `/actuator/health/liveness` (Spring Boot) or custom path for Go/Python

### Readiness Probe

Answers "is the application ready to serve traffic?"

- Return `200 OK` if ready, `503` if a private dependency is down
- Check only **private resources** (own DB connection pool, internal cache)
- **Never check shared/external services** -- all pods removed from routing simultaneously if shared service is down
- Default path: `/actuator/health/readiness` (Spring Boot) or custom path for Go/Python

For Helm probe configuration, see [helm.md](helm.md#health-probes).

## Prometheus Metrics

### Entur Metrics Starter (Spring Boot)

Use `org.entur.metrics:metrics-spring-boot-starter`. Provides autoconfiguration for Prometheus with Entur defaults assumed by [Grafana dashboards](https://grafana.entur.org). **Do not change default metric names** -- dashboards depend on them.

#### Setup

```kotlin
// build.gradle.kts
dependencies {
    implementation("org.entur.metrics:metrics-spring-boot-starter:<version>")  // check Artifactory for latest
}
```

Published to Entur's JFrog Artifactory. The `micrometer-registry-prometheus` dependency is included transitively.

```yaml
# application.yml
management:
  endpoints:
    web:
      exposure:
        include: health, info, prometheus
  metrics:
    tags:
      application: ${spring.application.name}
```

#### What the Starter Provides

Via `CustomMetricsAutoConfiguration`:

- **Custom histogram buckets** for HTTP server requests and Kafka consumer delay -- Apdex-friendly SLO boundaries (200ms, 800ms, 2.1s, 4s) with distribution factors (0.01x-32x of 800ms). Bucket count < 20 to control cardinality.
- **DCI tagging** on HTTP server metrics -- extracts `Entur-Distribution-Channel` header as `DCI` tag. Defaults to `"Unknown"` if absent.
- **URI filtering** -- drops metrics for `/actuator/**`, `/v3/api-docs`, `/favicon.ico`
- **Cache tag enrichment** -- parses `app-domain-cacheName` format to add `domain` and `name` tags

#### Standard Metric Names

Constants from `org.entur.metrics.config.Defaults`:

| Constant | Metric Name | Description |
|----------|-------------|-------------|
| `HTTP_SERVER_REQUESTS` | `http.server.requests` | HTTP server request duration (auto-registered) |
| `HTTP_CLIENT_REQUESTS` | `http.client.requests` | HTTP client request duration |
| `QUARTZ_JOB` | `quartz.job` | Quartz job execution time |
| `KAFKA_CONSUMER_PROCESS_TIME` | `kafka.consumer.consume.elapsed` | Kafka message processing time |
| `KAFKA_CONSUMER_CONSUME_DELAY` | `kafka.consumer.consume.delay` | Delay from production to consumption |

#### Kafka Consumer Metrics

Record processing time and consumption delay using standard metric names. Works alongside automatic Micrometer listeners (see [kafka.md](kafka.md#observability)):

```kotlin
// Processing time -- annotate the listener method
@Timed(
    value = KAFKA_CONSUMER_PROCESS_TIME,
    percentiles = [0.50, 0.75, 0.95, 0.99],
    extraTags = ["source", "MY_APP"]
)
@KafkaListener(topics = ["my-topic"], containerFactory = "enturListenerFactory")
fun onEvent(@Payload event: MyEvent) {
    processEvent(event)
}
```

```kotlin
// Consumption delay -- call as the first step in each consumer
private fun logConsumeDelay(eventTimestamp: String, topicEvent: String, partition: Int) {
    val timestamp = ZonedDateTime.parse(eventTimestamp)
    val differenceMs = ChronoUnit.MILLIS.between(timestamp, ZonedDateTime.now())
    Timer.builder(KAFKA_CONSUMER_CONSUME_DELAY)
        .tag("eventType", topicEvent)
        .tag("partition", partition.toString())
        .publishPercentiles(0.5, 0.75, 0.95, 0.99)
        .register(meterRegistry)
        .record(differenceMs, TimeUnit.MILLISECONDS)
}
```

#### Quartz Job Metrics

```kotlin
@Timed(
    value = QUARTZ_JOB,
    percentiles = [0.50, 0.75, 0.95, 0.99],
    extraTags = ["job", "MyJobName"]
)
override fun executeInternal(context: JobExecutionContext) {
    registerFireDelay(context, "MyJobName", meterRegistry)
    // ... job logic
}

private fun registerFireDelay(context: JobExecutionContext, jobName: String, meterRegistry: MeterRegistry) {
    val scheduleTime = context.scheduledFireTime.time.toDouble()
    val fireTime = context.fireTime.time
    DistributionSummary.builder("${QUARTZ_JOB}.fire.delay")
        .baseUnit("ms")
        .description("Delay between scheduled fire time and actual fire time")
        .tag("job", jobName)
        .register(meterRegistry)
        .record(fireTime - scheduleTime)
}
```

### Enabling Metrics (Non-Spring-Boot)

#### Go

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

mux.Handle("GET /metrics", promhttp.Handler())
```

#### Python

```python
from prometheus_client import start_http_server, Counter

REQUEST_COUNT = Counter('http_requests_total', 'Total HTTP requests', ['method', 'path', 'status'])
```

### Metrics Helm Configuration

See [helm.md](helm.md#prometheus-metrics) for Prometheus Helm values.

### Standard Metrics

All services should expose at minimum:

| Metric | Type | Description |
|--------|------|-------------|
| `http_server_requests_seconds` | Histogram | Request duration by method, path, status |
| `jvm_memory_used_bytes` | Gauge | JVM memory usage (Java/Kotlin) |
| `process_cpu_usage` | Gauge | CPU usage |
| `db_pool_active_connections` | Gauge | Database connection pool usage |

Spring Boot with Actuator and Entur metrics starter provides all of these automatically.

### Custom Metrics

Follow Prometheus naming conventions:

- `snake_case` with unit suffix: `_seconds`, `_bytes`, `_total`
- Use labels for dimensions (e.g. `route`, `status`)
- Keep cardinality low -- avoid high-cardinality labels (user IDs, request IDs)

```java
// Java/Kotlin
Counter.builder("routes_processed_total")
    .description("Total routes processed")
    .tag("status", "success")
    .register(meterRegistry)
    .increment();
```

```go
// Go
routesProcessed := promauto.NewCounterVec(prometheus.CounterOpts{
    Name: "routes_processed_total",
    Help: "Total routes processed",
}, []string{"status"})

routesProcessed.WithLabelValues("success").Inc()
```

## Distributed Tracing

### Tracing with Spring Boot

Use Micrometer Tracing with OpenTelemetry:

```yaml
# application.yml
management:
  tracing:
    sampling:
      probability: 1.0    # 100% sampling in dev/tst, lower in prd

# Add dependencies:
# micrometer-tracing-bridge-otel
# opentelemetry-exporter-otlp
```

### Tracing with Go

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("my-service")
ctx, span := tracer.Start(ctx, "findRoute")
defer span.End()
```

### Trace Propagation

- Use W3C Trace Context headers (`traceparent`, `tracestate`)
- Include `traceId` in all log entries for log-trace correlation
- Google Cloud Trace ingests traces from the OpenTelemetry exporter

## Google Cloud Profiler

Continuous, low-overhead production profiling for CPU, memory allocation, and lock contention.

### Enabling Profiler

1. Enable the Profiler API via `.entur` directory trigger
2. Attach the profiler agent:

```bash
java \
  -agentpath:/path/to/profiler_java_agent.so \
  -Dcom.google.cprof.service=my-service-name \
  -Dcom.google.cprof.service_version=1.0.0 \
  -cprof_project_id <your_app_project_id> \
  -jar app.jar
```

Use the **application's own GCP project ID**, not the cluster project. View in Cloud Console under **Profiler**.

## Grafana Dashboards

Key dashboards at `grafana.entur.org`:

- **VPA recommendations**: "kubernetes-vpa-recommendations" -- select cluster, namespace, target for recommended CPU/memory
- **PDB compliance**: "kubernetes-poddisruptionbudget" -- find deployments missing PDB
- **Traffic per service**: "traffic-pr-service" -- verify traffic before deprecating

CLI: `kubectl describe vpa <deployment-name> -n <namespace>`. VPA recommendations take weeks to stabilize for new deployments.

## Alerting

### Recommended Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| High error rate | 5xx rate > 5% for 5 minutes | Critical |
| High latency | p99 latency > 5s for 10 minutes | Warning |
| Pod restarts | > 3 restarts in 15 minutes | Warning |
| CPU saturation | CPU usage > 80% for 10 minutes | Warning |
| Memory saturation | Memory usage > 85% for 5 minutes | Critical |
| Health check failing | Readiness probe failing for 3 minutes | Critical |

Configure alerts in Google Cloud Monitoring or Prometheus AlertManager.
