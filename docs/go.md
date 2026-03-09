# Go Standards

Go conventions for Entur applications. Read [CONVENTIONS.md](../CONVENTIONS.md) first for cross-language standards.

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

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/entur/go-logging"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        logging.Error().Err(err).Msg("failed to load configuration")
        os.Exit(1)
    }

    mux := http.NewServeMux()
    registerRoutes(mux, cfg)

    server := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        logging.Info().Msgf("server starting on port %s", cfg.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logging.Error().Err(err).Msg("server failed")
            os.Exit(1)
        }
    }()

    <-ctx.Done()
    logging.Info().Msg("shutting down gracefully")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := server.Shutdown(shutdownCtx); err != nil {
        logging.Error().Err(err).Msg("forced shutdown")
    }
}
```

### Health Checks

```go
func registerRoutes(mux *http.ServeMux, cfg *config.Config) {
    mux.HandleFunc("GET /health/liveness", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"UP"}`))
    })

    mux.HandleFunc("GET /health/readiness", func(w http.ResponseWriter, r *http.Request) {
        if err := cfg.DB.PingContext(r.Context()); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte(`{"status":"DOWN"}`))
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"UP"}`))
    })
}
```

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

```go
type RouteHandler struct {
    service *RouteService
    logger  *logging.Logger
}

func NewRouteHandler(service *RouteService) *RouteHandler {
    return &RouteHandler{
        service: service,
        logger:  logging.New(),
    }
}

func (h *RouteHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    route, err := h.service.FindByID(r.Context(), id)
    if err != nil {
        h.logger.Error().Err(err).Str("id", id).Msg("failed to find route")
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    if route == nil {
        http.Error(w, "route not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(route)
}
```

### Error Handling

```go
// Define sentinel errors for known conditions
var (
    ErrRouteNotFound = errors.New("route not found")
    ErrInvalidInput  = errors.New("invalid input")
)

// Wrap errors with context
func (s *RouteService) FindByID(ctx context.Context, id string) (*Route, error) {
    route, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("finding route %s: %w", id, err)
    }
    return route, nil
}
```

- Wrap errors with `fmt.Errorf("context: %w", err)` to preserve the chain
- Use sentinel errors for expected conditions
- Use `errors.Is()` / `errors.As()` to check error types
- Never ignore errors

### Configuration

```go
type Config struct {
    Port        string `env:"PORT" envDefault:"8080"`
    DatabaseURL string `env:"DATABASE_URL,required"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }
    return cfg, nil
}
```

Use environment variables for all configuration. Use `caarlos0/env` or `os.Getenv`.

## Testing

```go
func TestRouteService_FindByID(t *testing.T) {
    t.Run("returns route when found", func(t *testing.T) {
        repo := &mockRepo{
            routes: map[string]*Route{"r1": {ID: "r1", Origin: "Oslo"}},
        }
        svc := NewRouteService(repo)

        route, err := svc.FindByID(context.Background(), "r1")

        require.NoError(t, err)
        assert.Equal(t, "r1", route.ID)
        assert.Equal(t, "Oslo", route.Origin)
    })

    t.Run("returns nil when not found", func(t *testing.T) {
        repo := &mockRepo{routes: map[string]*Route{}}
        svc := NewRouteService(repo)

        route, err := svc.FindByID(context.Background(), "unknown")

        require.NoError(t, err)
        assert.Nil(t, route)
    })
}
```

- Use `testing` package with `t.Run` for subtests
- Use `testify/require` for fatal assertions, `testify/assert` for non-fatal
- Use table-driven tests for multiple input scenarios
- Write mocks by hand or use `testify/mock`
- Use `testcontainers-go` for integration tests with databases

## Redis (Memorystore)

Entur uses **Google Memorystore for Redis**. Infrastructure via `terraform-google-memorystore` (see [terraform/modules.md](terraform/modules.md#memorystore-redis)). Credentials (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`) injected via Kubernetes secrets.

For use cases, key naming conventions, and best practices, see [java.md](java.md#redis-memorystore). This section covers Go-specific implementation.

### Client Setup

Use [`go-redis/redis`](https://github.com/redis/go-redis) (v9+):

```go
package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

func NewClient(cfg *Config) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
        Password:     cfg.RedisPassword,
        DB:           0,
        DialTimeout:  1 * time.Second,
        ReadTimeout:  2 * time.Second,
        WriteTimeout: 2 * time.Second,
        PoolSize:     10,
        MinIdleConns: 2,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("connecting to redis: %w", err)
    }
    return client, nil
}
```

Configuration:

```go
type Config struct {
    RedisHost     string `env:"REDIS_HOST,required"`
    RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
    RedisPassword string `env:"REDIS_PASSWORD,required"`
}
```

### Basic Operations

```go
// SET with TTL
err := client.Set(ctx, "myapp:route:123", routeJSON, 10*time.Minute).Err()

// GET
val, err := client.Get(ctx, "myapp:route:123").Result()
if errors.Is(err, redis.Nil) {
    // Key does not exist -- cache miss, fall back to database
} else if err != nil {
    // Redis error -- log and fall back to database
}

// DELETE
err = client.Del(ctx, "myapp:route:123").Err()

// SET if not exists (distributed lock / idempotency)
ok, err := client.SetNX(ctx, "myapp:lock:import-job", "owner-1", 60*time.Second).Result()
if ok {
    // Lock acquired
}

// Atomic increment (rate limiting)
count, err := client.Incr(ctx, "myapp:rate:client-xyz").Result()
if count == 1 {
    client.Expire(ctx, "myapp:rate:client-xyz", 1*time.Minute)
}
```

### Cache-Aside Pattern

```go
type RouteCache struct {
    redis *redis.Client
    repo  RouteRepository
    ttl   time.Duration
}

func (c *RouteCache) GetRoute(ctx context.Context, id string) (*Route, error) {
    key := "myapp:route:" + id

    // Try cache first
    val, err := c.redis.Get(ctx, key).Bytes()
    if err == nil {
        var route Route
        if err := json.Unmarshal(val, &route); err == nil {
            return &route, nil
        }
    }

    // Cache miss or error -- fall back to database
    route, err := c.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if route == nil {
        return nil, nil
    }

    // Store in cache (best-effort)
    if data, err := json.Marshal(route); err == nil {
        _ = c.redis.Set(ctx, key, data, c.ttl).Err()
    }

    return route, nil
}

func (c *RouteCache) InvalidateRoute(ctx context.Context, id string) {
    _ = c.redis.Del(ctx, "myapp:route:"+id).Err()
}
```

### Health Check with Redis

Include Redis in readiness only if it's a **private resource owned by this service** (see [observability.md](observability.md#readiness-probe)):

```go
mux.HandleFunc("GET /health/readiness", func(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    if err := db.PingContext(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(`{"status":"DOWN","reason":"database"}`))
        return
    }

    if err := redisClient.Ping(ctx).Err(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(`{"status":"DOWN","reason":"redis"}`))
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"UP"}`))
})
```

### Testing

Use Testcontainers for integration tests:

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupRedis(t *testing.T) *redis.Client {
    ctx := context.Background()
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "redis:7-alpine",
            ExposedPorts: []string{"6379/tcp"},
            WaitingFor:   wait.ForListeningPort("6379/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    t.Cleanup(func() { container.Terminate(ctx) })

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "6379")

    return redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%s", host, port.Port()),
    })
}
```

## Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

mux.Handle("GET /metrics", promhttp.Handler())
```

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

```go
import "github.com/entur/go-logging"

// Global logger (includes caller info automatically)
logging.Info().Msg("request processed")
logging.Error().Err(err).Str("query", queryName).Msg("database query failed")
logging.Info().Msgf("processed %d routes in %dms", count, elapsed.Milliseconds())

// Instance logger (for dependency injection)
logger := logging.New()
logger.Info().Msg("request processed")

// Instance logger with custom level
logger = logging.New(logging.WithLevel(logging.DebugLevel))

// Disable caller info when not needed
logger = logging.New(logging.WithNoCaller())
```

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

```go
// New error with stacktrace
err := logging.NewStackTraceError("route not found: %s", routeID)

// Wrap an existing error with stacktrace
err = logging.NewStackTraceError("%w", existingErr)

// Log it -- stacktrace is included automatically
logging.Error().Err(err).Msg("an internal error occurred")
```
