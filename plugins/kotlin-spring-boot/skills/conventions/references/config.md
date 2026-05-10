# Application configuration

`application.yml`, Cloud SQL, HikariCP, Flyway, Redis.

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

Probe paths: `/actuator/health/liveness`, `/actuator/health/readiness`. These match the Entur `common` Helm chart defaults — only change them if you also update Helm values.

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

Reaches the database via the Cloud SQL Auth Proxy sidecar (configured in Helm). Credentials come from Kubernetes secrets that `terraform-google-sql-db` creates.

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

For `database=exposed`, the `exposed-spring-boot-starter` reads `spring.datasource.*` directly — no extra config.

## HikariCP

Default pool size is 10. Total connections = `pods × maximum-pool-size`. Cloud SQL `max_connections` must cover the worst-case HPA pod count plus 3 reserved.

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

Exposed shares the same HikariCP pool — no separate config.

## Flyway

```yaml
spring:
  flyway:
    enabled: true
    locations: classpath:db/migration
    baseline-on-migrate: false
```

Migrations run on startup — if one fails, the service fails to start. For migration file naming and DAO/entity patterns, see `references/database.md`.

## Redis (Memorystore)

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

`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` come from Kubernetes secrets that `terraform-google-memorystore` creates.

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

### Redis rules

- Always set TTLs. Unbounded growth fills Memorystore.
- Eviction policy: `allkeys-lfu` (set in Terraform).
- Treat Redis as a cache, not the primary store. Tolerate failures.
- `SCAN`, never `KEYS *`.
- Values under ~100 KB. Use Cloud Storage for larger objects.
- Pipeline batch operations.
- Messaging goes through Kafka, not Redis Pub/Sub. Pub/Sub has no persistence or delivery guarantees.

