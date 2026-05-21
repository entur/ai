# Observability

Micrometer is on the classpath via `spring-boot-starter-actuator`. Metrics are exposed at `/actuator/prometheus` and scraped by the Entur platform. No extra dependency is needed to use `MeterRegistry`.

For request-level HTTP metrics (request counts, latency by status code, error rates), check the cloud-logging library first — it ships built-in integrations via the same BOM. See `references/logging.md` and https://github.com/entur/cloud-logging. The patterns below cover custom business metrics that cloud-logging does not provide.

## Injecting MeterRegistry

```kotlin
@Service
class RouteServiceImpl(
    private val routeDao: RouteDao,
    private val meterRegistry: MeterRegistry,
) : RouteService {
    // ...
}
```

Spring Boot auto-configures and injects `MeterRegistry`. Add it as a constructor parameter on any Spring-managed class.

---

## Metric Types

### Counter

Monotonically increasing count of discrete events:

```kotlin
private val routeCreatedCounter = Counter.builder("route.created")
    .description("Routes created")
    .register(meterRegistry)

override fun create(command: CreateRouteCommand): Route =
    routeDao.insert(command).also { routeCreatedCounter.increment() }
```

### Timer

Measures duration and throughput of operations:

```kotlin
private val findTimer = Timer.builder("route.find.duration")
    .description("Time to fetch a route by ID")
    .register(meterRegistry)

override fun findById(id: Long): Route =
    findTimer.record {
        routeDao.findById(id) ?: throw RouteNotFoundException(id)
    }!!
```

For suspend functions (`spring_stack=webflux`), use `Timer.Sample`:

```kotlin
override suspend fun findById(id: Long): Route {
    val sample = Timer.start(meterRegistry)
    return try {
        routeDao.findById(id) ?: throw RouteNotFoundException(id)
    } finally {
        sample.stop(findTimer)
    }
}
```

### Gauge

Reports a current instantaneous value. Pass a lambda — Micrometer polls it on scrape:

```kotlin
// Register once, e.g. in @PostConstruct or a @Bean initializer
Gauge.builder("route.active.count") { routeDao.countByStatus(RouteStatus.ACTIVE).toDouble() }
    .description("Number of active routes")
    .register(meterRegistry)
```

Never register gauges inside methods that run repeatedly — each call registers a new meter with the same name.

---

## Tag conventions

Tags add dimensions. Keep cardinality low — only a handful of distinct values per tag:

```kotlin
Counter.builder("route.request.count")
    .tag("status", "success")           // ✓ low cardinality
    .tag("operation", "create")         // ✓ low cardinality
    .register(meterRegistry)
```

Never use IDs, user identifiers, timestamps, or free-text as tag values. High-cardinality tags exhaust Prometheus memory.

---

## @Timed annotation

Simpler than manual `Timer.builder(...)` for service methods. Requires a `TimedAspect` bean:

```kotlin
@Configuration
class MetricsConfig {
    @Bean
    fun timedAspect(registry: MeterRegistry) = TimedAspect(registry)
}
```

Then on any Spring-managed method:

```kotlin
@Timed(value = "route.find.duration", description = "Time to fetch a route")
override fun findById(id: Long): Route =
    routeDao.findById(id) ?: throw RouteNotFoundException(id)
```

`@Timed` does not work on suspend functions or final classes without the AllOpen plugin. Use manual `Timer.builder(...)` in those cases.

---

## Test support

Use `SimpleMeterRegistry` in unit tests — it works without a Spring context:

```kotlin
class RouteServiceTest {
    private val meterRegistry = SimpleMeterRegistry()
    private val routeDao = mockk<RouteDao>()
    private val service = RouteServiceImpl(routeDao, meterRegistry)

    @Test
    fun `create route increments counter`() {
        val command = buildCreateRouteCommand()
        every { routeDao.insert(command) } returns buildRoute()

        service.create(command)

        meterRegistry.find("route.created").counter()?.count() shouldBe 1.0
    }
}
```

In `@SpringBootTest` integration tests, inject `MeterRegistry` directly:

```kotlin
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
class MetricsIntegrationTest(
    @Autowired private val meterRegistry: MeterRegistry,
) : IntegrationTestBase() {

    @Test
    @InternalTenant
    fun `creating a route increments counter`() {
        val before = meterRegistry.find("route.created").counter()?.count() ?: 0.0

        routeService.create(buildCreateRouteCommand())

        meterRegistry.find("route.created").counter()?.count() shouldBe before + 1.0
    }
}
```
