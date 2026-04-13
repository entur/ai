# Scenario: CI/CD Without Helm Chart

## Description

Verifies that when a project has no Helm chart, ci.yaml omits the helm-lint job, the docker job does not depend on helm-lint, and cd.yaml omits deploy jobs entirely (only builds and pushes images).

## Prompt

You are setting up CI/CD workflows for a Go service at Entur.

Details:

- Repository name: data-processor
- Language: Go
- NO Helm chart (no helm/ directory)
- NO Terraform
- Environments: dev, tst, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate:

1. The complete `.github/workflows/ci.yaml`
2. The complete `.github/workflows/cd.yaml`

Output the YAML content for each file, clearly labeled.

## Assertions

```json
{
  "must_contain": [
    "workflow_call:",
    "docker-lint:",
    "docker-push:",
    "entur/gha-docker/.github/workflows/build.yml@v1",
    "resolve-image:"
  ],
  "must_not_contain": [
    "helm-lint",
    "gha-helm",
    "values-kub-ent",
    "deploy-dev",
    "deploy-tst",
    "deploy-prd"
  ],
  "must_match": [
    "docker.*needs.*test.*docker-lint"
  ]
}
```

## Budget

0.12
