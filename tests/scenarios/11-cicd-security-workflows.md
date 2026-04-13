# Scenario: CodeQL and Dependabot-PR Workflows

## Description

Verifies codeql.yaml has correct schedule, Java configuration, and uses the Entur security reusable workflow. Verifies dependabot-pr.yaml triggers on pull_request_review (not pull_request), requires approval, and only runs for dependabot[bot].

## Prompt

You are setting up CI/CD workflows for a Kotlin project at Entur.

Details:

- Repository name: platform-examples
- Language: Kotlin/Java

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate two files:

1. The complete `.github/workflows/codeql.yaml`
2. The complete `.github/workflows/dependabot-pr.yaml`

Output the YAML content for each file, clearly labeled.

## Assertions

```json
{
  "must_contain": [
    "name: CodeQL Analysis",
    "entur/gha-security/.github/workflows/code-scan.yml@v2",
    "schedule:",
    "0 6 * * 1",
    "codeql_queries",
    "security-extended",
    "use_setup_java: true",
    "name: on-pull_request_review-submitted",
    "pull_request_review:",
    "types: [submitted]",
    "dependabot[bot]",
    "approved",
    "./.github/workflows/ci.yaml"
  ],
  "must_not_contain": [
    "temurin",
    "pull_request_target"
  ],
  "must_match": [
    "java_version.*25",
    "java_distribution.*liberica",
    "review\\.state.*approved.*dependabot|dependabot.*review\\.state.*approved"
  ]
}
```

## Budget

0.12
