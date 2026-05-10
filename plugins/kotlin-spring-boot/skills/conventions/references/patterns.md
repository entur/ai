# Patterns

## Domain model

`data class` for domain types. Don't write DTOs when `api_approach=contract-first` — OpenAPI Generator produces them.

```kotlin
data class Route(
    val id: Long,
    val name: String,
    val description: String?,
    val status: RouteStatus = RouteStatus.ACTIVE,
)

enum class RouteStatus { ACTIVE, INACTIVE, ARCHIVED }
```

## REST Controllers

### `api_approach=contract-first`

Controllers implement generated interfaces. DTOs come from codegen; never handwrite them.

```kotlin
@RestController
class RouteController(
    private val routeService: RouteService,
    private val routeMapper: RouteMapper,
) : RoutesApi {

    @PreAuthorize("hasAuthority('SCOPE_read:routes')")
    override fun getRoute(id: Long): ResponseEntity<RouteResponse> {
        val route = routeService.findById(id)
        return ResponseEntity.ok(routeMapper.toDto(route))
    }

    @PreAuthorize("hasAuthority('SCOPE_write:routes')")
    override fun createRoute(request: CreateRouteRequest): ResponseEntity<RouteResponse> {
        val route = routeService.create(routeMapper.toDomain(request))
        return ResponseEntity.status(HttpStatus.CREATED).body(routeMapper.toDto(route))
    }
}
```

### `api_approach=traditional`

```kotlin
@RestController
@RequestMapping("/api/v1/routes")
class RouteController(
    private val routeService: RouteService,
) {

    @GetMapping("/{id}")
    @PreAuthorize("hasAuthority('SCOPE_read:routes')")
    fun getRoute(@PathVariable id: Long): ResponseEntity<RouteResponse> =
        ResponseEntity.ok(routeService.findById(id).toResponse())

    @PostMapping
    @PreAuthorize("hasAuthority('SCOPE_write:routes')")
    fun createRoute(@Valid @RequestBody request: CreateRouteRequest): ResponseEntity<RouteResponse> {
        val route = routeService.create(request.toDomain())
        return ResponseEntity.status(HttpStatus.CREATED).body(route.toResponse())
    }
}

data class CreateRouteRequest(
    @field:NotBlank val name: String,
    val description: String?,
)

data class RouteResponse(val id: Long, val name: String, val status: String)

fun Route.toResponse() = RouteResponse(id = id, name = name, status = status.name)
fun CreateRouteRequest.toDomain() = CreateRouteCommand(name = name, description = description)
```

### `spring_stack=webflux`

Controller methods are `suspend`. Spring bridges coroutines to reactive automatically. Apply on top of either `api_approach` above.

```kotlin
@RestController
@RequestMapping("/api/v1/routes")
class RouteController(
    private val routeService: RouteService,
) {
    @GetMapping("/{id}")
    suspend fun getRoute(@PathVariable id: Long): ResponseEntity<RouteResponse> =
        ResponseEntity.ok(routeService.findById(id).toResponse())
}
```

## Service layer

Interface + implementation pair. Services orchestrate DAOs, validate inputs, throw typed domain exceptions.

```kotlin
interface RouteService {
    fun findById(id: Long): Route
    fun findAll(): List<Route>
    fun create(command: CreateRouteCommand): Route
    fun update(id: Long, command: UpdateRouteCommand): Route
}

@Service
class RouteServiceImpl(
    private val routeDao: RouteDao,
) : RouteService {

    override fun findById(id: Long): Route =
        routeDao.findById(id) ?: throw RouteNotFoundException(id)

    override fun create(command: CreateRouteCommand): Route {
        RouteValidator.validate(command)
        return routeDao.insert(command)
    }

    override fun update(id: Long, command: UpdateRouteCommand): Route {
        findById(id)   // throws if not found
        return routeDao.update(id, command) ?: throw RouteNotFoundException(id)
    }
}
```

## Mapper Pattern (contract-first)

```kotlin
@Component
class RouteMapper {

    fun toDomain(dto: CreateRouteRequest) = CreateRouteCommand(
        name = dto.name,
        description = dto.description,
    )

    fun toDto(route: Route) = RouteResponse(
        id = route.id,
        name = route.name,
        description = route.description,
        status = route.status.name,
    )

    fun toDtoList(routes: List<Route>) = routes.map(::toDto)
}
```

## Exception Handling

```kotlin
class RouteNotFoundException(id: Long) :
    RuntimeException("Route not found: $id")

class RouteValidationException(message: String) :
    RuntimeException(message)

@RestControllerAdvice
class GlobalExceptionHandler {

    @ExceptionHandler(RouteNotFoundException::class)
    fun handleNotFound(ex: RouteNotFoundException): ResponseEntity<ErrorResponse> =
        ResponseEntity.status(HttpStatus.NOT_FOUND)
            .body(ErrorResponse(code = "ROUTE_NOT_FOUND", message = ex.message ?: "Not found"))

    @ExceptionHandler(RouteValidationException::class)
    fun handleValidation(ex: RouteValidationException): ResponseEntity<ErrorResponse> =
        ResponseEntity.status(HttpStatus.BAD_REQUEST)
            .body(ErrorResponse(code = "VALIDATION_ERROR", message = ex.message ?: "Bad request"))

    @ExceptionHandler(MethodArgumentNotValidException::class)
    fun handleBeanValidation(ex: MethodArgumentNotValidException): ResponseEntity<ErrorResponse> {
        val message = ex.bindingResult.fieldErrors
            .joinToString("; ") { "${it.field}: ${it.defaultMessage}" }
        return ResponseEntity.status(HttpStatus.BAD_REQUEST)
            .body(ErrorResponse(code = "VALIDATION_ERROR", message = message))
    }
}

data class ErrorResponse(val code: String, val message: String)
```

## Input validation

```kotlin
internal object RouteValidator {
    fun validate(command: CreateRouteCommand) {
        require(command.name.isNotBlank()) { "Route name must not be blank" }
        require(command.name.length <= 100) { "Route name must be 100 characters or fewer" }
    }
}
```

Controller-boundary validation uses bean validation: `@field:NotBlank`, `@field:Size(max = 100)` on the data class, `@Valid` on the controller parameter.

## Kotlin idioms

### Null-safety

```kotlin
fun findRoute(id: Long): Route?           // good
fun findRoute(id: Long): Optional<Route>  // never in Kotlin
```

### Sealed class for results/errors

```kotlin
sealed interface RouteResult {
    data class Success(val route: Route) : RouteResult
    data class NotFound(val id: Long) : RouteResult
    data class ValidationError(val message: String) : RouteResult
}

// At call site
when (val result = routeService.tryCreate(command)) {
    is RouteResult.Success -> ResponseEntity.ok(result.route.toResponse())
    is RouteResult.NotFound -> ResponseEntity.notFound().build()
    is RouteResult.ValidationError -> ResponseEntity.badRequest().body(result.message)
}
```

### Stateless singleton (object)

```kotlin
object RouteConstants {
    const val MAX_NAME_LENGTH = 100
    const val DEFAULT_STATUS = "ACTIVE"
}
```

### Scope functions

```kotlin
// apply: configure an object being constructed
val config = RedisConfig().apply {
    host = "localhost"
    port = 6379
}

// let: transform nullable value
route?.let { routeMapper.toDto(it) }

// also: side effects (logging, debugging)
routeDao.insert(command).also { log.info("Created route: {}", it.id) }

// run: compute a result from an object
val summary = route.run { "$name ($status)" }
```
