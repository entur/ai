# Application Configuration Reference

Spring Boot application configuration, Cloud SQL, HikariCP, Flyway, and Redis for Entur Kotlin services.

## application.yml (defaults)

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

logging:
  level:
    root: INFO
    no.entur.myapp: INFO
```

Health endpoints:
- Liveness: `/actuator/health/liveness`
- Readiness: `/actuator/health/readiness`

These are the defaults in the Entur common Helm chart. Do not change unless you also update Helm values.

## application-local.yml

Overrides for local development. Never commit credentials.

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/myapp_local
    username: myapp
    password: myapp

entur:
  logging:
    style: humanReadablePlain

springdoc:
  swagger-ui:
    enabled: true
```

## Cloud SQL (PostgreSQL)

Connects via Cloud SQL Auth Proxy sidecar (configured in Helm). Credentials injected via Kubernetes secrets created by `terraform-google-sql-db`.

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/${DB_NAME}
    username: ${PG_USER}
    password: ${PG_PASSWORD}
```

For `database=jpa`, add:

```yaml
spring:
  jpa:
    hibernate:
      ddl-auto: validate    # never create/update — use Flyway
    open-in-view: false     # disable OSIV — it causes N+1 queries
```

For `database=exposed`, Exposed uses its own JDBC connection config via the `exposed-spring-boot-starter`. It reads Spring's `spring.datasource.*` automatically.

## HikariCP Connection Pool

Default pool size is 10. Total connections = `num_pods × max_pool_size`. Ensure Cloud SQL `max_connections` handles worst-case HPA pod count (include 3 reserved connections).

```yaml
spring:
  datasource:
    hikari:
      maximum-pool-size: 10
      minimum-idle: 2
      connection-timeout: 30000       # ms — fail fast if pool exhausted
      idle-timeout: 600000            # ms — close idle connections after 10 min
      max-lifetime: 1800000           # ms — recycle connections every 30 min
```

Exposed with `exposed-spring-boot-starter` uses the same HikariCP pool. No separate configuration needed.

## Flyway

```yaml
spring:
  flyway:
    enabled: true
    locations: classpath:db/migration
    baseline-on-migrate: false
```

Migration naming: `V{version}__{description}.sql`

```
src/main/resources/db/migration/
  V1__create_routes_table.sql
  V2__add_route_status_index.sql
  V3__add_stops_table.sql
```

Never modify an applied migration. Always run on startup — if migration fails, the service fails to start.

## Redis (Memorystore)

Dependency:

```kotlin
implementation("org.springframework.boot:spring-boot-starter-data-redis")
```

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

`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` come from Kubernetes secrets created by `terraform-google-memorystore`.

### Spring Cache abstraction

```kotlin
@Configuration
@EnableCaching
class CacheConfig {

    @Bean
    fun cacheManager(connectionFactory: RedisConnectionFactory): RedisCacheManager {
        val defaultConfig = RedisCacheConfiguration.defaultCacheConfig()
            .entryTtl(Duration.ofMinutes(10))
            .serializeValuesWith(
                RedisSerializationContext.SerializationPair.fromSerializer(
                    GenericJackson2JsonRedisSerializer()
                )
            )

        return RedisCacheManager.builder(connectionFactory)
            .cacheDefaults(defaultConfig)
            .withCacheConfiguration("routes", defaultConfig.entryTtl(Duration.ofMinutes(30)))
            .build()
    }
}
```

```kotlin
@Service
class RouteServiceImpl(private val routeDao: RouteDao) : RouteService {

    @Cacheable("routes", key = "#id")
    override fun findById(id: Long): Route = routeDao.findById(id) ?: throw RouteNotFoundException(id)

    @CacheEvict("routes", key = "#id")
    override fun update(id: Long, command: UpdateRouteCommand): Route = routeDao.update(id, command)!!
}
```

### Key naming convention

```
{app}:{domain}:{id}          → route cache: products-api:route:42
{app}:rate:{clientId}        → rate limiting: products-api:rate:partner-xyz
{app}:lock:{resource}        → distributed locks: products-api:lock:route-42
{app}:dedup:{messageId}      → idempotency keys: products-api:dedup:msg-abc123
```

### Redis best practices

- Always set TTLs — unbounded growth exhausts Memorystore memory
- Use `allkeys-lfu` eviction policy (configured in Terraform)
- Handle Redis failures gracefully — it's a cache, not the primary store
- Use `SCAN` (never `KEYS *`) for key iteration in production scripts
- Keep values small — aim for < 100 KB per key; use Cloud Storage for large objects
- Use pipelining for batch operations
- ALWAYS use Kafka for messaging — Redis Pub/Sub lacks persistence and delivery guarantees

## mise.toml (tool version management)

```toml
[tools]
java    = "liberica-<pin>"      # current LTS/stable Liberica build
kotlin  = "<pin>"               # current stable Kotlin

[settings]
experimental = true
```

Pin to current stable versions — read from existing Entur repos or check upstream releases. For full `mise.toml` setup (Terraform, Python, Google Cloud SDK), see https://github.com/entur/ai (`CONVENTIONS.md`) via the `guides` plugin.
