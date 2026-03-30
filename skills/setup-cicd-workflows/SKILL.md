---
name: setup-cicd-workflows
description: >
  Generate Entur-standard CI/CD GitHub Actions workflows for a project.
  Detects language (Kotlin/Java, Go, Python) and generates all required workflow
  files: ci.yml, ci-pr.yml, deploy.yml, codeql.yml, dependabot.yml, and optional
  lint workflows. Use this skill when the user says "set up CI/CD", "create pipelines",
  "add GitHub Actions", "configure deployment", or needs CI/CD workflows for a new
  or existing Entur project. Always uses Entur reusable workflows -- never custom steps.
---

# CI/CD Workflow Setup

Generate the complete set of GitHub Actions workflows for an Entur project using Entur reusable workflows.

## Step 1: Detect Project Configuration

Determine from the project files or ask the user:

| Input | How to detect | Default |
|-------|--------------|---------|
| **Language** | `build.gradle.kts` → Kotlin/Java, `go.mod` → Go, `requirements.txt`/`pyproject.toml` → Python | Kotlin |
| **Has Helm chart** | `helm/` directory exists | yes |
| **Has Terraform** | `terraform/` directory exists | yes |
| **Has OpenAPI specs** | `specs/` directory exists | no |
| **Has Artifactory deps** | `build.gradle.kts` references `entur2.jfrog.io` | no |
| **Multi-namespace deploy** | Multiple `values-kub-ent-*.yaml` files in helm env | no |
| **Environments** | from self-service manifest or ask | `[dev, tst, prd]` |

## Step 2: Generate `.github/workflows/ci.yml`

Reusable build workflow called by both PR and deploy workflows.

```yaml
name: ci
on:
  workflow_call:
    outputs:
      image_and_tag:
        description: Fully qualified image reference
        value: ${{ jobs.docker-push.outputs.image_and_tag }}

jobs:
  docker-lint:
    uses: entur/gha-docker/.github/workflows/lint.yml@v1

  docker-build:
    needs: [docker-lint]
    uses: entur/gha-docker/.github/workflows/build.yml@v1
```

### Add build secrets block (if Artifactory deps detected):

```yaml
    secrets:
      BUILD_SECRETS: |
        "ARTIFACTORY_AUTH_USER=${{ secrets.ARTIFACTORY_AUTH_USER }}"
        "ARTIFACTORY_AUTH_TOKEN=${{ secrets.ARTIFACTORY_AUTH_TOKEN }}"
```

### Add test job by language:

**Kotlin/Java:**

```yaml
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: liberica
          java-version: 25
      - uses: gradle/actions/setup-gradle@v5
      - run: ./gradlew test --no-daemon
        env:
          ARTIFACTORY_AUTH_USER: ${{ secrets.ARTIFACTORY_AUTH_USER }}
          ARTIFACTORY_AUTH_TOKEN: ${{ secrets.ARTIFACTORY_AUTH_TOKEN }}
      - uses: dorny/test-reporter@v2
        if: always()
        with:
          name: test-results
          path: build/test-results/test/*.xml
          reporter: java-junit
```

**Go:**

```yaml
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./...
```

**Python:**

```yaml
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version-file: '.python-version'
      - run: pip install -r requirements.txt && pytest
```

### Add scan and push jobs:

```yaml
  docker-scan:
    if: github.event_name == 'pull_request' || github.ref == 'refs/heads/main'
    needs: [docker-build, test]
    uses: entur/gha-security/.github/workflows/docker-scan.yml@v2
    with:
      image_artifact: ${{ needs.docker-build.outputs.image_artifact }}
    secrets: inherit

  docker-push:
    if: github.ref == 'refs/heads/main'
    needs: [docker-scan]
    uses: entur/gha-docker/.github/workflows/push.yml@v1
    secrets: inherit
```

## Step 3: Generate `.github/workflows/ci-pr.yml`

PR verification with conventional commit validation:

```yaml
name: ci-pr
on:
  pull_request:
    branches: [main]
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  pr-lint:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: read
    steps:
      - uses: amannn/action-semantic-pull-request@v6
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          ignoreLabels: dependencies
          validateSingleCommit: true

  build:
    needs: [pr-lint]
    uses: ./.github/workflows/ci.yml
    secrets: inherit
```

## Step 4: Generate `.github/workflows/deploy.yml`

Deployment pipeline: build → deploy dev → deploy tst → deploy prd.

For each environment, the pattern is:

```yaml
  # Terraform
  terraform-plan-{env}:
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    with:
      environment: {env}

  terraform-apply-{env}:
    needs: [terraform-plan-{env}]
    uses: entur/gha-terraform/.github/workflows/apply.yml@v2
    with:
      environment: {env}
      has_changes: ${{ needs.terraform-plan-{env}.outputs.has_changes }}

  # Helm deploy
  deploy-{env}:
    needs: [docker-push-or-build, terraform-apply-{env}]
    if: ${{ always() && !cancelled() && !contains(needs.*.result, 'failure') }}
    uses: entur/gha-helm/.github/workflows/deploy.yml@v1
    with:
      environment: {env}
      image: ${{ needs.build.outputs.image_and_tag }}
    secrets: inherit
```

### Multi-namespace deploy (if detected):

Use matrix strategy instead of single deploy:

```yaml
  helm-deploy-{env}:
    strategy:
      fail-fast: false
      matrix:
        include:
          - environment: {env}
            namespace: {namespace1}
            release_name: {release1}
            values: values-kub-ent-{env}.yaml
          - environment: {env}
            namespace: {namespace2}
            release_name: {release2}
            values: values-kub-ent-{env}-{variant}.yaml
    uses: entur/gha-helm/.github/workflows/deploy.yml@v1
    with:
      environment: ${{ matrix.environment }}
      chart: helm/{repoName}
      namespace: ${{ matrix.namespace }}
      release_name: ${{ matrix.release_name }}
      image: ${{ needs.build.outputs.image_and_tag }}
      values: ${{ matrix.values }}
    secrets: inherit
```

### Skip Terraform (if no `terraform/` directory):

Omit all terraform-plan and terraform-apply jobs. The deploy jobs depend only on the docker push.

### Skip Helm (if no `helm/` directory):

Omit deploy jobs entirely. Only build and push the Docker image.

## Step 5: Generate `.github/workflows/codeql.yml`

**This file MUST be named exactly `codeql.yml`** -- security tooling depends on it.

```yaml
name: CodeQL
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 6 * * 1'

jobs:
  code-scan:
    uses: entur/gha-security/.github/workflows/code-scan.yml@v2
    secrets: inherit
```

For Java/Kotlin projects, add:

```yaml
    with:
      java_version: "25"
      java_distribution: "temurin"
      use_setup_java: true
      codeql_queries: "security-extended"
```

## Step 6: Generate Optional Lint Workflows

### `lint-helm.yml` (if Helm chart exists):

```yaml
name: lint-helm
on:
  pull_request:
    paths: ['helm/**']
jobs:
  helm-lint:
    uses: entur/gha-helm/.github/workflows/lint.yml@v1
    strategy:
      matrix:
        environment: [{environments}]
    with:
      chart: helm/{repoName}
      environment: ${{ matrix.environment }}
```

### `lint-api.yml` (if OpenAPI specs exist):

```yaml
name: lint-api
on:
  pull_request:
    paths: ['specs/**']
jobs:
  api-lint:
    uses: entur/gha-api/.github/workflows/lint.yml@v5
    if: github.actor != 'dependabot[bot]'
    secrets: inherit
    with:
      spec: specs/*.yaml
```

## Step 7: Generate `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "docker"
    directories: ["/"]
    schedule:
      interval: "weekly"
```

Add language-specific ecosystems:

**Kotlin/Java** -- add:

```yaml
  - package-ecosystem: "gradle"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Go** -- add:

```yaml
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Python** -- add:

```yaml
  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "weekly"
```

## Step 8: Print Summary

List all generated files and their purpose:

```text
Generated CI/CD workflows:
  .github/workflows/ci.yml          - Reusable build (lint, test, scan, push)
  .github/workflows/ci-pr.yml       - PR verification + CI build
  .github/workflows/deploy.yml      - Build + deploy dev → tst → prd
  .github/workflows/codeql.yml      - Security code scanning
  .github/workflows/lint-helm.yml   - Helm chart linting (if applicable)
  .github/workflows/lint-api.yml    - API spec linting (if applicable)
  .github/dependabot.yml            - Automated dependency updates
```

## Critical Rules

- **Always use Entur reusable workflows** -- never write custom CI/CD steps
- **Pin workflow versions** to major tags: `@v1`, `@v2`
- **CodeQL workflow must be named `codeql.yml`** exactly
- **Use `secrets: inherit`** for security scanning and docs workflows
- **Use `has_changes` output** from terraform-plan to skip unnecessary applies
- **Use the conditional pattern** for jobs after terraform-apply:

  ```yaml
  if: ${{ always() && !cancelled() && !contains(needs.*.result, 'failure') }}
  ```

- **Use concurrency groups** with `cancel-in-progress: true` on PR workflows
- **Upload test reports** using `dorny/test-reporter` for PR check visibility
