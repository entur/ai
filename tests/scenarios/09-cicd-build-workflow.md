# Scenario: Build and PR Workflows

## Description

Verifies that build.yaml is a lightweight PR trigger that calls ci.yaml with correct concurrency and path filtering, and that pr.yaml uses the Entur PR verification reusable workflow with proper event types.

## Prompt

You are setting up CI/CD workflows for a Kotlin project at Entur.

Details:

- Repository name: platform-examples
- Has Terraform in terraform/

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate two files:

1. The complete `.github/workflows/build.yaml`
2. The complete `.github/workflows/pr.yaml`

Output the YAML content for each file, clearly labeled.

## Assertions

```json
{
  "must_contain": [
    "name: Build",
    "pull_request:",
    "branches: [main]",
    "paths-ignore:",
    "'**/*.md'",
    "'terraform/**'",
    "cancel-in-progress: true",
    "./.github/workflows/ci.yaml",
    "secrets: inherit",
    "name: PR Verification",
    "entur/gha-meta/.github/workflows/verify-pr.yml@v1",
    "opened, synchronize, reopened, edited"
  ],
  "must_not_contain": [
    "workflow_dispatch"
  ],
  "must_match": [
    "build-\\$\\{\\{.*github\\.ref",
    "github\\.workflow.*github\\.ref",
    "Verify PR"
  ]
}
```

## Budget

0.12
