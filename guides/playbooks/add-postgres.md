# Add Managed PostgreSQL

Provision a Cloud SQL PostgreSQL database and wire it into a Spring Boot or Go service.

## Goal

A managed PostgreSQL instance, accessed via Cloud SQL proxy sidecar, with credentials injected as env vars. Flyway-managed schema. Backwards-compatible migrations.

## Prerequisites

- The application is already on the platform (otherwise start with [bootstrap-service.md](bootstrap-service.md)).
- The GCP project `ent-{appid}-{env}` exists for every environment you intend to deploy to.

## Steps

1. **Provision the instance via the Entur Terraform module.** Use `terraform-google-sql-db//modules/postgresql?ref=v1`. The module creates the instance, application user, `PG_USER`/`PG_PASSWORD` secrets in Secret Manager, and Kubernetes resources. Default sizing is environment-aware (zonal in dev/tst, regional HA in prd). See [terraform-modules.md](../platform/terraform-modules.md#cloud-sql-postgresql).

2. **Enable the Cloud SQL proxy sidecar in Helm.** Set `common.postgres.enabled: true` in `values.yaml`. The application connects to `localhost:5432`. Do **not** add `PG_USER`/`PG_PASSWORD` to `common.secrets` -- they are injected automatically. See [common-helm.md](../platform/common-helm.md#database-cloud-sql-proxy).

3. **Configure the application datasource.** Spring Boot: set `spring.datasource.url`, `username`, `password` from the injected env vars. See [java.md](../reference/java.md#cloud-sql-connectivity). Go: use `database/sql` with the `pgx` driver and read the same env vars.

4. **Add Flyway migrations.** Place SQL files under `src/main/resources/db/migration/` (Spring Boot) or your Go migration directory. Name as `V1__create_route_table.sql`. Migrations are immutable once applied. See [architecture.md](../reference/architecture.md#database-design) for naming conventions and rolling-deploy rules.

5. **Size the connection pool.** Total connections = `pods × max_pool_size`. Verify against Cloud SQL `max_connections` (minus 3 reserved) at peak HPA pod count. See [java.md](../reference/java.md#connection-pool-sizing).

6. **(Optional) IAM authentication.** For password-less access from operators, enable `cloudsql.iam_authentication` and add `google_sql_user` resources at the group level. See [terraform-modules.md](../platform/terraform-modules.md#postgresql-iam-authentication).

## Verify

- `terraform plan` shows the instance, database, user, secrets, and ConfigMap creation.
- After deploy, the pod logs show successful Flyway migration on startup.
- Readiness probe stays green after pod start (verify with `kubectl get pod` in the namespace).
- Connect locally for inspection: `gcloud sql connect <instance>` or via the Cloud SQL proxy on your dev machine.

## See also

- Secrets pattern: [security.md](../reference/security.md#secret-management)
- Helm chart reference: [common-helm.md](../platform/common-helm.md)
- Approved IAM roles: [iam-roles.md](../platform/iam-roles.md)
