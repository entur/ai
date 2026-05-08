# Logging Reference

Structured logging for Entur Kotlin Spring Boot services. Source of truth: https://github.com/entur/cloud-logging — fetch the README for the current artifact list, BOM coordinates, and supported web stack variants (web vs webflux, with/without request-response, on-demand). Do not invent artifact names from this file.

## Setup

Pin the cloud-logging version in `gradle/libs.versions.toml` (or Maven `<dependencyManagement>`). Apply the BOM, then the GCP web starter that matches the project's Spring stack — exact artifact names per repo README.

```kotlin
// build.gradle.kts — structure only; replace <starter> with names from cloud-logging README
dependencies {
    implementation(platform("no.entur.logging.cloud:bom:${libs.versions.enturCloudLogging.get()}"))
    implementation("no.entur.logging.cloud:<gcp-web-starter>")
    testImplementation(platform("no.entur.logging.cloud:bom:${libs.versions.enturCloudLogging.get()}"))
    testImplementation("no.entur.logging.cloud:<gcp-web-test-starter>")
}
```

Remove any existing `logback.xml` or `logback-spring.xml` — cloud-logging provides its own configuration.

## Usage

Standard SLF4J — cloud-logging handles JSON formatting, GCP severity mapping, and correlation-ID propagation automatically.

```kotlin
import org.slf4j.LoggerFactory

private val log = LoggerFactory.getLogger(RouteService::class.java)

@Service
class RouteServiceImpl(private val routeDao: RouteDao) : RouteService {

    override fun findById(id: Long): Route {
        log.info("Fetching route id={}", id)
        return routeDao.findById(id)
            ?.also { log.debug("Found route id={} name={}", it.id, it.name) }
            ?: run {
                log.warn("Route not found id={}", id)
                throw RouteNotFoundException(id)
            }
    }
}
```

Configure log levels in `application.yml`:

```yaml
logging:
  level:
    root: INFO
    no.entur.myapp: INFO
    no.entur.myapp.dao: DEBUG   # fine-grained per package
```

## Optional: Request-Response Logging

Logs full HTTP request and response bodies for debugging. Adds overhead — use selectively. Add the request-response starter from cloud-logging (exact artifact name in repo README).

```kotlin
dependencies {
    implementation("no.entur.logging.cloud:<request-response-starter>")
    testImplementation("no.entur.logging.cloud:<request-response-test-starter>")
}
```

```yaml
logbook:
  exclude:
    - /actuator/**          # exclude health/metrics endpoints
    - /v3/api-docs/**       # exclude OpenAPI docs
```

## Optional: On-Demand Logging

Buffers log statements per request and only flushes the full log for failed requests. Reduces GCP logging costs significantly for high-traffic services. Add the on-demand starter from cloud-logging (exact artifact name in repo README).

```kotlin
dependencies {
    implementation("no.entur.logging.cloud:<on-demand-starter>")
}
```

```yaml
entur:
  logging:
    http:
      ondemand:
        enabled: true
        success:
          level: warn        # suppress logs for successful requests
        failure:
          level: info        # flush full log for requests with status >= 400
          http:
            status-code:
              equal-or-higher-than: 400
          logger:
            level: error     # flush full log for ERROR-level log statements
```

## Local Development

Human-readable colored output in test scope:

```yaml
# src/test/resources/application.yml or application-local.yml
entur:
  logging:
    style: humanReadablePlain    # humanReadablePlain | humanReadableJson | machineReadableJson
```

## DevOpsLogger — additional severity levels

cloud-logging includes `DevOpsLogger` with alerting-oriented severity methods:

```kotlin
import no.entur.logging.cloud.api.DevOpsLoggerFactory

private val devOpsLog = DevOpsLoggerFactory.getLogger(RouteService::class.java)

// Use these for operations alerts — maps to GCP severity levels
devOpsLog.errorTellMeTomorrow("Non-critical data sync failed: {}", error)    // ERROR
devOpsLog.errorInterruptMyDinner("Payment processing unavailable: {}", error) // CRITICAL
devOpsLog.errorWakeMeUpRightNow("Database connection pool exhausted")         // ALERT
```

Use standard `log.error()` for application errors that don't need paging. Use `DevOpsLogger` only for production incidents that require human intervention.
