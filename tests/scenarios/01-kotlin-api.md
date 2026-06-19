# Scenario: Kotlin API Identity Chain

## Description

Verifies the full identity chain for a standard Kotlin Spring Boot API: metadata.id maps to GCP project names, Helm shortname, Terraform app_id, state bucket, and Docker base image.

## Prompt

You are setting up a new Kotlin Spring Boot REST API service at Entur.

Details:

- Repository name: some-repo
- Team: team-reise
- App ID (metadata.id for self-service): someapp
- Environments: dev, tst, prd
- Needs: PostgreSQL, Secret Manager, Auth0 M2M

Read the Entur AI documentation in this repository (start with AGENTS.md) to answer.
Output each answer in `key: value` format on its own line:

- gcp_project_dev: <GCP project ID for dev>
- gcp_project_tst: <GCP project ID for tst>
- gcp_project_prd: <GCP project ID for prd>
- helm_shortname: <Helm shortname value>
- terraform_app_id: <Terraform app_id value>
- terraform_state_bucket: <Terraform state bucket name>
- docker_base_image: <Docker runtime base image>
- secret_manager_project_dev: <SPRING_CLOUD_GCP_SECRETMANAGER_PROJECTID for dev>

## Assertions

```json
{
  "must_contain": [
    "ent-someapp-dev",
    "ent-someapp-tst",
    "ent-someapp-prd",
    "ent-gcs-tfa-someapp"
  ],
  "must_not_contain": [
    "ent-some-repo-dev",
    "ent-someapp-prod",
    "ent-someapp-staging",
    "ent-someapp-test"
  ],
  "must_match": [
    "helm_shortname.*someapp",
    "terraform_app_id.*someapp",
    "secret_manager_project_dev.*ent-someapp-dev",
    "docker_base_image.*gcr\\.io/distroless/java25-debian13:nonroot"
  ]
}
```

## Budget

0.10
