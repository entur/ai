# Entur Terraform Modules

> **GCP projects are not created via Terraform.** Terraform manages resources _within_ an existing GCP project. To provision a new GCP project, use the self-service YAML manifests in `.entur/` -- see [self-service.md](../self-service.md). For help, ask in `#talk-utviklerplattform`.

Always use Entur shared modules instead of raw `google_*` resources for managed services.

## Modules

| Module | Purpose | Source |
|--------|---------|--------|
| [terraform-google-init](https://github.com/entur/terraform-google-init) | Platform and app discovery (required by other modules) | `github.com/entur/terraform-google-init//modules/init?ref=v1` |
| [terraform-google-sql-db](https://github.com/entur/terraform-google-sql-db) | Cloud SQL (PostgreSQL) | `github.com/entur/terraform-google-sql-db//modules/postgresql?ref=v1` |
| [terraform-google-memorystore](https://github.com/entur/terraform-google-memorystore) | Memorystore (Redis) | `github.com/entur/terraform-google-memorystore//modules/redis?ref=v2` |
| [terraform-google-cloud-storage](https://github.com/entur/terraform-google-cloud-storage) | Cloud Storage buckets | `github.com/entur/terraform-google-cloud-storage//modules/bucket?ref=v0.2.2` |

## Directory Structure

```text
terraform/
  main.tf             # Module declarations
  variables.tf        # Input variables
  outputs.tf          # Output values
  providers.tf        # Provider configuration
  env/
    dev.tfvars         # Dev environment values
    tst.tfvars         # Test environment values
    prd.tfvars         # Production environment values
```

## Init Module (Required)

Data-only module that discovers GCP platform and application attributes. All other Entur modules depend on it.

```hcl
module "init" {
  source      = "github.com/entur/terraform-google-init//modules/init?ref=v1"
  app_id      = var.app_id
  environment = var.environment
}
```

Standard variables:

```hcl
variable "app_id" {
  description = "Application ID"
  type        = string
}

variable "environment" {
  description = "Environment: dev, tst, or prd"
  type        = string
}
```

```hcl
# env/dev.tfvars
app_id      = "my-application"
environment = "dev"
```

### Init Outputs

| Output | Description |
|--------|-------------|
| `module.init.app.id` | Application ID |
| `module.init.app.name` | Application name |
| `module.init.app.project_id` | GCP project ID |
| `module.init.app.project_number` | GCP project number |
| `module.init.app.owner` | Application owner |
| `module.init.environment` | Environment descriptor |
| `module.init.is_production` | Boolean: is this production? |
| `module.init.kubernetes.project_id` | GKE project ID |
| `module.init.kubernetes.namespace` | Kubernetes namespace |
| `module.init.labels` | Standard labels for all resources |
| `module.init.networks.project_id` | Network project ID |
| `module.init.networks.vpc_name` | VPC name |
| `module.init.networks.vpc_id` | VPC ID |
| `module.init.service_accounts` | Application and project service accounts |

## Cloud SQL (PostgreSQL)

Provisions Cloud SQL PostgreSQL with automatic secret and Kubernetes resource creation.

```hcl
module "postgresql" {
  source    = "github.com/entur/terraform-google-sql-db//modules/postgresql?ref=v1"
  init      = module.init
  databases = ["mydb"]
}
```

### Key Configuration

```hcl
module "postgresql" {
  source    = "github.com/entur/terraform-google-sql-db//modules/postgresql?ref=v1"
  init      = module.init
  databases = ["mydb"]

  # Machine sizing (optional, has sensible defaults per environment)
  machine_size = {
    cpu    = 1
    memory = 3840
  }

  database_version = "POSTGRES_16"    # Default: POSTGRES_14
  region           = "europe-west1"   # Default

  # Backup (defaults shown)
  enable_backup                    = true
  backup_start_time                = "00:00"
  point_in_time_recovery_enabled   = true

  # Storage
  disk_size             = 10          # Initial size in GB
  disk_autoresize       = true
  disk_autoresize_limit = null        # Auto: 500 (prd), 50 (non-prd)

  # Additional users (optional)
  additional_users = {
    readonly = {}
  }

  # Query performance (optional)
  query_insights_enabled = true
}
```

### Default Machine Sizing

| Environment | vCPU | Memory | High Availability |
|-------------|------|--------|-------------------|
| Non-production (dev, tst) | Shared | 600 MB | No (ZONAL) |
| Production (prd) | 1 dedicated | 3840 MB | Yes (REGIONAL) |

### What Gets Created

- Cloud SQL PostgreSQL instance + database(s)
- Application user with password
- Secret Manager secrets: `PG_USER`, `PG_PASSWORD`
- Kubernetes ConfigMap (connection info) + Secret (credentials)

### Application Configuration (Spring Boot)

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/${DB_NAME}
    username: ${PG_USER}
    password: ${PG_PASSWORD}
```

The Cloud SQL proxy sidecar (enabled via `postgres.enabled: true` in Helm) handles connectivity. The application connects to `localhost:5432`.

## Memorystore (Redis)

Provisions Redis with automatic secret and Kubernetes resource creation.

```hcl
module "redis" {
  source = "github.com/entur/terraform-google-memorystore//modules/redis?ref=v2"
  init   = module.init
}
```

### Redis Configuration

```hcl
module "redis" {
  source = "github.com/entur/terraform-google-memorystore//modules/redis?ref=v2"
  init   = module.init

  memory_size_gb    = 1             # Default: 1
  redis_version     = "REDIS_7_0"  # Default
  region            = "europe-west1"

  # High availability
  availability_type = "REGIONAL"    # Default (or ZONAL)
  enable_replicas   = false         # Default
  replica_count     = 1             # 1-5 if replicas enabled

  # Redis configuration
  redis_configs = {
    activedefrag     = "yes"
    maxmemory-policy = "allkeys-lfu"
  }

  maintenance_window = {
    day  = "TUESDAY"
    hour = 0
  }
}
```

### What Gets Created (Redis)

- Memorystore Redis instance
- Secret Manager secrets: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- Kubernetes ConfigMap and Secret with connection info

### Redis Application Configuration (Spring Boot)

```yaml
spring:
  data:
    redis:
      host: ${REDIS_HOST}
      port: ${REDIS_PORT}
      password: ${REDIS_PASSWORD}
```

### Helm Configuration for Redis Secrets

```yaml
# helm/<app>/values.yaml
common:
  secrets:
    redis-credentials:
      - REDIS_HOST
      - REDIS_PORT
      - REDIS_PASSWORD
```

## Cloud Storage

Buckets are private by default. Avoid making them public unless absolutely necessary.

```hcl
module "cloud-storage" {
  source = "github.com/entur/terraform-google-cloud-storage//modules/bucket?ref=v0.2.2"
  init   = module.init
}

resource "google_storage_bucket_iam_member" "reader" {
  bucket = module.cloud-storage.cloud_storage_bucket.name
  role   = "roles/storage.objectViewer"
  member = "user:example-user@entur.org"
}
```

## PostgreSQL IAM Authentication

Enable IAM authentication to avoid password-based access:

```hcl
module "postgresql" {
  source    = "github.com/entur/terraform-google-sql-db//modules/postgresql?ref=v1"
  init      = module.init
  databases = ["my-database"]
  database_flags = {
    iam_authentication = { name = "cloudsql.iam_authentication", value = "on" }
  }
}

# IAM user:
resource "google_sql_user" "iam_user" {
  name     = "example-user@entur.org"
  instance = module.postgresql.instance.name
  type     = "CLOUD_IAM_USER"
}

# IAM group:
resource "google_sql_user" "iam_group" {
  name     = "sg-dig-team-name@entur.org"
  instance = module.postgresql.instance.name
  type     = "CLOUD_IAM_GROUP"
}
```

## BigQuery

```hcl
resource "google_bigquery_dataset" "example" {
  dataset_id = "example_dataset"
  project    = module.init.app.project_id
  location   = "EU"
  labels     = module.init.labels
}

resource "google_bigquery_dataset_iam_member" "viewer" {
  dataset_id = google_bigquery_dataset.example.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = "user:example-user@entur.org"
}
```

Grant IAM at the group level when possible. Be cautious with project-level `roles/bigquery.user` -- it allows creating new datasets.

## Workspaces

Entur uses workspaces to manage environments with a single configuration. Each workspace has its own state file.

```bash
terraform init
terraform workspace new dev
terraform apply -var-file=env/dev.tfvars

# Switch environment:
terraform workspace select tst
terraform apply -var-file=env/tst.tfvars
```

Workspace-specific logic:

```hcl
module "postgresql" {
  source       = "github.com/entur/terraform-google-sql-db//modules/postgresql?ref=v1"
  init         = module.init
  databases    = ["my-database"]
  machine_size = {
    tier = terraform.workspace == "prd" ? "db-custom-1-3840" : "db-f1-micro"
  }
}
```

## Backend State

State stored in GCS bucket auto-created by the platform (naming: `ent-gcs-tfa-<appId>`):

```hcl
terraform {
  backend "gcs" {
    bucket = "ent-gcs-tfa-<appId>"
  }
}
```

## Troubleshooting

- **State locked to newer version**: Contact `#talk-utviklerplattform`. GCS bucket has versioning; state can be restored.
- **State locked, no deployments running**: Caused by cancelled `terraform apply`. Use `terraform force-unlock`. Last resort: manually delete lock file from GCS bucket.
- **Resource not created or manually deleted**: `terraform state list` → `terraform state rm <id>` → re-run. For externally created resources, use `terraform import`.
- **Cannot find resource in cloud project**: Ensure tfvars are in `./terraform/env/`. Verify cluster and workspace:

```bash
gcloud container clusters get-credentials <cluster-name> --region europe-west1 --project <gcp-project-id>
terraform workspace select dev
```

## Best Practices

1. **Pin module versions** with `?ref=TAG` -- never use unversioned references
2. **Always use the init module** as base for all other modules and resources
3. **Use `module.init.labels`** on all resources for consistent labeling
4. **Use `module.init.app.project_id`** for the `project` field on all resources
5. **Use environment-specific tfvars** in `terraform/env/`
6. **Use GCS backend** for remote state (configured by platform)
7. **Only use IAM roles from the [approved list](iam-roles.md)**
8. **Use Entur modules** for Cloud SQL and Memorystore -- no raw `google_sql_*` or `google_redis_*`
9. **Run `terraform plan` in CI** and `terraform apply` in CD via `gha-terraform` workflows
