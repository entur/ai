# Architecture Standards

Guidelines for application architecture at Entur. All services run on GKE in `europe-west1`.

## Design Principles

- Keep business logic independent of frameworks, databases, and external services
- Outer layers depend on inner layers (dependency inversion)
- Design for testability -- business logic should be testable without infrastructure
- Build responsive and resilient systems

## Service Architecture

### Microservice Principles

- Each service owns its data -- no shared databases
- Communicate via REST, gRPC, Kafka, or Pub/Sub
- ALWAYS use the Entur `common` Helm chart for Kubernetes deployments
- Independently deployable with own repository, CI/CD pipeline, and K8s namespace
- Golden path: repository name = application name = namespace = Helm release name
- **"You build it, you run it"**
- All versions must support rollback; maintain backwards compatibility
- Applications must start even when dependencies are unavailable

## Infrastructure

### GCP Project Structure

- Each application has a GCP project per environment (dev, tst, prd)
- `terraform-google-init` module discovers project configuration
- Shared networking managed centrally (VPC, subnets)

### Data Stores

| Need | Service | Provisioning |
|------|---------|-------------|
| Relational data | Cloud SQL (PostgreSQL) | `terraform-google-sql-db` module |
| Caching | Memorystore (Redis or Valkey) | `terraform-google-memorystore` module |
| Object storage | Cloud Storage | [`terraform-google-cloud-storage`](https://github.com/entur/terraform-google-cloud-storage) module |
| Event streaming | Apache Kafka (Aiven) | See [kafka.md](kafka.md) |
| Analytics | BigQuery | `google_bigquery_dataset` / `table` resources |

Always use Entur Terraform modules for Cloud SQL and Memorystore -- they handle naming, networking, secrets, and K8s integration.

### Asynchronous Messaging

Entur uses Apache Kafka on Aiven. Use **Avro** (default) or **Protobuf** with Confluent Schema Registry. Keep messages self-contained with correlation IDs for tracing. See [kafka.md](kafka.md) for full documentation.

## Database Design

### PostgreSQL Conventions

- `snake_case` for table and column names
- Singular table names: `route`, `stop_place`
- Always include `id`, `created`, `changed` columns
- Use UUID/ULID for distributed systems or `Long` auto-increment for simpler domains
- Use Flyway for migrations; prefix with version: `V1__create_route_table.sql`
- Migrations must be backward-compatible (rolling deployments)
- ALWAYS treat applied migrations as immutable

For SQL library choice, see language-specific guides: [kotlin.md](kotlin.md), [java.md](java.md).

## Resilience

Use circuit breakers, retry with backoff, and explicit timeouts for all external service calls. Design for graceful degradation -- return cached or partial data when dependencies are unavailable. See [api-design.md](api-design.md#rate-limiting-and-resilience) for details.

## Production Hardening

The common Helm chart handles most of this automatically. See [helm.md](helm.md) for configuration.

| Concern | Requirement |
|---------|-------------|
| Replicas | Production must run >1 pod |
| HPA | Auto-configured; scales on CPU (80%); favor horizontal scaling |
| PDB | `minAvailable: 50%` in prd; `0` in dev/tst with single pod |
| Pod anti-affinity | Distributes across nodes (auto-configured) |
| Zone anti-affinity | Distributes across AZs for critical services |
| VPA | Enabled by default for resource recommendations |
| CPU sizing | Set request, **no limit** (allow bursting) |
| Memory sizing | Set request and limit to **same value** (OOM kill on exceed) |

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
