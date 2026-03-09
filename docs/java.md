# Java Standards

Java conventions for Entur applications. Read [CONVENTIONS.md](../CONVENTIONS.md) first for cross-language standards.

## Runtime and Build

- **Java version**: 25 or newer
- **Build tool**: Gradle with Kotlin DSL (`build.gradle.kts`)
- **Framework**: Spring Boot 3.x
- **Dependency management**: Gradle version catalogs (`gradle/libs.versions.toml`)
- **JDK distribution**: Liberica JDK (preferred) or Eclipse Temurin

## Project Setup

### build.gradle.kts

```kotlin
plugins {
    java
    id("org.springframework.boot") version libs.versions.springBoot
    id("io.spring.dependency-management") version libs.versions.springDependencyManagement
}

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
}

tasks.withType<Test> {
    useJUnitPlatform()
}
```

### Dockerfile

See [docker.md](docker.md) for Dockerfile conventions, base images, and multi-stage builds. Simple example:

```dockerfile
FROM eclipse-temurin:25-jre-alpine
WORKDIR /app
COPY build/libs/*.jar app.jar
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

## Logging

Use [entur/cloud-logging](https://github.com/entur/cloud-logging) for structured JSON logging on GCP. See [logging.md](logging.md) for general standards.

### Setup

```kotlin
// build.gradle.kts
val cloudLoggingVersion = "x.y.z"  // check Maven Central for latest

dependencies {
    implementation(platform("no.entur.logging.cloud:bom:$cloudLoggingVersion"))
    testImplementation(platform("no.entur.logging.cloud:bom:$cloudLoggingVersion"))
    implementation("no.entur.logging.cloud:spring-boot-starter-gcp-web")
    testImplementation("no.entur.logging.cloud:spring-boot-starter-gcp-web-test")
}
```

Remove any existing `logback.xml` or `logback-spring.xml` -- cloud-logging provides its own configuration.

### Usage

Standard SLF4J -- cloud-logging handles JSON formatting, GCP severity mapping, and correlation-id propagation:

```java
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

private static final Logger LOG = LoggerFactory.getLogger(RouteService.class);

LOG.info("Route found for id {}", routeId);
LOG.warn("Retry attempt {} for route {}", attempt, routeId);
LOG.error("Failed to fetch route {}", routeId, exception);
```

Configure levels via Spring properties:

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

```java
import no.entur.logging.cloud.api.DevOpsLogger;
import no.entur.logging.cloud.api.DevOpsLoggerFactory;

private static final DevOpsLogger LOGGER = DevOpsLoggerFactory.getLogger(MyService.class);

LOGGER.errorTellMeTomorrow("Non-urgent error");        // ERROR level
LOGGER.errorInterruptMyDinner("Critical error");        // CRITICAL level
LOGGER.errorWakeMeUpRightNow("System down");            // ALERT level
```

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
          include: readinessState, db
  metrics:
    tags:
      application: ${spring.application.name}
```

### Health Checks

Spring Boot Actuator provides Kubernetes health endpoints:

- Liveness: `/actuator/health/liveness`
- Readiness: `/actuator/health/readiness`

These are defaults in the Entur common Helm chart. Do not change unless you also update Helm values.

## Coding Patterns

### REST Controllers

```java
@RestController
@RequestMapping("/api/v1/routes")
public class RouteController {

    private final RouteService routeService;

    public RouteController(RouteService routeService) {
        this.routeService = routeService;
    }

    @GetMapping("/{id}")
    public ResponseEntity<RouteResponse> getRoute(@PathVariable String id) {
        return routeService.findById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<RouteResponse> createRoute(
            @Valid @RequestBody CreateRouteRequest request) {
        RouteResponse created = routeService.create(request);
        URI location = URI.create("/api/v1/routes/" + created.id());
        return ResponseEntity.created(location).body(created);
    }
}
```

### Service Layer

```java
@Service
public class RouteService {

    private final RouteRepository routeRepository;

    public RouteService(RouteRepository routeRepository) {
        this.routeRepository = routeRepository;
    }

    @Transactional(readOnly = true)
    public Optional<RouteResponse> findById(String id) {
        return routeRepository.findById(id)
                .map(RouteMapper::toResponse);
    }

    @Transactional
    public RouteResponse create(CreateRouteRequest request) {
        Route route = RouteMapper.toEntity(request);
        Route saved = routeRepository.save(route);
        return RouteMapper.toResponse(saved);
    }
}
```

### Key Principles

- Use constructor injection (not field injection with `@Autowired`)
- Use Java records for DTOs and value objects
- Use `Optional` for return types that may be absent -- never return null
- Use `@Transactional(readOnly = true)` for read operations
- Validate inputs at the controller boundary with `@Valid`
- Use a mapper layer to convert between entities and DTOs

### Exception Handling

```java
@RestControllerAdvice
public class GlobalExceptionHandler {

    private static final Logger LOG = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleNotFound(ResourceNotFoundException ex) {
        LOG.warn("Resource not found: {}", ex.getMessage());
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(new ErrorResponse("NOT_FOUND", ex.getMessage()));
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ErrorResponse> handleValidation(MethodArgumentNotValidException ex) {
        String message = ex.getBindingResult().getFieldErrors().stream()
                .map(e -> e.getField() + ": " + e.getDefaultMessage())
                .collect(Collectors.joining(", "));
        return ResponseEntity.badRequest()
                .body(new ErrorResponse("VALIDATION_ERROR", message));
    }
}

public record ErrorResponse(String code, String message) {}
```

## Testing

### Unit Tests

```java
@ExtendWith(MockitoExtension.class)
class RouteServiceTest {

    @Mock
    private RouteRepository routeRepository;

    @InjectMocks
    private RouteService routeService;

    @Test
    @DisplayName("findById returns route when it exists")
    void findById_existingRoute_returnsRoute() {
        Route route = TestFixtures.aRoute().build();
        when(routeRepository.findById("route-1")).thenReturn(Optional.of(route));

        Optional<RouteResponse> result = routeService.findById("route-1");

        assertThat(result).isPresent();
        assertThat(result.get().id()).isEqualTo("route-1");
    }
}
```

### Integration Tests

```java
@SpringBootTest
@Testcontainers
class RouteRepositoryIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine");

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
    }

    @Autowired
    private RouteRepository routeRepository;

    @Test
    void savesAndRetrievesRoute() {
        Route route = new Route("route-1", "Oslo S", "Bergen");
        routeRepository.save(route);

        Optional<Route> found = routeRepository.findById("route-1");
        assertThat(found).isPresent();
    }
}
```

### Test Libraries

- JUnit 5 for test framework
- AssertJ for fluent assertions (preferred over Hamcrest)
- Mockito for mocking
- Testcontainers for integration tests with databases and message brokers
- Spring Boot Test for application context tests

## Redis (Memorystore)

Entur uses **Google Memorystore for Redis** as a managed key-value store. Infrastructure via `terraform-google-memorystore` (see [terraform/modules.md](terraform/modules.md#memorystore-redis)). Credentials (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`) injected via Kubernetes secrets.

### When to Use Redis

| Use Case | Redis? | Notes |
|----------|--------|-------|
| Caching (HTTP responses, DB queries) | Yes | Primary use case |
| Session storage | Yes | Shared sessions across pods |
| Rate limiting / counters | Yes | Atomic `INCR` with TTL |
| Distributed locks | Yes | Use Redisson or Spring Integration |
| Idempotency keys (Kafka dedup) | Yes | `SET key value NX EX ttl` pattern |
| Primary data store | **No** | Use PostgreSQL |
| Complex queries / joins | **No** | Use PostgreSQL |
| Large objects (> 1 MB) | **No** | Use Cloud Storage |

### Dependencies

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

Use `@Cacheable` annotations with Redis as the backing store:

```java
@Configuration
@EnableCaching
public class CacheConfig {

    @Bean
    public RedisCacheConfiguration cacheConfiguration() {
        return RedisCacheConfiguration.defaultCacheConfig()
            .entryTtl(Duration.ofMinutes(10))
            .disableCachingNullValues()
            .serializeValuesWith(
                RedisSerializationContext.SerializationPair
                    .fromSerializer(new GenericJackson2JsonRedisSerializer())
            );
    }

    @Bean
    public RedisCacheManager cacheManager(RedisConnectionFactory factory) {
        return RedisCacheManager.builder(factory)
            .cacheDefaults(cacheConfiguration())
            .withCacheConfiguration("routes",
                cacheConfiguration().entryTtl(Duration.ofHours(1)))
            .withCacheConfiguration("stops",
                cacheConfiguration().entryTtl(Duration.ofMinutes(30)))
            .build();
    }
}
```

```java
@Service
public class RouteService {

    @Cacheable(value = "routes", key = "#id")
    public Route findById(String id) {
        return routeRepository.findById(id).orElseThrow();
    }

    @CacheEvict(value = "routes", key = "#id")
    public void update(String id, UpdateRouteRequest request) {
        // Cache entry is evicted after update
    }

    @CacheEvict(value = "routes", allEntries = true)
    public void refreshAll() {
        // Evict all entries in the "routes" cache
    }
}
```

### Direct RedisTemplate Usage

For counters, locks, sets, hashes:

```java
@Component
public class RateLimiter {

    private final StringRedisTemplate redis;

    public RateLimiter(StringRedisTemplate redis) {
        this.redis = redis;
    }

    public boolean isAllowed(String clientId, int maxRequests, Duration window) {
        String key = "rate:" + clientId;
        Long count = redis.opsForValue().increment(key);
        if (count == 1) {
            redis.expire(key, window);
        }
        return count <= maxRequests;
    }
}
```

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
- **Avoid `KEYS *`** in production -- blocks Redis. Use `SCAN` instead
- **Use pipelining** for batch operations
- **Namespace keys** with app name to avoid collisions
- **Monitor memory** -- alert on `used_memory` vs `maxmemory` (see [observability.md](observability.md))
- **Do not use Redis as a message queue** -- use Kafka. Redis Pub/Sub has no persistence

### Testing

Use Testcontainers for integration tests:

```java
@SpringBootTest
@Testcontainers
class RedisCacheIntegrationTest {

    @Container
    static GenericContainer<?> redis = new GenericContainer<>("redis:7-alpine")
        .withExposedPorts(6379);

    @DynamicPropertySource
    static void configureRedis(DynamicPropertyRegistry registry) {
        registry.add("spring.data.redis.host", redis::getHost);
        registry.add("spring.data.redis.port", () -> redis.getMappedPort(6379));
        registry.add("spring.data.redis.password", () -> "");
    }
}
```

For unit tests, mock the cache or use `@MockBean` on `RedisTemplate`:

```java
@WebMvcTest(RouteController.class)
class RouteControllerTest {

    @MockBean
    private RouteService routeService;  // caching is transparent

    // Test controller behavior -- caching is an implementation detail
}
```

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

Ensure Cloud SQL `max_connections` (minus 3 reserved) handles worst-case HPA pod count. See [Terraform modules](terraform/modules.md) for Cloud SQL sizing.

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
| `org.entur.data:entur-kafka-spring-starter` | Kafka producer/consumer ([docs](kafka.md)) |
| `org.entur.openapi:entur-springdoc-starter` | OpenAPI extensions ([docs](api-design.md#entur-springdoc-starter)) |
| `org.entur.metrics:metrics-spring-boot-starter` | Prometheus metrics with Entur defaults ([docs](observability.md#entur-metrics-starter-spring-boot)) |

### Cloud SQL Connectivity

Connects via Cloud SQL proxy sidecar (configured in Helm):

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/${DB_NAME}
    username: ${PG_USER}
    password: ${PG_PASSWORD}
```

`PG_USER`, `PG_PASSWORD`, `DB_NAME` come from Kubernetes secrets created by `terraform-google-sql-db`.
