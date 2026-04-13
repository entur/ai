# Scenario: CI/CD for Python Service

## Description

Verifies that ci.yaml for a Python project uses setup-python with python-version-file, runs pytest, does NOT use Java/Gradle/Go patterns, and dependabot uses pip ecosystem.

## Prompt

You are setting up CI/CD workflows for a Python service at Entur.

Details:

- Repository name: data-enrichment
- Language: Python
- Has requirements.txt and .python-version file
- Has Helm chart in helm/data-enrichment
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
    "setup-python@v5",
    ".python-version",
    "pytest",
    "docker-lint:",
    "helm-lint:",
    "docker-push:",
    "entur/gha-docker/.github/workflows/build.yml@v1",
    "pip"
  ],
  "must_not_contain": [
    "setup-java",
    "setup-gradle",
    "gradlew",
    "setup-go",
    "go test",
    "go-version-file",
    "upload-artifact",
    "build_artifact_name",
    "build_artifact_path",
    "dorny/test-reporter",
    "ARTIFACTORY",
    "liberica",
    "temurin",
    "gradle",
    "gomod"
  ],
  "must_match": [
    "helm/data-enrichment",
    "package-ecosystem.*pip",
    "dependabot\\[bot\\]"
  ]
}
```

## Budget

0.12
