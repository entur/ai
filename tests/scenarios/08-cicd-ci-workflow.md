# Scenario: CI Reusable Workflow for Kotlin

## Description

Verifies that ci.yaml is a reusable workflow (workflow_call only) with the correct job structure for a Kotlin project: docker-lint, helm-lint, test (Java 25, liberica, Gradle with SHA-pinned action, test-reporter, upload-artifact), docker build with artifact, docker-scan, docker-push.

## Prompt

You are setting up CI/CD workflows for a Kotlin Spring Boot API at Entur.

Details:

- Repository name: platform-examples
- Module name: app
- Has Helm chart in helm/platform-examples
- Environments: dev, tst, prd
- Uses Artifactory for dependencies

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate the complete `.github/workflows/ci.yaml` file for this project. Output ONLY the YAML content.

## Assertions

```json
{
  "must_contain": [
    "workflow_call:",
    "image_and_tag",
    "docker-lint:",
    "helm-lint:",
    "docker-push:",
    "docker-scan:",
    "entur/gha-docker/.github/workflows/lint.yml@v1",
    "entur/gha-helm/.github/workflows/lint.yml@v1",
    "entur/gha-docker/.github/workflows/build.yml@v1",
    "entur/gha-docker/.github/workflows/push.yml@v1",
    "entur/gha-security/.github/workflows/docker-scan.yml@v2",
    "dorny/test-reporter",
    "upload-artifact",
    "build_artifact_name: build",
    "ARTIFACTORY_AUTH_USER",
    "ARTIFACTORY_AUTH_TOKEN"
  ],
  "must_not_contain": [
    "pull_request:",
    "push:",
    "temurin",
    "setup-gradle@v4",
    "setup-gradle@v5"
  ],
  "must_match": [
    "distribution.*liberica",
    "java-version.*25",
    "setup-gradle@50e97c2cd7a37755bbfafc9c5b7cafaece252f6e",
    "build_artifact_path.*app/build/distributions",
    "app/build/test-results/test",
    "app/build/distributions/app\\.tar",
    "helm/platform-examples",
    "values-kub-ent-.*environment",
    "dependabot\\[bot\\]"
  ]
}
```

## Budget

0.15
