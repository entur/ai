# Go Standards

Go conventions for Entur applications. Read [CONVENTIONS.md](../../CONVENTIONS.md) first for cross-language standards.

## Runtime and Build

- **Go version**: latest stable (currently 1.25+)
- **Modules**: Go modules (`go.mod`)
- **Linting**: `golangci-lint`
- **Framework**: standard library `net/http` (Go 1.25+ routing) or minimal frameworks only

## Project Structure

```text
my-service/
  cmd/
    my-service/
      main.go               # Entry point
  internal/
    handler/                # HTTP handlers
    service/                # Business logic
    repository/             # Data access
    model/                  # Domain types
    config/                 # Configuration loading
  pkg/                      # Public reusable packages (if any)
  Dockerfile
  go.mod
  go.sum
  .golangci.yml
```

- `internal/` for application-private code (enforced by Go compiler)
- `cmd/` for application entry points
- `pkg/` only for code intended to be imported by other projects (rare)

## Dockerfile

See [docker.md](docker.md) for Dockerfile conventions and base images. Go-specific example:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/my-service

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

## Application Patterns

### Main Entry Point

The main function should:

- Load configuration from environment variables
- Create `http.ServeMux` and register routes
- Configure `http.Server` with timeouts (ReadTimeout: 10s, WriteTimeout: 30s, IdleTimeout: 60s)
- Implement graceful shutdown using `signal.NotifyContext` for SIGINT/SIGTERM
- Use `server.Shutdown` with a timeout context for clean shutdown

### Health Checks

Register liveness (`GET /health/liveness`) and readiness (`GET /health/readiness`) endpoints. Liveness returns `200 OK` unconditionally. Readiness checks private dependencies (e.g., DB ping) and returns `503` if unavailable.

Helm values for custom health paths:

```yaml
common:
  container:
    probes:
      liveness:
        path: /health/liveness
      readiness:
        path: /health/readiness
```

### HTTP Handlers

Use handler structs with service and logger dependencies, created via constructor functions (`NewXxxHandler`). Use `r.PathValue("id")` for path parameters (Go 1.22+). Return JSON with `json.NewEncoder(w).Encode()`. Handle errors by returning appropriate HTTP status codes -- ALWAYS return client-safe error responses.

### Error Handling

- Define sentinel errors for known conditions (`var ErrNotFound = errors.New(...)`)
- Wrap errors with `fmt.Errorf("context: %w", err)` to preserve the chain
- Use sentinel errors for expected conditions
- Use `errors.Is()` / `errors.As()` to check error types
- ALWAYS handle errors explicitly

### Configuration

Use environment variables for all configuration. Use `caarlos0/env` struct tags (e.g., `env:"PORT" envDefault:"8080"`) or `os.Getenv`. Create a `Config` struct and a `Load()` function that returns `(*Config, error)`.

## Testing

- Use `testing` package with `t.Run` for subtests
- Use `testify/require` for fatal assertions, `testify/assert` for non-fatal
- Use table-driven tests for multiple input scenarios
- Write mocks by hand or use `testify/mock`
- Use `testcontainers-go` for integration tests with databases

## Postgres (Cloud SQL)

Entur uses **Google Cloud SQL for PostgreSQL**. Infrastructure via `terraform-google-sql-db` (see [terraform-modules.md](../platform/terraform-modules.md#cloud-sql-postgresql)). The app connects via the Cloud SQL Auth Proxy sidecar (configured in Helm, see [common-helm.md](../platform/common-helm.md#database-cloud-sql-proxy)) at `localhost:5432`. Credentials (`DB_NAME`, `PG_USER`, `PG_PASSWORD`) inject from Kubernetes secrets.

For query design, indexing, transactions, migrations, and Cloud SQL operational patterns see [sql.md](sql.md). Go-specific notes below.

### Driver

Prefer [`jackc/pgx/v5`](https://github.com/jackc/pgx) (`pgxpool` for pooled access). `database/sql` + `lib/pq` is acceptable when an existing dependency requires the `sql.DB` interface.

### Connection Pool

Tune against worst-case HPA pod count -- total connections = `pods * MaxOpenConns` must stay below Cloud SQL `max_connections - 3`.

```go
pool.Config().MaxConns = 10
pool.Config().MinConns = 2
pool.Config().MaxConnLifetime = 30 * time.Minute
pool.Config().MaxConnIdleTime = 10 * time.Minute
```

For `database/sql`: `db.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` -- keep `ConnMaxLifetime` ≤ 30 minutes so the pool tolerates Cloud SQL proxy restarts.

### Migrations

Use [`golang-migrate`](https://github.com/golang-migrate/migrate) with migration files in `db/migration/`. Run on application startup or as a dedicated K8s Job. See migration safety patterns in [sql.md](sql.md#migrations-flyway).

### Testing

Use `testcontainers-go` with the `postgres` module and run migrations during test setup so a broken migration fails CI rather than a deploy.

## Redis (Memorystore)

Entur uses **Google Memorystore for Redis**. Infrastructure via `terraform-google-memorystore` (see [terraform/modules.md](../platform/terraform-modules.md#memorystore-redis)). Credentials (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`) injected via Kubernetes secrets.

For use cases, key naming conventions, and best practices, see [java.md](java.md#redis-memorystore). This section covers Go-specific implementation.

### Client Setup

Use `go-redis/redis` (v9+). Create a `NewClient(cfg *Config)` function that configures `redis.Options` with address, password, timeouts (DialTimeout: 1s, ReadTimeout: 2s, WriteTimeout: 2s), and pool settings (PoolSize: 10, MinIdleConns: 2). Verify connectivity with `client.Ping()`.

Configuration uses `REDIS_HOST` (required), `REDIS_PORT` (default 6379), and `REDIS_PASSWORD` (required) environment variables.

### Basic Operations

Key operations: `Set` (with TTL), `Get` (check `redis.Nil` for cache miss), `Del`, `SetNX` (distributed locks/idempotency), `Incr` with `Expire` (rate limiting). Always handle `redis.Nil` separately from other errors -- fall back to database on cache miss or Redis error.

### Cache-Aside Pattern

Implement cache-aside with a struct wrapping both `redis.Client` and a repository. Try cache first (`Get`), fall back to database on miss, then populate cache with TTL (best-effort `Set`). Include an invalidation method using `Del`. Use JSON marshal/unmarshal for serialization.

### Health Check with Redis

Include Redis in readiness only if it's a **private resource owned by this service** (see [observability.md](observability.md#readiness-probe)). Ping both database and Redis in the readiness handler; return `503` with a `reason` field if either is down.

### Testing

Use `testcontainers-go` for Redis integration tests. Create a helper `setupRedis(t *testing.T) *redis.Client` that starts a `redis:7-alpine` container, waits for the port, and returns a connected client. Use `t.Cleanup` for container termination.

## Prometheus Metrics

Register the Prometheus HTTP handler at `GET /metrics` using `promhttp.Handler()` from `github.com/prometheus/client_golang`.

Helm values:

```yaml
common:
  container:
    prometheus:
      enabled: true
      path: /metrics
```

## Logging

Use [entur/go-logging](https://github.com/entur/go-logging) for structured JSON logging on GCP. See [logging.md](logging.md) for general standards.

### Install

```bash
go get github.com/entur/go-logging
```

### Usage

Use global functions (`logging.Info()`, `logging.Error()`) or create instance loggers with `logging.New()` for dependency injection. Chain fields with `.Str()`, `.Err()`, `.Msgf()` etc. Options: `logging.WithLevel()`, `logging.WithNoCaller()`.

### Log Level

Default level from `LOG_LEVEL` env var (defaults to `warning`).

Valid values: `fatal`, `panic`, `error`, `warning`, `info`, `debug`, `trace` (or short: `ftl`, `pnc`, `err`, `wrn`, `inf`, `dbg`, `trc`).

```yaml
common:
  container:
    env:
      - name: LOG_LEVEL
        value: "info"    # info for dev/tst, warning for prd
```

### Errors with Stacktraces

Use `logging.NewStackTraceError()` to create errors with stacktraces, or wrap existing errors with `logging.NewStackTraceError("%w", existingErr)`. Stacktraces are included automatically when logged with `.Err()`.
