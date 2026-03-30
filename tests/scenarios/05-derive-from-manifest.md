# Scenario: Derive Config from Existing Manifest

## Description

The core confusion test: given a self-service manifest, can the agent correctly distinguish metadata.id from metadata.name and derive the right Helm shortname, Terraform app_id, and GCP project IDs?

## Prompt

An Entur team has the following self-service manifest in `.entur/products.yaml`:

```yaml
apiVersion: orchestrator.entur.io/apps/v1
kind: GoogleCloudApplication
metadata:
  id: products
  displayName: "Products API"
  name: products-api
  owner: team-produkt
spec:
  environments: [dev, tst, prd]
  repositories: [products-api]
  terraform:
    createBackend: true
  secretManager:
    enabled: true
    serviceAccount: application
```

Read the Entur AI documentation in this repository (start with AGENTS.md, then helm.md and terraform/modules.md) to answer.
Output each answer in `key: value` format on its own line:

- gcp_project_dev: <GCP project ID for dev>
- gcp_project_tst: <GCP project ID for tst>
- gcp_project_prd: <GCP project ID for prd>
- helm_shortname: <Helm shortname value>
- helm_app: <Helm app value>
- terraform_app_id: <Terraform app_id value>
- terraform_state_bucket: <Terraform state bucket name>
- k8s_namespace: <Kubernetes namespace>

## Assertions

```json
{
  "must_contain": [
    "ent-products-dev",
    "ent-products-tst",
    "ent-products-prd",
    "ent-gcs-tfa-products"
  ],
  "must_not_contain": [
    "ent-products-api-dev",
    "ent-products-api-prd",
    "ent-gcs-tfa-products-api"
  ],
  "must_match": [
    "helm_shortname[:\\s]+products\\b",
    "terraform_app_id[:\\s]+products\\b",
    "helm_app.*products-api",
    "k8s_namespace.*products-api"
  ]
}
```

## Budget

0.10
