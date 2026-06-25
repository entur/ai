# Java Standards

Java conventions for Entur applications. Read [CONVENTIONS.md](../../CONVENTIONS.md) first for cross-language standards.

- **Target audience**: developers and AI agents building Java services for Entur.
- **Intent**: Java services use the Entur JVM, Spring Boot, logging, health, Redis, and dependency conventions consistently.
- **Scope**: Java runtime/build setup, Spring Boot configuration, coding patterns, testing, Redis, Artifactory, MVC/WebFlux, database connections, and common dependencies. Kotlin-specific rules live in [kotlin.md](kotlin.md).

## Runtime and Build

- **Java version**: 25 or newer
- **Build tool**: Gradle with Kotlin DSL (`build.gradle.kts`)
- **Framework**: Spring Boot 3.x
- **Dependency management**: Gradle version catalogs (`gradle/libs.versions.toml`)
- **JDK distribution**: Eclipse Temurin

## Project Setup

### build.gradle.kts

Use the version catalog (`gradle/libs.versions.toml`) for all versions, and apply Spring plugins via `alias(libs.plugins.x)`. See [kotlin.md](kotlin.md#version-catalog-gradlelibsversionstoml) for a full catalog example.

```kotlin
plugins {
    java
    alias(libs.plugins.spring.boot)
    alias(libs.plugins.spring.dependency.mgmt)
}

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
        vendor = JvmVendorSpec.ADOPTIUM
    }
}

tasks.withType<Test> {
    useJUnitPlatform()
}
```

### Dockerfile

See [docker.md](docker.md) for Dockerfile conventions, base images, and multi-stage builds. The preferred runtime base image for Java/Kotlin is `gcr.io/distroless/java25-debian13:nonroot`; use `eclipse-temurin:25-jre-alpine` only when the container needs a shell, package manager, or runtime debugging tools.

## Logging

Use `entur/cloud-logging` for structured JSON logging on GCP. See [logging.md](logging.md) for general standards.

### Setup

Pin the cloud-logging version in `gradle/libs.versions.toml` (see [kotlin.md](kotlin.md#version-catalog-gradlelibsversionstoml) for the canonical catalog layout), then:

```kotlin
// build.gradle.kts
dependencies {
    implementation(platform(libs.entur.cloud.logging.bom))
    testImplementation(platform(libs.entur.cloud.logging.bom))
    implementation("no.entur.logging.cloud:spring-boot-starter-gcp-web")
    testImplementation("no.entur.logging.cloud:spring-boot-starter-gcp-web-test")
}
```

Remove any existing `logback.xml` or `logback-spring.xml` -- cloud-logging provides its own configuration.

### Usage

Standard SLF4J -- cloud-logging handles JSON formatting, GCP severity mapping, and correlation-id propagation. Configure levels via Spring properties:

```yaml
logging:
  level:
    root: INFO
    no.entur.myapp: INFO
```

### Optional: Request-Response Logging

```kotlin
dependencies {
    implementation("no.entur.logging.cloud:request-response-spring-boot-starter-gcp-web")
    testImplementation("no.entur.logging.cloud:request-response-spring-boot-starter-gcp-web-test")
}
```

```yaml
logbook:
  exclude:
    - /actuator/**
    - /v3/api-docs/**
```

### Optional: On-Demand Logging

Reduces logging costs -- buffers log statements and only flushes full logs for failed requests:

```kotlin
dependencies {
    implementation("no.entur.logging.cloud:on-demand-spring-boot-starter-gcp-web")
}
```

```yaml
entur:
  logging:
    http:
      ondemand:
        enabled: true
        success:
          level: warn
        failure:
          level: info
          http:
            status-code:
              equal-or-higher-than: 400
          logger:
            level: error
```

### Local Development

In test scope, cloud-logging provides human-readable colored output:

```yaml
entur:
  logging:
    style: humanReadablePlain    # humanReadablePlain | humanReadableJson | machineReadableJson
```

### DevOpsLogger (Additional Severity Levels)

cloud-logging includes `DevOpsLogger` (from `DevOpsLoggerFactory`) with additional severity methods: `errorTellMeTomorrow` (ERROR), `errorInterruptMyDinner` (CRITICAL), `errorWakeMeUpRightNow` (ALERT).

## Application Configuration

### application.yml (defaults)

```yaml
server:
  port: 8080

spring:
  application:
    name: ${APPLICATION_NAME:my-application}

management:
  endpoints:
    web:
      exposure:
        include: health, info, prometheus
  endpoint:
    health:
      probes:
        enabled: true
      group:
        liveness:
          include: livenessState
        readiness:
          include: readinessState     # add ',db' for services with a database; including 'db' without a DataSource breaks startup
  metrics:
    tags:
      application: ${APPLICATION_NAME:my-application}
```

### Health Checks

Spring Boot Actuator provides Kubernetes health endpoints:

- Liveness: `/actuator/health/liveness`
- Readiness: `/actuator/health/readiness`

These are defaults in the Entur common Helm chart. Do not change unless you also update Helm values.

## Coding Patterns

### Key Principles

- Use constructor injection (not field injection with `@Autowired`)
- Use Java records for DTOs and value objects
- Use `Optional` for return types only when absence is a normal outcome. Never use `Optional` for fields, parameters, or collection elements (return an empty collection instead).
- Use `@Transactional(readOnly = true)` for read operations
- Validate inputs at the controller boundary with `@Valid`
- Use a mapper layer to convert between entities and DTOs

### Exception Handling

Use `@RestControllerAdvice` with `@ExceptionHandler` methods for centralized error handling. Return a structured error response record (e.g., `ErrorResponse(String code, String message)`) with appropriate HTTP status codes. ALWAYS map domain exceptions to client-friendly responses with safe error messages.

## Testing

### Test Libraries

- JUnit 5 for test framework
- AssertJ for fluent assertions (preferred over Hamcrest)
- Mockito for mocking
- Testcontainers for integration tests with databases and message brokers
- Spring Boot Test for application context tests

## Redis (Memorystore)

Entur uses **Google Memorystore for Redis** when a service needs a managed cache or key-value store. Do not add Redis by default; add it only when the service has a concrete cache, locking, rate-limiting, session, or idempotency need. Infrastructure is provisioned via `terraform-google-memorystore` (see [terraform-modules.md](../platform/terraform-modules.md#memorystore-redis)). Credentials (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`) are injected via Kubernetes secrets.

### When to Use Redis

| Need | Redis fit? | Notes |
|------|------------|-------|
| Shared cache for expensive HTTP responses or DB queries | Good fit | Use cache-aside with bounded TTLs |
| Pod-local, best-effort cache | Usually no | Prefer an in-memory cache such as Caffeine |
| Session storage across pods | Good fit | Only when the service actually owns server-side sessions |
| Rate limiting / counters | Good fit | Atomic `INCR` with TTL |
| Distributed locks | Good fit when needed | Use Redisson or Spring Integration; keep lock TTLs short |
| Idempotency keys / Kafka dedup | Good fit | `SET key value NX EX ttl` pattern |
| Primary data store | **No** | Use PostgreSQL |
| Complex queries / joins | **No** | Use PostgreSQL |
| Large objects (> 1 MB) | **No** | Use Cloud Storage |
| Durable messaging / event streaming | **No** | Use Kafka |

### Dependencies

Add the Redis dependency only in services that use Redis:

```kotlin
dependencies {
    implementation("org.springframework.boot:spring-boot-starter-data-redis")
}
```

### Configuration

```yaml
spring:
  data:
    redis:
      host: ${REDIS_HOST}
      port: ${REDIS_PORT:6379}
      password: ${REDIS_PASSWORD}
      timeout: 2000ms
      connect-timeout: 1000ms
      lettuce:
        pool:
          max-active: 8
          max-idle: 8
          min-idle: 2
          max-wait: 1000ms
```

### Spring Cache Abstraction

For Redis-backed caching, use `@EnableCaching` with `RedisCacheManager` and `@Cacheable`/`@CacheEvict` annotations. Configure per-cache TTLs via `RedisCacheConfiguration`. Serialize with `GenericJackson2JsonRedisSerializer`.

### Direct RedisTemplate Usage

For counters, locks, sets, and hashes, use `StringRedisTemplate` directly. Use `opsForValue().increment()` with TTL for rate limiting, `opsForSet()` for set operations, etc.

### Key Naming Conventions

```text
{app}:{domain}:{id}           -- entity cache
{app}:rate:{clientId}          -- rate limiting
{app}:lock:{resource}          -- distributed locks
{app}:dedup:{messageId}        -- idempotency keys
```

Examples: `products-api:route:ENT:Route:123`, `products-api:rate:partner-xyz`

### Best Practices

- **Always set TTLs** -- unbounded growth exhausts memory
- **Use `allkeys-lfu`** eviction policy (configured in Terraform)
- **Keep values small** -- JSON, aim for < 100 KB per key
- **Use `NX` (set-if-not-exists)** for distributed locks and idempotency
- **Handle failures gracefully** -- Redis is a cache, not primary store. Fall back to DB on failure
- **ALWAYS use `SCAN`** for key iteration in production (it is non-blocking, unlike `KEYS *`)
- **Use pipelining** for batch operations
- **Namespace keys** with app name to avoid collisions
- **Monitor memory** -- alert on `used_memory` vs `maxmemory` (see [observability.md](observability.md))
- **Use Kafka for durable messaging** -- do not use Redis Pub/Sub when persistence, replay, consumer groups, or delivery guarantees matter

### Testing

Use Testcontainers (`GenericContainer` with `redis:7-alpine`) for Redis integration tests. Configure `spring.data.redis.host` and `spring.data.redis.port` via `@DynamicPropertySource`. For unit tests, caching is transparent -- mock the service layer with `@MockBean`.

## Artifactory (JFrog)

Configure Gradle to resolve from Entur's JFrog Artifactory:

```kotlin
repositories {
    val entur_artifactory_user: String? by project
    val entur_artifactory_password: String? by project

    maven {
        name = "Entur JFrog"
        url = URI("https://entur2.jfrog.io/entur2/entur-release-standard/")
        credentials {
            username = entur_artifactory_user ?: System.getenv("ARTIFACTORY_AUTH_USER")
            password = entur_artifactory_password ?: System.getenv("ARTIFACTORY_AUTH_TOKEN")
        }
    }
}
```

Credentials: `$HOME/.gradle/gradle.properties` locally, or `ARTIFACTORY_AUTH_USER`/`ARTIFACTORY_AUTH_TOKEN` org secrets in GitHub Actions.

## Spring MVC vs WebFlux

**Default to Spring MVC** unless you need high-concurrency I/O with an end-to-end non-blocking stack (R2DBC, reactive HTTP clients).

- **WebFlux**: streaming, SSE, WebSockets, many concurrent I/O ops
- **MVC**: blocking stack (JDBC/JPA), CPU-bound, simpler debugging, broader library support
- **Risks of WebFlux**: steep learning curve, harder debugging, testing complexity, blocking code degrades performance

## Connection Pool Sizing

Total DB connections = `number_of_pods * max_pool_size_per_pod`. HikariCP defaults to pool size 10. With 5 pods: `5 * 10 = 50` connections.

Ensure Cloud SQL `max_connections` (minus 3 reserved) handles worst-case HPA pod count. See [Terraform modules](../platform/terraform-modules.md) for Cloud SQL sizing, and [sql.md](sql.md#connection-pooling) for tuning guidance.

## Rate Limiting

Under heavy load, threads can block waiting for DB connections, causing HTTP 503. Options:

- **Spring**: `OncePerRequestFilter` QoS filter returning 503 when rate exceeded
- **Resilience4j**: `@RateLimiter` annotation

Client-side connection timeout must be shorter than server-side timeout.

## Dependencies

### Common Dependencies

| Dependency | Purpose |
|-----------|---------|
| `spring-boot-starter-web` | REST API |
| `spring-boot-starter-actuator` | Health checks, metrics |
| `spring-boot-starter-data-jpa` | Database access (JPA) |
| `spring-boot-starter-validation` | Input validation |
| `no.entur.logging.cloud:spring-boot-starter-gcp-web` | Structured logging (cloud-logging) |
| `micrometer-registry-prometheus` | Prometheus metrics |
| `spring-cloud-gcp-starter` | GCP integration |
| `spring-cloud-gcp-starter-secretmanager` | Secret Manager integration |
| `flyway-core` | Database migrations |
| `postgresql` | PostgreSQL driver |
| `org.entur.data:entur-kafka-spring-starter` | Kafka producer/consumer ([docs](../platform/entur-kafka-starter.md)) |
| `org.entur.openapi:entur-springdoc-starter` | OpenAPI extensions ([docs](api-design.md#entur-springdoc-starter)) |
| `org.entur.metrics:metrics-spring-boot-starter` | Prometheus metrics with Entur defaults ([docs](observability.md#entur-metrics-starter-spring-boot)) |

### Cloud SQL Connectivity

Connects via Cloud SQL proxy sidecar (configured in Helm):

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/${DB_NAME}
    username: ${PGUSER}
    password: ${PGPASSWORD}
```

`PGUSER`, `PGPASSWORD`, `DB_NAME` come from Kubernetes secrets created by `terraform-google-sql-db`.

For query performance, indexing, transactions, and migration patterns see [sql.md](sql.md).
