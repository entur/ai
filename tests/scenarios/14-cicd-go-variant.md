# Scenario: CI/CD for Go Service

## Description

Verifies that ci.yaml for a Go project uses setup-go with go-version-file, runs `go test`, does NOT upload build artifacts or use test-reporter, and the docker build job does NOT reference build_artifact_name or build_artifact_path. Also verifies dependabot uses gomod ecosystem instead of gradle.

## Prompt

You are setting up CI/CD workflows for a Go service at Entur.

Details:

- Repository name: stop-lookup
- Language: Go
- Has go.mod at repo root
- Has Helm chart in helm/stop-lookup
- No Terraform
- Environments: dev, tst, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate:

1. The complete `.github/workflows/ci.yaml`
2. The dependabot configuration file

Output the YAML content for each file, clearly labeled.

## Assertions

```json
{
  "must_contain": [
    "workflow_call:",
    "setup-go@v5",
    "go-version-file: go.mod",
    "go test ./...",
    "docker-lint:",
    "helm-lint:",
    "docker-push:",
    "entur/gha-docker/.github/workflows/build.yml@v1",
    "gomod"
  ],
  "must_not_contain": [
    "setup-java",
    "setup-gradle",
    "gradlew",
    "upload-artifact",
    "build_artifact_name",
    "build_artifact_path",
    "dorny/test-reporter",
    "ARTIFACTORY",
    "liberica",
    "temurin",
    "gradle",
    "pip",
    "setup-python"
  ],
  "must_match": [
    "helm/stop-lookup",
    "dependabot\\[bot\\]"
  ]
}
```

## Budget

0.12
