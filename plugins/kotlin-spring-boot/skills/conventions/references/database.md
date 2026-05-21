# Database

Apply the section matching the active `database` axis. Skip the rest.

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

Reusable joins as extension functions on the table:

```kotlin
object RouteTable : LongIdTable("routes") {
    // ...
    fun joinStops() = join(StopTable, JoinType.LEFT, id, StopTable.routeId)
}

RouteTable.joinStops()
    .selectAll()
    .where { RouteTable.status eq RouteStatus.ACTIVE }
    .map { row -> RouteWithStops(row.toRoute(), row.toStop()) }
```

### Transaction boundaries

`transaction {}` belongs in the DAO. Never call it from services — that couples services to persistence.

---

## `database=spring-data-jdbc`

Spring Data JDBC: simpler than JPA. No lazy loading, no entity lifecycle, no bytecode manipulation. Plain Kotlin data classes work without modification.

### Entity

```kotlin
@Table("routes")
data class RouteEntity(
    @Id val id: Long? = null,           // null = unsaved; set by database on insert
    val name: String,
    val description: String? = null,
    val status: String = RouteStatus.ACTIVE.name,
    val createdAt: Instant = Instant.now(),
)
```

- No `@Entity`, no `@GeneratedValue` — Spring Data JDBC detects a `null` `@Id` field as a new entity and uses the database-generated value after insert
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
        id          = requireNotNull(id) { "Entity missing persisted ID" },
        name        = name,
        description = description,
        status      = RouteStatus.valueOf(status),
    )
}
```

### Notes

- Aggregates load completely — no lazy loading
- One-to-many: child table column references parent `@Id`
- Complex queries: `@Query` with raw SQL beats Spring Data JDBC's query derivation
- Pagination: `PagingAndSortingRepository<Entity, Long>` instead of `CrudRepository`
- Multi-step service methods: `@Transactional`

---

## `database=jpa`

Spring Data JPA. Use when integrating with an existing JPA schema or when the team is already fluent with Hibernate.

### Kotlin/JPA gotchas

- Hibernate mutates fields — `@Entity` classes need `var`, not `val`. Apply the `kotlin-jpa` plugin for no-arg constructors.
- Lazy loading needs proxying — use `open` classes or the `kotlin-allopen` plugin. `kotlin-spring` already does this for `@Entity`.
- Don't `copy()` a managed entity — it creates a detached instance.

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

### Notes

- Avoid `FetchType.LAZY` on `@OneToMany` unless the session is guaranteed open — prefer explicit JPQL joins
- `@EntityGraph` controls fetch plan without N+1
- `@BatchSize` on collections reduces N+1
- Pagination: `Pageable` parameter in repository methods
- `findByIdOrNull` (Kotlin extension) over `findById` which returns `Optional`
