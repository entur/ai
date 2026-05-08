# Testing Reference

Kotlin Spring Boot testing conventions for Entur projects.

## Test stack

| Library | Purpose | Always? |
|---|---|---|
| JUnit 5 | Test framework | Yes |
| Spring Boot Test | `@SpringBootTest`, `@WebMvcTest`, `@DataJdbcTest` | Yes |
| TestContainers | Dockerized PostgreSQL/Redis in integration tests | When database/Redis |
| Entur Auth Test | `TenantJsonWebToken` for simulating auth | When service has auth |
| SpringMockK (`com.ninja-squad:springmockk`) | `@MockkBean` — MockK mocks in Spring context | When `test_mocking=mockk` |
| Mockito-Kotlin (`org.mockito.kotlin:mockito-kotlin`) | `@MockBean` — Mockito mocks in Spring context | When `test_mocking=mockito-kotlin` |
| Kotest assertions | `shouldBe`, `shouldThrow`, etc. | When `test_assertions=kotest` |
| AssertJ | `assertThat(...).isEqualTo(...)` | When `test_assertions=assertj` |

---

## Assertion syntax by configuration

### `test_assertions=kotest`

```kotlin
// Equality
result shouldBe expected
result shouldNotBe null

// Collections
list shouldHaveSize 3
list shouldContain item
list shouldBeEmpty()

// Strings
string shouldStartWith "prefix"
string shouldContain "substring"
string shouldHaveLength 10

// Exceptions
shouldThrow<RouteNotFoundException> { service.findById(99L) }
shouldThrow<RouteNotFoundException> { service.findById(99L) }
    .also { it.message shouldContain "99" }

// Nullable
result.shouldNotBeNull()
result shouldBe null
```

### `test_assertions=assertj`

```kotlin
// Equality
assertThat(result).isEqualTo(expected)
assertThat(result).isNotNull()
assertThat(result).isNull()

// Collections
assertThat(list).hasSize(3)
assertThat(list).contains(item)
assertThat(list).isEmpty()

// Strings
assertThat(string).startsWith("prefix")
assertThat(string).contains("substring")
assertThat(string).hasSize(10)

// Exceptions
assertThatThrownBy { service.findById(99L) }
    .isInstanceOf(RouteNotFoundException::class.java)
    .hasMessageContaining("99")
```

---

## Mocking syntax by configuration

### `test_mocking=mockk`

```kotlin
// In @WebMvcTest / @SpringBootTest
@MockkBean
lateinit var routeService: RouteService

// Stubbing — regular functions
every { routeService.findById(1L) } returns route
every { routeService.findById(any()) } throws RouteNotFoundException(0L)

// Stubbing — suspend functions (webflux/coroutine)
coEvery { routeService.findByIdSuspend(1L) } returns route

// Verification
verify(exactly = 1) { routeService.findById(1L) }
verify(exactly = 0) { routeService.delete(any()) }
coVerify { routeService.findByIdSuspend(1L) }

// Capture arguments
val slot = slot<Long>()
every { routeService.findById(capture(slot)) } returns route
// ... call code ...
slot.captured shouldBe 1L

// In pure unit tests (no Spring context)
val routeDao = mockk<RouteDao>()
```

### `test_mocking=mockito-kotlin`

```kotlin
// In @WebMvcTest / @SpringBootTest
@MockBean
lateinit var routeService: RouteService

// Stubbing
whenever(routeService.findById(1L)).thenReturn(route)
whenever(routeService.findById(any())).thenThrow(RouteNotFoundException(0L))

// Verification
verify(routeService).findById(1L)
verify(routeService, times(1)).findById(1L)
verify(routeService, never()).delete(any())

// Capture arguments
val captor = argumentCaptor<Long>()
verify(routeService).findById(captor.capture())
assertThat(captor.firstValue).isEqualTo(1L)

// In pure unit tests
val routeDao = mock<RouteDao>()
```

---

## Unit Tests

```kotlin
class RouteServiceTest {

    // mockk style:
    private val routeDao = mockk<RouteDao>()

    // mockito-kotlin style:
    // private val routeDao = mock<RouteDao>()

    private val service = RouteServiceImpl(routeDao)

    @Test
    fun `findById returns route when found`() {
        val route = buildRoute(id = 1L)
        every { routeDao.findById(1L) } returns route

        val result = service.findById(1L)

        result shouldBe route
    }

    @Test
    fun `findById throws RouteNotFoundException when not found`() {
        every { routeDao.findById(99L) } returns null

        shouldThrow<RouteNotFoundException> { service.findById(99L) }
    }
}
```

Use backtick function names for readable sentences. Each test covers one behaviour. Use the Arrange-Act-Assert pattern.

---

## Controller Tests (`@WebMvcTest`)

```kotlin
@WebMvcTest(RouteController::class)
@Import(SecurityConfig::class)
@ExtendWith(TenantJsonWebToken::class)
class RouteControllerTest(
    @Autowired private val mockMvc: MockMvc,
) {

    @MockkBean                        // or @MockBean for mockito-kotlin
    lateinit var routeService: RouteService

    @Test
    @InternalTenant
    fun `GET route returns 200 with route data`() {
        val route = buildRoute(id = 1L, name = "Oslo - Bergen")
        every { routeService.findById(1L) } returns route

        mockMvc.get("/api/v1/routes/1")
            .andExpect {
                status { isOk() }
                jsonPath("$.id") { value(1L) }
                jsonPath("$.name") { value("Oslo - Bergen") }
            }
    }

    @Test
    fun `GET route returns 401 when unauthenticated`() {
        mockMvc.get("/api/v1/routes/1")
            .andExpect { status { isUnauthorized() } }
    }

    @Test
    @InternalTenant
    fun `GET route returns 404 when not found`() {
        every { routeService.findById(99L) } throws RouteNotFoundException(99L)

        mockMvc.get("/api/v1/routes/99")
            .andExpect { status { isNotFound() } }
    }
}
```

---

## Integration Tests (TestContainers)

### Shared container configuration

```kotlin
@TestConfiguration(proxyBeanMethods = false)
class TestContainersConfig {
    @Bean
    @ServiceConnection
    fun postgresContainer() = PostgreSQLContainer("postgres:16-alpine")
}
```

### Base test class pattern

```kotlin
@SpringBootTest
@Import(TestContainersConfig::class)
@ExtendWith(TenantJsonWebToken::class)
@Sql(
    scripts = ["/test-data/cleanup.sql"],
    executionPhase = Sql.ExecutionPhase.BEFORE_TEST_METHOD,
)
abstract class IntegrationTestBase
```

### Integration test

```kotlin
class RouteIntegrationTest(
    @Autowired private val routeService: RouteService,
) : IntegrationTestBase() {

    @Test
    @InternalTenant
    fun `creating a route persists it and returns it with a generated ID`() {
        val command = CreateRouteCommand(name = "Test Route", description = null)

        val created = routeService.create(command)

        created.id shouldBeGreaterThan 0L
        created.name shouldBe "Test Route"
        created.status shouldBe RouteStatus.ACTIVE
    }
}
```

Place cleanup SQL at `src/test/resources/test-data/cleanup.sql`:

```sql
DELETE FROM routes;
```

---

## Slice Tests for Database Layer

### `database=spring-data-jdbc`

```kotlin
@DataJdbcTest
@Import(TestContainersConfig::class)
class RouteRepositoryTest(
    @Autowired private val repository: RouteRepository,
) {
    @Test
    fun `findByStatus returns matching routes`() {
        repository.save(RouteEntity(name = "Active Route"))
        repository.save(RouteEntity(name = "Inactive Route", status = "INACTIVE"))

        val active = repository.findByStatus("ACTIVE")

        active shouldHaveSize 1
        active.first().name shouldBe "Active Route"
    }
}
```

### `database=jpa`

```kotlin
@DataJpaTest
@Import(TestContainersConfig::class)
class RouteRepositoryTest(
    @Autowired private val repository: RouteRepository,
) {
    // same pattern as above
}
```

---

## Test Data Builders

```kotlin
fun buildRoute(
    id: Long = 1L,
    name: String = "Test Route",
    description: String? = null,
    status: RouteStatus = RouteStatus.ACTIVE,
) = Route(id = id, name = name, description = description, status = status)

fun buildCreateRouteCommand(
    name: String = "Test Route",
    description: String? = null,
) = CreateRouteCommand(name = name, description = description)
```

Builders provide sensible defaults so tests only declare the values they care about.

---

## Entur Auth Test

For Entur OIDC auth (resource server, JWT, scopes, tenant tokens), source of truth: https://github.com/entur/oidc-auth-resource-server. Fetch the README for current test artifact names and the supported tenant/scope annotations.

Pattern (annotation names per repo README):

```kotlin
@ExtendWith(TenantJsonWebToken::class)
class RouteControllerTest {

    @Test
    @InternalTenant                        // simulates an internal Entur service token
    fun `endpoint returns 200 for internal tenant`() { ... }

    @Test
    @TravellerTenant                       // simulates an end-user token
    fun `endpoint returns 200 for traveller tenant`() { ... }

    @Test
    fun `endpoint returns 401 when unauthenticated`() { ... }
}
```

For broader authorization patterns (scopes, role mapping, configuration), see the `entur/ai` repo guides via the `guides` plugin or directly at https://github.com/entur/ai.
