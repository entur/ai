# SQL / Database

Entur uses **Cloud SQL for PostgreSQL** as the standard relational database. Apps connect through the **Cloud SQL Auth Proxy** sidecar on `localhost:5432`. This guide covers application-side query design, schema, transactions, pooling, migrations, and operational practices that affect performance and reliability.

For provisioning and machine sizing see [terraform/modules.md](terraform/modules.md#cloud-sql-postgresql). For the proxy sidecar see [helm.md](helm.md#database-cloud-sql-proxy). For language-specific access patterns see [java.md](java.md), [kotlin.md](kotlin.md), and [go.md](go.md).

> **About this guide**: *Schema Design*, *Indexing*, *Query Performance*, and the *Online-safe patterns* under *Migrations* are general Postgres practice we expect in Entur services -- not unique to our platform. The genuinely Entur-specific operational requirements start at *Connection Pooling* (HPA pod-count math, Cloud SQL proxy restart window) and continue through *Security*, *Testing*, and *Observability*. If you already know Postgres well, skim the early sections and read the later ones carefully.

## Scope

- **Database**: Cloud SQL Postgres (default major version: the latest GA supported by the Entur Terraform module)
- **Connectivity**: Cloud SQL proxy sidecar -- no direct public IP
- **Auth**: Prefer **IAM database auth**, fall back to password from Secret Manager
- **Migrations**: Flyway (Java/Kotlin); `golang-migrate` (Go)
- **Out of scope**: BigQuery, AlloyDB, BigTable, Spanner -- those have their own patterns

## Schema Design

### Naming

- Tables and columns in `snake_case` (see [CONVENTIONS.md](../CONVENTIONS.md#naming-conventions))
- Table names plural (`orders`, `route_segments`); column names singular
- Primary key column: `id`; foreign keys: `<referenced_table_singular>_id` (e.g. `order_id`)
- Index names: `ix_<table>_<col1>_<col2>`; uniques: `ux_<table>_<col>`

### Types

- Use `bigint` (`bigserial`) primary keys; avoid `int` -- migrations to widen later are expensive
- Use `timestamptz` for any time -- never `timestamp` (without zone)
- Use `text` rather than `varchar(n)` unless you need a real length constraint
- Use real Postgres `enum` types only for **stable, application-controlled** value sets. For anything user-editable, use a lookup table with FK
- Use `jsonb` (not `json`) when needed -- but prefer real columns. JSONB is for genuinely variable shapes; queries against it need GIN indexes or expression indexes

### NOT NULL by default

Every column should be `NOT NULL` unless absence is a meaningful business state. Provide defaults where the value is derivable (`created_at timestamptz NOT NULL DEFAULT now()`).

### Audit columns

Every business table:

```sql
created_at timestamptz NOT NULL DEFAULT now(),
updated_at timestamptz NOT NULL DEFAULT now()
```

Update `updated_at` from the application layer (JPA `@LastModifiedDate`, Exposed `EntityHooks`, or explicit `UPDATE`), not a trigger -- triggers hide cost.

## Indexing

### When to add an index

Add an index when **all** of the following hold:

1. The column appears in `WHERE`, `JOIN`, or `ORDER BY` of a query that runs frequently or against many rows
2. The column has reasonable selectivity (a boolean with a 50/50 split is rarely worth indexing)
3. The write cost is acceptable -- every index slows `INSERT`/`UPDATE`/`DELETE`

Every **foreign key column** should be indexed. Postgres does not create FK indexes automatically and unindexed FKs cause table scans on cascade and on join.

### Composite indexes

Column order matters. Put the most selective / equality columns first, range columns last:

```sql
-- Good for: WHERE tenant_id = ? AND status = ? AND created_at > ?
CREATE INDEX ix_orders_tenant_status_created
  ON orders (tenant_id, status, created_at);
```

A composite `(a, b, c)` also serves queries on `(a)` and `(a, b)` -- never `(b)` alone.

### Partial indexes

For queries that hit a small subset of a large table:

```sql
CREATE INDEX ix_orders_pending
  ON orders (created_at)
  WHERE status = 'PENDING';
```

### Covering indexes (`INCLUDE`)

When a query selects a few small columns alongside the index keys, `INCLUDE` lets Postgres serve the query index-only:

```sql
CREATE INDEX ix_orders_lookup
  ON orders (tenant_id, id)
  INCLUDE (status, total_cents);
```

### Avoid

- Over-indexing write-heavy tables -- benchmark write impact
- Indexing low-cardinality columns alone (booleans, status with 2--3 values) -- use partial indexes instead
- Functional indexes you never query through -- they cost writes for nothing

## Query Performance

### Always EXPLAIN

Before merging a non-trivial query, run `EXPLAIN (ANALYZE, BUFFERS)`. Use `EXPLAIN` (without `ANALYZE`) for a quick plan estimate; add `ANALYZE` when you need real timing and row counts.

> **Warning**: `ANALYZE` **executes the query**. On `INSERT`/`UPDATE`/`DELETE` this mutates data. Wrap write statements in a transaction you can roll back:
>
> ```sql
> BEGIN;
> EXPLAIN (ANALYZE, BUFFERS) UPDATE orders SET status = 'SHIPPED' WHERE id = 1;
> ROLLBACK;
> ```
>
> For read queries in production, prefer running against a read replica or a `tst` environment where real data and index statistics are representative.

```sql
EXPLAIN (ANALYZE, BUFFERS) <query>;
```

Look for:

- `Seq Scan` on large tables -- usually a missing or unused index
- `Rows Removed by Filter` >> rows returned -- predicate is not selective via index
- `Sort` with high cost -- consider an index that already orders the data
- Buffer hits vs reads -- `read=` numbers grow when the working set doesn't fit cache

### Pagination

Use **keyset (seek) pagination**, not `OFFSET`, for any list that can grow:

```sql
-- Bad: O(offset) scan on each page
SELECT * FROM orders ORDER BY id LIMIT 50 OFFSET 10000;

-- Good: O(log n) per page
SELECT * FROM orders WHERE id > :last_id ORDER BY id LIMIT 50;
```

### Avoid implicit casts

Implicit casts defeat indexes. Match parameter types to column types -- a `text` query parameter against a `uuid` column forces a full scan.

### Prefer set-based operations

Replace per-row loops with a single statement:

```sql
-- Instead of N round trips
UPDATE orders SET status = 'SHIPPED'
WHERE id = ANY(:ids);
```

Use `RETURNING` to avoid a follow-up `SELECT`:

```sql
INSERT INTO orders (...) VALUES (...)
RETURNING id, created_at;
```

### Never `SELECT *` in production code

It locks consumers to schema details, breaks projection optimisations, and wastes bytes over the wire. List the columns you need.

## N+1 and Joins

The single most common performance bug. Symptoms: one query for the parent, then one query per child row.

### Java / JPA

```java
// N+1: one query per order
List<Order> orders = orderRepository.findAll();
orders.forEach(o -> o.getItems().size());

// Fix: fetch graph
@EntityGraph(attributePaths = {"items"})
List<Order> findAll();
```

Use `@EntityGraph`, fetch joins (`JOIN FETCH`), or `@BatchSize(size = 50)` to batch lazy collections. Never trigger lazy loads inside a loop.

### Kotlin / Exposed

Define joins as reusable extensions on `Join` rather than fetching parents and then iterating to fetch children. See [kotlin.md](kotlin.md#joins-and-queries).

```kotlin
fun Join.withItems() = join(OrderItems, JoinType.LEFT, Orders.id, OrderItems.orderId)

Orders.innerJoin(OrderItems)
    .selectAll().where { Orders.tenantId eq tenant }
    .map { it.toOrderWithItem() }
```

### Go

Prefer one `JOIN` query with row aggregation in code over a loop of single-row fetches.

```go
rows, err := db.QueryContext(ctx, `
  SELECT o.id, o.status, i.id, i.sku
  FROM orders o
  LEFT JOIN order_items i ON i.order_id = o.id
  WHERE o.tenant_id = $1`, tenantID)
```

## Transactions

### Keep them short

Open the transaction immediately before the writes, close it immediately after. **Never** call out to HTTP, message brokers, or Redis from inside a transaction -- a slow external call holds the connection and blocks pool throughput.

### Mark read paths read-only

Read-only transactions enable Postgres optimisations and prevent accidental writes:

- Java / Spring: `@Transactional(readOnly = true)`
- Kotlin / Exposed: `transaction(readOnly = true) { ... }`
- Go / pgx: `tx, _ := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})`

### Isolation

Default to **READ COMMITTED** (Postgres default). Escalate only with reason:

| Level | Use |
|-------|-----|
| `READ COMMITTED` | Default. Almost all OLTP work. |
| `REPEATABLE READ` | Multi-statement reads that must see a stable snapshot. |
| `SERIALIZABLE` | Genuine concurrency-safety requirement (rare). Be ready to retry on `40001`. |

### Deadlocks

If two transactions touch the same rows in different order they will deadlock. Mitigate by:

- Touching rows in a deterministic order (sort by id before updating)
- Catching `SQLState 40P01` / `40001` and retrying with backoff (3 attempts)
- Splitting long writes into smaller transactions

### Advisory locks

For application-level coordination (e.g. "only one pod runs this job"), use `pg_advisory_lock` / `pg_try_advisory_lock` rather than building a lock table.

## Connection Pooling

Total connections to Cloud SQL = `pods × max_pool_size_per_pod`. Postgres reserves 3 superuser slots, so:

```text
pods_max × max_pool_size  ≤  cloud_sql_max_connections - 3
```

Cloud SQL `max_connections` is set by the Terraform module ([terraform/modules.md](terraform/modules.md#cloud-sql-postgresql)) and depends on instance tier. Plan against your HPA `maxReplicas`, not the steady-state pod count.

### Java / Kotlin (HikariCP)

```yaml
spring:
  datasource:
    hikari:
      maximum-pool-size: 10          # default; lower for many-pod deployments
      minimum-idle: 2
      connection-timeout: 3000       # ms -- fail fast, do not let requests pile up
      idle-timeout: 600000
      max-lifetime: 1800000          # 30 min; below Cloud SQL proxy restart window
```

See [java.md](java.md#connection-pool-sizing) for HPA pod-count math. Exposed in Kotlin shares the same DataSource.

### Go

```go
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(2)
db.SetConnMaxLifetime(30 * time.Minute)
db.SetConnMaxIdleTime(10 * time.Minute)
```

`ConnMaxLifetime` ≤ 30 min keeps the pool tolerant of Cloud SQL proxy restarts and IP changes.

### Metrics

Watch `hikaricp_connections_active` / `hikaricp_connections_pending` (Java/Kotlin) and your Go pool stats. Sustained pending > 0 means the pool is undersized **or** queries are too slow. See [observability.md](observability.md).

## Migrations (Flyway)

### Location and naming

```text
src/main/resources/db/migration/
  V202605211030__add_orders_status_index.sql
  V202605221400__backfill_order_totals.sql
  R__order_summary_view.sql
```

- Versioned: `V<yyyymmddHHmm>__<snake_case_description>.sql`
- Repeatable (re-runs when checksum changes): `R__<name>.sql` -- use for views and functions
- One logical change per migration

### Never edit applied migrations

Once a migration has run in any environment, it is immutable. Forward-fix with a new migration. Editing a checksum-validated migration will break deployment.

### Online-safe patterns

Postgres locks tables for many `ALTER` statements. On a busy table this stalls the app. Use these patterns:

**Adding a NOT NULL column**: do it in three releases.

1. `ALTER TABLE orders ADD COLUMN region text;` (nullable)
2. Application writes both old and new; backfill in batches
3. `ALTER TABLE orders ALTER COLUMN region SET NOT NULL;`

**Creating an index** on a large table:

```sql
CREATE INDEX CONCURRENTLY ix_orders_region ON orders (region);
```

`CONCURRENTLY` cannot run inside a transaction. Set `flyway.transactionalLock=false` for that migration, or split it.

**Renaming a column**: avoid. Add new, dual-write, backfill, switch reads, drop old.

**Dropping a column**: deploy code that no longer reads it first; drop in a later release.

### Test migrations

Run Flyway in your integration test bootstrap (Testcontainers) so a broken migration fails CI rather than a deploy.

## Security

- **Always parameterise queries.** Never concatenate user input into SQL. JPA, Exposed, and `pgx` parameterise by default; raw `Statement` / `fmt.Sprintf` builds are the bug.
- **Prefer IAM database authentication** over password auth (configured in the Entur Cloud SQL Terraform module). The app uses its workload-identity-bound service account; no rotating secret in Secret Manager.
- **Least privilege**: the application user owns its schema but is not `SUPERUSER`. Read replicas use a separate read-only role.
- **Audit / observability** via [Cloud SQL Query Insights](terraform/modules.md#cloud-sql-postgresql) -- enabled by the Terraform module.
- **No raw queries from string templates.** If you must build dynamic SQL (e.g. filter assembly), build it from a fixed set of column whitelisted names plus parameters, never concatenated user input.

See the security checklist in [code-review.md](code-review.md#database).

## Testing

- Use **Testcontainers PostgreSQL** for integration tests -- never an in-memory database (H2/HSQLDB). Behaviour diverges from real Postgres and bugs slip through.
- Run Flyway during test startup against the container so migrations are exercised on every CI run.
- Use `@Sql` to load fixtures (see [CONVENTIONS.md](../CONVENTIONS.md#testing)) and a `cleanup.sql` between tests for deterministic state.
- For Go, use the `postgres` module of [`testcontainers-go`](https://github.com/testcontainers/testcontainers-go) and run migrations via `golang-migrate` in the test setup.

```kotlin
@TestConfiguration
class TestContainersConfig {
  @Bean
  @ServiceConnection
  fun postgres() = PostgreSQLContainer("postgres:16-alpine")
}
```

Detail patterns per language: [kotlin.md](kotlin.md#integration-tests-testcontainers), [java.md](java.md#testing).

## Observability

- **Pool**: `hikaricp_connections_active`, `hikaricp_connections_pending`, `hikaricp_connections_usage_seconds` (Java/Kotlin); equivalent gauges via `database/sql.DBStats` for Go. Surface `db_pool_active_connections` as documented in [observability.md](observability.md).
- **Query latency**: enable `pg_stat_statements` (default on in the Entur Cloud SQL module) and inspect via Query Insights.
- **Slow queries**: investigate anything > 1 s. Set `log_min_duration_statement = '500ms'` for triage windows.
- **Readiness probe**: include a 1-statement DB ping (`SELECT 1`) only if the service genuinely cannot serve requests without the DB. Otherwise keep it out of readiness so a transient DB blip doesn't take traffic away from a healthy pod.

## Anti-patterns Checklist

Use during [code review](code-review.md#database):

- [ ] `OFFSET`-based pagination on tables that grow
- [ ] `SELECT *` in production code
- [ ] SQL built with string concatenation or string templates over user input
- [ ] N+1 patterns: lazy loads inside loops, per-row child queries
- [ ] Long-running transactions or transactions that wrap HTTP / Kafka / Redis calls
- [ ] Missing index on a foreign key
- [ ] Repository methods that return unbounded result sets (`findAll()` on growing tables)
- [ ] Editing an already-applied Flyway migration instead of writing a new one
- [ ] `CREATE INDEX` (without `CONCURRENTLY`) on a large production table
- [ ] In-memory DB (H2/HSQLDB) used for tests instead of Testcontainers Postgres
