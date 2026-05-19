# Run the Service Locally

Set up a working local dev loop: pinned tool versions, Docker Compose for dependencies, local Spring profile, secrets fetched on demand.

## Goal

A developer (or agent) can clone the repo and run `./gradlew bootRun` (or `go run ./cmd/...`) against a local Postgres/Redis/Kafka without manual setup beyond `mise install`.

## Prerequisites

- [mise](https://mise.jdx.dev/) installed locally.
- `gcloud` CLI authenticated with Entur SSO if secrets need to be fetched.
- Docker / Docker Desktop / colima for the Compose stack.

## Steps

1. **Pin tool versions in `.mise.toml`.** Declare Java/Kotlin/Go/Terraform versions, then `mise install` once provisions everything. See [CONVENTIONS.md](../../CONVENTIONS.md#tool-version-management-mise).

2. **Add a `compose.yaml` at the repository root.** Spin up Postgres, Redis, and any other local dependencies. Use minimal images (`postgres:16-alpine`, `redis:7-alpine`). See [CONVENTIONS.md](../../CONVENTIONS.md#docker-compose-and-local-spring-profile).

3. **Add `application-local.yml` for Spring Boot.** Human-readable logging (`entur.logging.style: humanReadablePlain`), local DB URL pointing at the Compose Postgres, Swagger UI enabled, `FULL_ACCESS` permission cache or `LOCAL_TEST_CACHE` with named test users. See [permission-store.md](../platform/permission-store.md#full_access-development) and [logging.md](../reference/logging.md#local-development).

4. **Fetch secrets on demand via `gcloud`.** Never commit secrets to `application-local.yml`. Use `gcloud secrets versions access latest --secret <name> --project ent-<appid>-dev` and export to env vars. See [security.md](../reference/security.md#local-development).

5. **Use the public Aiven Kafka cluster for local Kafka.** Set `entur.kafka.kafkaCluster: AIVEN_PUBLIC_TEST_INT` in the local profile -- the `*_INT` clusters require running inside the VPC. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#per-environment-configuration).

6. **Run the app.** Spring Boot: `SPRING_PROFILES_ACTIVE=local ./gradlew bootRun`. Go: `go run ./cmd/<service>` with environment variables set. Both should expose the standard health endpoints on `localhost:8080`.

## Verify

- `curl localhost:8080/actuator/health/readiness` returns 200.
- Logs are human-readable, not JSON.
- The application can reach the Compose Postgres / Redis and complete a simple read/write.
- For Kafka, a local Testcontainers integration test passes (`./gradlew test`).

## See also

- Bootstrap a new service: [bootstrap-service.md](bootstrap-service.md)
- Logging in dev vs prd: [logging.md](../reference/logging.md)
