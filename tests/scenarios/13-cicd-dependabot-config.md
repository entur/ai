# Scenario: Dependabot Configuration

## Description

Verifies the dependabot config includes the correct package ecosystems for a Kotlin project: github-actions, docker, and gradle. Must not include ecosystems for languages not in use.

## Prompt

You are setting up CI/CD workflows for a Kotlin Spring Boot API at Entur.

Details:

- Repository name: platform-examples
- Language: Kotlin (Gradle build system)
- Has Dockerfile at repo root

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate the dependabot configuration file. Output ONLY the YAML content.

## Assertions

```json
{
  "must_contain": [
    "version: 2",
    "github-actions",
    "docker",
    "gradle",
    "weekly"
  ],
  "must_not_contain": [
    "gomod",
    "pip",
    "npm",
    "cargo",
    "nuget"
  ],
  "must_match": [
    "package-ecosystem.*github-actions",
    "package-ecosystem.*docker",
    "package-ecosystem.*gradle"
  ]
}
```

## Budget

0.10
