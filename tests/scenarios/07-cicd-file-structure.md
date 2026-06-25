# Scenario: CI/CD Pipeline File Structure

## Description

Verifies the skill generates the correct set of workflow files with proper naming (`.yml` for workflows) and the right pipeline architecture for a Kotlin project with Helm and Terraform.

## Prompt

You are setting up CI/CD workflows for a new Kotlin Spring Boot API at Entur.

Details:

- Repository name: platform-examples
- Language: Kotlin (Gradle)
- Has Helm chart in helm/platform-examples
- Has Terraform in terraform/
- Environments: dev, tst, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

List ALL workflow files and the dependabot config file that should be created under .github/. For each file, give the exact path and a one-line description of its purpose. Use the format:

- path: <exact file path relative to repo root>
  purpose: <one-line description>

## Assertions

```json
{
  "must_contain": [
    ".github/workflows/ci.yml",
    ".github/workflows/build.yml",
    ".github/workflows/cd.yml",
    ".github/workflows/pr.yml",
    ".github/workflows/codeql.yml",
    ".github/workflows/dependabot-pr.yml",
    ".github/workflows/terraform.yml",
    ".github/workflows/terraform-drift-detection.yml"
  ],
  "must_not_contain": [
    ".github/workflows/ci.yaml",
    ".github/workflows/build.yaml",
    ".github/workflows/cd.yaml",
    ".github/workflows/pr.yaml",
    ".github/workflows/codeql.yaml",
    ".github/workflows/dependabot-pr.yaml",
    ".github/workflows/terraform.yaml",
    ".github/workflows/terraform-drift-detection.yaml",
    "ci-pr.yaml",
    "ci-pr.yml",
    "deploy.yaml",
    "deploy.yml",
    "lint-helm.yaml",
    "lint-helm.yml"
  ],
  "must_match": [
    "dependabot\\.yml"
  ]
}
```

## Budget

0.10
