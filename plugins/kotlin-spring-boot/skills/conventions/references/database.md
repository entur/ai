# Database Reference

Apply the section matching your active `database` configuration. Skip sections for inactive options.

---

## Flyway Migrations (all `database` values except `none`)

Migration files live at `src/main/resources/db/migration/V{version}__{description}.sql`. Naming examples: `V1__create_routes_table.sql`, `V2__add_status_column.sql`. Never modify an applied migration — always add a new versioned file.

For `application.yml` Flyway settings (locations, `baseline-on-migrate`), see `references/config.md`.

---

## `database=exposed`

JetBrains Exposed: typesafe SQL DSL without ORM magic. No `@Entity`, no lazy loading.

### Table definition

```kotlin
object RouteTable : LongIdTable("routes") {
    val name        = varchar("name", 100)
    val description = varchar("description", 500).nullable()
    val status      = enumerationByName<RouteStatus>("status", 20)
    val createdAt   = timestamp("created_at").defaultExpression(CurrentTimestamp)
    val updatedAt   = timestamp("updated_at").defaultExpression(CurrentTimestamp)
}
```

### DAO interface

```kotlin
interface RouteDao {
    fun findById(id: Long): Route?
    fun findAll(): List<Route>
    fun findByStatus(status: RouteStatus): List<Route>
    fun insert(command: CreateRouteCommand): Route
    fun update(id: Long, command: UpdateRouteCommand): Route?
    fun delete(id: Long): Boolean
}
```

### DAO implementation

```kotlin
@Repository
class RouteDaoImpl : RouteDao {

    override fun findById(id: Long): Route? = transaction {
        RouteTable
            .selectAll()
            .where { RouteTable.id eq id }
            .singleOrNull()
            ?.toRoute()
    }

    override fun findAll(): List<Route> = transaction {
        RouteTable.selectAll().map { it.toRoute() }
    }

    override fun findByStatus(status: RouteStatus): List<Route> = transaction {
        RouteTable
            .selectAll()
            .where { RouteTable.status eq status }
            .map { it.toRoute() }
    }

    override fun insert(command: CreateRouteCommand): Route = transaction {
        val id = RouteTable.insertAndGetId {
            it[name]        = command.name
            it[description] = command.description
            it[status]      = RouteStatus.ACTIVE
        }
        findById(id.value)!!
    }

    override fun update(id: Long, command: UpdateRouteCommand): Route? = transaction {
        val updated = RouteTable.update({ RouteTable.id eq id }) {
            command.name?.let { v -> it[name] = v }
            command.description?.let { v -> it[description] = v }
        }
        if (updated == 0) null else findById(id)
    }

    override fun delete(id: Long): Boolean = transaction {
        RouteTable.deleteWhere { RouteTable.id eq id } > 0
    }

    private fun ResultRow.toRoute() = Route(
        id          = this[RouteTable.id].value,
        name        = this[RouteTable.name],
        description = this[RouteTable.description],
        status      = this[RouteTable.status],
    )
}
```

### Joins

Define reusable join functions as extension functions:

```kotlin
object RouteTable : LongIdTable("routes") {
    // ...
    fun joinStops() = join(StopTable, JoinType.LEFT, id, StopTable.routeId)
}

// Usage
RouteTable.joinStops()
    .selectAll()
    .where { RouteTable.status eq RouteStatus.ACTIVE }
    .map { row -> RouteWithStops(row.toRoute(), row.toStop()) }
```

### Transaction boundaries

Keep `transaction {}` in the DAO layer. Never call `transaction {}` from services — that couples services to the persistence layer.

---

## `database=spring-data-jdbc`

Spring Data JDBC: simpler than JPA. No lazy loading, no entity lifecycle, no bytecode manipulation. Plain Kotlin data classes work without modification.

### Entity

```kotlin
@Table("routes")
data class RouteEntity(
    @Id val id: Long = 0,
    val name: String,
    val description: String? = null,
    val status: String = RouteStatus.ACTIVE.name,
    val createdAt: Instant = Instant.now(),
)
```

- No `@Entity`, no `@GeneratedValue` — Spring Data JDBC uses the database-generated ID automatically when `id = 0`
- No lazy loading to avoid — all associations are eager by default

### Repository

```kotlin
interface RouteRepository : CrudRepository<RouteEntity, Long> {

    fun findByStatus(status: String): List<RouteEntity>

    @Query("SELECT * FROM routes WHERE name ILIKE :pattern ORDER BY name")
    fun searchByName(pattern: String): List<RouteEntity>

    @Query("SELECT COUNT(*) FROM routes WHERE status = :status")
    fun countByStatus(status: String): Long
}
```

### DAO wrapping the repository

```kotlin
interface RouteDao {
    fun findById(id: Long): Route?
    fun findAll(): List<Route>
    fun insert(command: CreateRouteCommand): Route
}

@Repository
class RouteDaoImpl(
    private val repository: RouteRepository,
) : RouteDao {

    override fun findById(id: Long): Route? =
        repository.findByIdOrNull(id)?.toDomain()

    override fun findAll(): List<Route> =
        repository.findAll().map { it.toDomain() }

    override fun insert(command: CreateRouteCommand): Route =
        repository.save(
            RouteEntity(name = command.name, description = command.description)
        ).toDomain()

    private fun RouteEntity.toDomain() = Route(
        id          = id,
        name        = name,
        description = description,
        status      = RouteStatus.valueOf(status),
    )
}
```

### Spring Data JDBC notes

- Aggregates are loaded completely — no lazy loading
- One-to-many relationships: reference child table's column pointing to parent `@Id`
- For complex queries, prefer `@Query` with raw SQL over Spring Data JDBC's limited query derivation
- Pagination: extend `PagingAndSortingRepository<Entity, Long>` instead of `CrudRepository`
- Use `@Transactional` on service methods that call multiple repository operations

---

## `database=jpa`

Spring Data JPA. Use when integrating with an existing JPA schema or when the team is already fluent with Hibernate.

### Known Kotlin/JPA friction points

- Kotlin data classes with `val` properties conflict with Hibernate's need to mutate fields — use `var` in `@Entity` classes, or apply the `kotlin-jpa` plugin (generates no-arg constructors) and accept `var`
- Lazy loading requires proxying — use `open` classes or the `kotlin-allopen` plugin. The `kotlin-spring` plugin already does this for `@Entity`
- Avoid `copy()` on entities managed by the persistence context — it creates a detached instance

### Entity

```kotlin
@Entity
@Table(name = "routes")
class RouteEntity(
    @Id @GeneratedValue(strategy = GenerationType.IDENTITY)
    var id: Long = 0,

    var name: String,

    var description: String? = null,

    @Enumerated(EnumType.STRING)
    var status: RouteStatus = RouteStatus.ACTIVE,

    @CreationTimestamp
    var createdAt: Instant = Instant.now(),
)
```

### Repository

```kotlin
interface RouteRepository : JpaRepository<RouteEntity, Long> {

    fun findByStatus(status: RouteStatus): List<RouteEntity>

    @Query("SELECT r FROM RouteEntity r WHERE r.name LIKE :pattern")
    fun searchByName(pattern: String): List<RouteEntity>
}
```

### Service with `@Transactional`

```kotlin
@Service
@Transactional(readOnly = true)         // default all methods to read-only
class RouteServiceImpl(
    private val repository: RouteRepository,
) : RouteService {

    override fun findById(id: Long): Route =
        repository.findByIdOrNull(id)?.toDomain() ?: throw RouteNotFoundException(id)

    @Transactional                       // override for write methods
    override fun create(command: CreateRouteCommand): Route {
        val entity = RouteEntity(name = command.name, description = command.description)
        return repository.save(entity).toDomain()
    }
}

private fun RouteEntity.toDomain() = Route(id = id, name = name, description = description, status = status)
```

### JPA notes

- Avoid `FetchType.LAZY` on `@OneToMany` unless you can guarantee the session is open — prefer explicit JPQL joins
- Use `@EntityGraph` for controlling fetch plan without N+1 queries
- Use `@BatchSize` on collections to reduce N+1 queries
- Paginated queries: use `Pageable` parameter in repository methods
- `findByIdOrNull` (Spring Data Kotlin extension) is safer than `findById` which returns `Optional`
