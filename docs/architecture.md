# Architecture Standards

Guidelines for application architecture at Entur. All services run on GKE in `europe-west1`.

## Design Methodology

Entur combines **Domain-Driven Design (DDD)** with **Hexagonal Architecture**:

- **DDD**: common language between technical and domain experts, testable code by design, implies SOLID principles
- **Hexagonal Architecture** (ports and adapters): domain core independent of frameworks/databases/external services -- outer layers depend on inner layers, never the reverse
- **Reactive Systems** principles: responsive, elastic, resilient, event-driven

## Service Architecture

### Microservice Principles

- Each service owns its data -- no shared databases
- Communicate via REST, gRPC, or Kafka
- Independently deployable with own repository, CI/CD pipeline, and K8s namespace
- Golden path: repository name = application name = namespace = Helm release name
- **"You build it, you run it"**
- All versions must support rollback; maintain backwards compatibility
- Applications must start even when dependencies are unavailable

### Layered Architecture (within a service)

```text
Controller                 # HTTP request handling, DTO transformation
       |
    Mapper                 # DTO <-> Domain transformation
       |
   Service                 # Business logic, validation, orchestration
       |
     DAO                   # Data access, SQL queries
       |
   Entity                  # Database table definition
       |
  Domain Model             # Domain data classes, value objects
```

Key rules:

- **Controllers**: HTTP concerns only -- receive DTOs, call mappers, return DTOs with status codes
- **Mappers**: transform between generated DTOs and domain models at API boundary
- **Services**: business logic, orchestrate DAOs. Define interfaces with implementations
- **DAOs**: data persistence (Exposed or Spring Data)
- **Entities**: database table structure
- **Domain Models**: plain data classes, no framework dependencies
- Depend inward; keep controllers thin; never leak entities into API responses
- Service layer works exclusively with domain models, never DTOs

### Package-by-Feature

Organize by domain feature, not by technical layer:

```text
org.entur.myapp/
  common/
    api/                   # Shared API utilities (ErrorResponse, GlobalExceptionHandler)
    base/                  # Base interfaces (BaseDAO, BaseService)
    exception/             # Typed exceptions
    utils/                 # Shared utilities, constants
  config/                  # Spring configuration classes
  version/                 # Feature package
    Version.kt             # Domain model (data class)
    VersionController.kt   # REST controller (implements generated interface)
    VersionDAO.kt           # Data access object
    VersionEntity.kt        # Database table definition (Exposed)
    VersionInputValidator.kt # Input validation rules
    VersionMapper.kt        # DTO <-> Domain mapper
    VersionService.kt       # Service interface
    VersionServiceImpl.kt   # Service implementation
  priceableobject/         # Another feature package
    PriceableObjectEntity.kt
    PriceableObjectDAO.kt
    fareprice/             # Sub-feature
      FarePriceEntity.kt
```

### Mapper Pattern

Dedicated mapper classes transform between generated DTOs and domain models. See [kotlin.md](kotlin.md) for implementation details.

```text
Controller Layer       ← receives/returns DTOs
       ↓
   Mapper Layer        ← transforms DTO ↔ Domain
       ↓
  Service Layer        ← works ONLY with domain models
       ↓
    DAO Layer           ← transforms Domain ↔ Entity/ResultRow
```

Benefits: API contracts separated from business logic, API changes don't ripple through codebase, read-only fields handled at boundary, testable in isolation.

### Composition Over Inheritance

Prefer constructor injection over inheritance hierarchies:

```kotlin
@Service
class VersionServiceImpl(private val versionDAO: VersionDAO) : VersionService {
    // Business logic delegates to DAO -- composition, not inheritance
}
```

Entities compose relationships using extension functions:

```kotlin
object PriceableObjectEntity : LongIdTable("priceable_object") {
    fun Join.joinPriceableObjectWithChildren() =
        Join(this)
            .joinFarePriceToPriceableObject()
            .joinLimitingRuleToFarePrice()
            .withVersionsForPriceableObjects()
}
```

## Infrastructure Architecture

### GCP Project Structure

- Each application has a GCP project per environment (dev, tst, prd)
- `terraform-google-init` module discovers project configuration
- Shared networking managed centrally (VPC, subnets)

### Data Stores

| Need | Service | Provisioning |
|------|---------|-------------|
| Relational data | Cloud SQL (PostgreSQL) | `terraform-google-sql-db` module |
| Caching | Memorystore (Redis) | `terraform-google-memorystore` module |
| Object storage | Cloud Storage | [`terraform-google-cloud-storage`](https://github.com/entur/terraform-google-cloud-storage) module |
| Event streaming | Apache Kafka (Aiven) | See [kafka.md](kafka.md) |
| Analytics | BigQuery | `google_bigquery_dataset` / `table` resources |

Always use Entur Terraform modules for Cloud SQL and Memorystore -- they handle naming, networking, secrets, and K8s integration.

### Connectivity

- **Cloud SQL**: via Cloud SQL proxy sidecar (`postgres.enabled: true` in Helm)
- **Memorystore**: via private IP (auto-configured, exposed as `REDIS_HOST`)
- **Inter-service**: Kubernetes service DNS

## Asynchronous Messaging

Entur uses **Apache Kafka on Aiven**. See [kafka.md](kafka.md) for full documentation.

### Message Design

- Use **Avro** (default) or **Protobuf** -- schemas managed via Confluent Schema Registry
- Include correlation ID for tracing (auto-added by starter as `X-Correlation-Id` header)
- Keep messages self-contained -- avoid requiring callbacks to producer
- Use Gradle Avro plugin (`com.github.davidmc24.gradle.plugin.avro`) to generate classes from `.avsc` files

## Database Design

### PostgreSQL Conventions

- `snake_case` for table and column names
- Singular table names: `route`, `stop_place`
- Always include `id`, `created`, `changed` columns
- Use UUID/ULID for distributed systems or `Long` auto-increment for simpler domains
- Use Flyway for migrations; prefix with version: `V1__create_route_table.sql`
- Organize migrations: `schema/` for DDL, `reusablefunctions/` for stored procedures, `kotlin/` for code-based
- Use database views for complex reads spanning multiple tables

### SQL Library Choice (Kotlin)

Prefer **JetBrains Exposed SQL-DSL** over JPA/Hibernate for Kotlin:

- Typesafe Kotlin DSL close to SQL (not an ORM)
- Works naturally with immutable `data class` models
- Better query/join/subquery control; no magic, no lazy loading

```kotlin
object VersionEntity : LongIdTable("version") {
    val netexId = varchar("netex_id", 70)
    val created = timestamp("created")
    val status = varchar("status", 70)
    val startDate = date("start_date")
    val endDate = date("end_date").nullable()
}
```

For Java projects, **Spring Data JPA** remains the default.

### Migration Best Practices

- Migrations must be backward-compatible (rolling deployments)
- Separate schema changes from data migrations
- Never modify applied migrations
- Test against production data copy before deploying
- Use `baseline-on-migrate` in test configuration

## Resilience Patterns

Use circuit breakers, retry with backoff, and timeouts for external service calls. See [api-design.md](api-design.md#rate-limiting-and-resilience) for details.

```java
@CircuitBreaker(name = "externalService", fallbackMethod = "fallback")
public RouteData fetchFromExternalService(String id) { ... }

@Retry(name = "externalService", fallbackMethod = "fallback")
public RouteData fetchWithRetry(String id) { ... }
```

Timeout guidelines: connect 5s, read 30s, never infinite. Design for graceful degradation -- return cached or partial data when dependencies are unavailable.

## Production Hardening

All production services must meet these requirements. The common Helm chart handles most automatically. See [helm.md](helm.md) for configuration.

### Multiple Replicas

Production must run >1 pod. Nodes can be downscaled at any time.

### Horizontal Pod Autoscaler (HPA)

Auto-configured in production. Scales between `replicas` and `maxReplicas` (default 10) based on CPU (default 80%). Favor horizontal over vertical scaling.

### Pod Disruption Budget (PDB)

Auto-created in production with `minAvailable: 50%`. In dev/tst with single pod, set `minAvailable: 0`.

### Pod Anti-Affinity

Distributes workloads across nodes. Auto-configured by common Helm chart.

### Zone Anti-Affinity

Distributes across availability zones for critical production services.

### Vertical Pod Autoscaler (VPA)

Enabled by default for resource recommendations. Takes weeks to stabilize for new deployments.

### Resource Sizing

- CPU: set request for normal load, **no limit** (allow bursting)
- Memory: set request and limit to **same value** (exceeding causes OOM kill)
- Start small, tune based on VPA recommendations

## Application Lifecycle

### Deprecating an Application

1. Verify traffic is zero via Grafana dashboard
2. Scale to 0 replicas in GCP Console (soft delete)
3. Clean up Apigee proxies (`#talk-utviklerplattform`)
4. Request domain name removal if applicable

### Deleting an Application

1. Delete `.entur` folder and follow self-service workflow
2. Archive the GitHub repository
3. Clean up container artifacts

A deleted GCP project can be restored within 30 days via `#talk-utviklerplattform`.
