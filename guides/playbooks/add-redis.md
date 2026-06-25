# Add Memorystore Redis

Provision Memorystore for Redis (or Valkey) and wire it into a service for caching, locks, rate limiting, or Kafka dedup.

## Goal

A managed Redis instance, credentials injected as env vars, language client configured with sensible timeouts and TTL discipline.

## Prerequisites

- The application is already on the platform (otherwise start with [bootstrap-service.md](bootstrap-service.md)).
- You have a real cache use case. Redis is **not** a primary data store -- see [java.md](../reference/java.md#when-to-use-redis) for the decision table.

## Steps

1. **Provision the instance via the Entur Terraform module.** Use `terraform-google-memorystore//modules/redis?ref=v2`. The module creates the instance and `REDIS_HOST`/`REDIS_PORT`/`REDIS_PASSWORD` secrets. Configure `maxmemory-policy: allkeys-lfu` so eviction is predictable. See [terraform-modules.md](../platform/terraform-modules.md#memorystore-redis).

2. **Mount the secrets into the pod via Helm.** Add a `common.secrets.redis-credentials` entry listing the three keys. See [common-helm.md](../platform/common-helm.md#secrets-externalsecrets).

3. **Configure the client.** Spring Boot: see [java.md](../reference/java.md#redis-memorystore) (Lettuce pool sizing, key naming conventions, cache-aside via `@Cacheable`). Go: see [go.md](../reference/go.md#redis-memorystore) (`go-redis` setup, `redis.Nil` handling).

4. **Choose your access pattern.** Cache-aside for read-heavy DB lookups, `SET NX EX` for distributed locks and Kafka dedup, `INCR + EXPIRE` for rate limiting. Always set a TTL -- unbounded growth exhausts memory.

5. **Decide on readiness inclusion.** Redis is a private dependency of *your* service only -- including it in readiness is acceptable. Do **not** include shared Redis (rare) -- a shared outage would remove all pods. See [observability.md](../reference/observability.md#readiness-probe).

6. **Namespace your keys.** `{app}:{domain}:{id}` for entity cache, `{app}:rate:{client}` for limits, `{app}:dedup:{eventId}` for Kafka dedup. Use `SCAN`, not `KEYS *`, in production.

## Verify

- `terraform plan` shows the instance, secrets, and ConfigMap creation.
- After deploy, `kubectl logs` shows the client connecting on startup.
- Grafana Memorystore dashboard shows non-zero ops/sec and bounded memory growth.

## See also

- Kafka deduplication: [add-kafka.md](add-kafka.md)
- Helm secrets pattern: [common-helm.md](../platform/common-helm.md#secrets-externalsecrets)
