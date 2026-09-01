# Terraform State Management

Entur uses GCS remote state with one workspace per environment. State is managed by the platform — teams configure the backend and use standard Terraform commands to operate on it.

## Backend Configuration

The state bucket is auto-created by the platform using your self-service `metadata.id`. Configure the backend in `terraform/main.tf` or a dedicated `terraform/backend.tf`:

```hcl
terraform {
  backend "gcs" {
    bucket = "ent-gcs-tfa-<metadata.id>"  # e.g. "ent-gcs-tfa-products"
  }
}
```

The bucket name is always `ent-gcs-tfa-{metadata.id}`. This bucket has versioning enabled — all state versions are retained and can be restored by the platform team.

## Workspaces

Entur uses one workspace per environment (`dev`, `tst`, `prd`). Each workspace stores its state in a separate prefix within the GCS bucket.

### Initial Setup

```bash
terraform init
terraform workspace new dev
terraform workspace new tst
terraform workspace new prd
```

### Daily Use

```bash
# Check current workspace
terraform workspace show

# Switch environment
terraform workspace select tst

# Apply with environment-specific vars
terraform apply -var-file=env/tst.tfvars
```

### Workspace-Conditional Logic

Use `terraform.workspace` to vary configuration by environment:

```hcl
locals {
  is_production = terraform.workspace == "prd"
}

resource "google_sql_database_instance" "main" {
  settings {
    tier              = local.is_production ? "db-custom-2-7680" : "db-f1-micro"
    availability_type = local.is_production ? "REGIONAL" : "ZONAL"
  }
}
```

Prefer passing environment differences through `tfvars` over inline conditionals when possible — it keeps the configuration readable and explicit.

## State Operations

### Inspecting State

```bash
# List all resources in current workspace
terraform state list

# Show full details of a specific resource
terraform state show google_storage_bucket.my_bucket
```

### Moving Resources

Use `terraform state mv` when renaming a resource or moving it into a module without destroying and recreating it:

```bash
terraform state mv google_storage_bucket.old_name google_storage_bucket.new_name

# Moving a root resource into a module
terraform state mv google_storage_bucket.my_bucket module.storage.google_storage_bucket.bucket
```

### Removing Resources from State

Use `terraform state rm` to stop managing a resource without deleting it from GCP:

```bash
terraform state rm google_storage_bucket.my_bucket
```

After removal, the resource continues to exist in GCP but is no longer tracked by Terraform. Re-import it if you want to manage it again.

### Importing Existing Resources

Use `terraform import` to bring an existing GCP resource under Terraform management:

```bash
# Format: terraform import <resource_address> <gcp_resource_id>
terraform import google_storage_bucket.my_bucket ent-products-dev-my-bucket

terraform import google_sql_database_instance.main projects/ent-products-dev/instances/my-instance
```

Add the resource block to your configuration before importing — Terraform will error if the address does not exist in config.

## State Locking

Terraform automatically acquires a lock before writing state. The lock is stored as a file in the GCS bucket alongside the state.

### When Locks Get Stuck

A stuck lock is usually caused by a cancelled or interrupted `terraform apply`. Identify the lock:

```bash
terraform plan
# Output will include the LockID if state is locked
```

Unlock it:

```bash
terraform force-unlock <LockID>
```

Only run `force-unlock` if you are certain no apply is in progress. As a last resort, the lock file can be deleted directly from the GCS bucket — contact `#talk-utviklerplattform` if you are unsure.

## Sensitive Values in State

State files contain all resource attributes, including secrets and passwords created by Terraform. The GCS bucket has:

- **Versioning** enabled — accidental deletions are recoverable
- **IAM-controlled access** — only CI/CD service accounts and authorised principals can read it

Do not store the state bucket name in public documentation or commit credentials that grant bucket access.

## Troubleshooting

| Problem | Action |
|---------|--------|
| State locked, no apply running | `terraform force-unlock <LockID>` |
| State locked by newer Terraform version | Contact `#talk-utviklerplattform` — state can be restored from GCS versioning |
| Resource drifted (changed outside Terraform) | `terraform plan` shows diff; run `terraform apply` to reconcile or `terraform import` to re-sync |
| Resource deleted outside Terraform | `terraform state rm <address>` to remove from state, then re-create via apply |
| Wrong workspace active | `terraform workspace show`; switch with `terraform workspace select <env>` |
| Cannot find bucket | Verify `metadata.id` in `.entur/app-<appid>.yaml` matches the bucket suffix |

## Best Practices

1. **Always verify your workspace** before running `apply` — applying `prd` tfvars in the `dev` workspace targets the wrong project
2. **Use `-var-file`** explicitly; do not rely on default variable values for environment-specific config
3. **Never edit state manually** — use `terraform state mv`, `rm`, or `import` instead
4. **Run `terraform plan` in CI** and `terraform apply` in CD via the `gha-terraform` workflow; see [gha-workflows.md](gha-workflows.md)
5. **Prefer tfvars over workspace conditionals** for environment differences — easier to review and audit
6. **Do not share state buckets** across applications — each app has its own `ent-gcs-tfa-{app_id}` bucket
