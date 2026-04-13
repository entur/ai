# Scenario: Terraform Workflows

## Description

Verifies terraform.yaml is a separate workflow (not mixed into cd.yaml) with correct lint/plan/apply structure: plan all envs, apply dev on PR, apply tst+prd on merge. Verifies terraform-drift-detection.yaml runs weekly with drift-check job that creates GitHub issues.

## Prompt

You are setting up CI/CD workflows for a project at Entur that has Terraform in terraform/.

Details:

- Environments: dev, tst, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate two files:

1. The complete `.github/workflows/terraform.yaml`
2. The complete `.github/workflows/terraform-drift-detection.yaml`

Output the YAML content for each file, clearly labeled.

## Assertions

```json
{
  "must_contain": [
    "name: Terraform",
    "workflow_dispatch",
    "entur/gha-terraform/.github/workflows/lint.yml@v2",
    "entur/gha-terraform/.github/workflows/plan.yml@v2",
    "entur/gha-terraform/.github/workflows/apply.yml@v2",
    "terraform-lint:",
    "tf-plan-dev:",
    "tf-plan-tst:",
    "tf-plan-prd:",
    "tf-apply-dev:",
    "tf-apply-tst:",
    "tf-apply-prd:",
    "has_changes",
    "name: Terraform Drift Detection",
    "drift-check:",
    "has_drift",
    "drifted_envs",
    "gh issue create"
  ],
  "must_not_contain": [
    "deploy-dev",
    "deploy-tst",
    "deploy-prd",
    "gha-helm"
  ],
  "must_match": [
    "paths.*terraform/\\*\\*",
    "0 10 \\* \\* 4",
    "event_name != 'push'"
  ]
}
```

## Budget

0.15
