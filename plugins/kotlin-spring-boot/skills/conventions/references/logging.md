# Logging

Source: https://github.com/entur/cloud-logging. Fetch the README for current artifact names, BOM coordinates, and web stack variants (web/webflux, with/without request-response, on-demand). Don't invent artifact names from this file.

## Setup

Pin the cloud-logging version in `gradle/libs.versions.toml` (or Maven `<dependencyManagement>`). Apply the BOM, then the GCP web starter matching the Spring stack.

```kotlin
// build.gradle.kts — structure only; replace <starter> with names from cloud-logging README
dependencies {
    implementation(platform("no.entur.logging.cloud:bom:${libs.versions.enturCloudLogging.get()}"))
    implementation("no.entur.logging.cloud:<gcp-web-starter>")
    testImplementation(platform("no.entur.logging.cloud:bom:${libs.versions.enturCloudLogging.get()}"))
    testImplementation("no.entur.logging.cloud:<gcp-web-test-starter>")
}
```

Delete any existing `logback.xml` or `logback-spring.xml` — cloud-logging ships its own.

## Usage

Standard SLF4J. Cloud-logging handles JSON formatting, GCP severity, and correlation-ID propagation.

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

Logs full HTTP request/response bodies. Has overhead — use selectively. Add the request-response starter from cloud-logging.

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

Buffers per request, flushes only on failure. Cuts GCP logging cost for high-traffic services. Add the on-demand starter from cloud-logging.

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

## DevOpsLogger

`DevOpsLogger` from cloud-logging maps method names to GCP severity for alerts:

```kotlin
import no.entur.logging.cloud.api.DevOpsLoggerFactory

private val devOpsLog = DevOpsLoggerFactory.getLogger(RouteService::class.java)

// Use these for operations alerts — maps to GCP severity levels
devOpsLog.errorTellMeTomorrow("Non-critical data sync failed: {}", error)    // ERROR
devOpsLog.errorInterruptMyDinner("Payment processing unavailable: {}", error) // CRITICAL
devOpsLog.errorWakeMeUpRightNow("Database connection pool exhausted")         // ALERT
```

Use plain `log.error()` for non-paging errors. `DevOpsLogger` is for production incidents needing human intervention.
