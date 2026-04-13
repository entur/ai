# Entur Reusable GitHub Actions Workflows

Always use Entur reusable workflows instead of custom pipeline steps.

## Available Workflows

| Repository | Purpose | Version |
|-----------|---------|---------|
| [gha-docker](https://github.com/entur/gha-docker) | Docker lint, build, push | `@v1` |
| [gha-helm](https://github.com/entur/gha-helm) | Helm lint, unit test, deploy | `@v1` |
| [gha-terraform](https://github.com/entur/gha-terraform) | Terraform lint, plan, apply | `@v2` |
| [gha-security](https://github.com/entur/gha-security) | Code scan (CodeQL), Docker scan (Grype) | `@v2` |
| [gha-meta](https://github.com/entur/gha-meta) | Releases, PR verification, auth actions | `@v1` |
| [gha-firebase](https://github.com/entur/gha-firebase) | Firebase Hosting preview and deploy | `@v1` |
| [gha-docs](https://github.com/entur/gha-docs) | Documentation publishing | `@v1` |
| [gha-slack](https://github.com/entur/gha-slack) | Slack notifications | `@v2` |
| [gha-artifactory](https://github.com/entur/gha-artifactory) | Artifactory publishing (Maven/Gradle) | `@v1` |

## Pipeline Architecture

Split CI/CD into focused workflow files. Uses the image promotion model: PRs build and push images, merges resolve the PR-built image via git tag and deploy it.

```text
ci.yaml                        ← Reusable CI (lint, test, Docker build/scan/push) via workflow_call
build.yaml                     ← PR: calls ci.yaml
cd.yaml                        ← Deploy: resolve PR-built image, deploy dev → tst → prd
pr.yaml                        ← PR verification (title/body validation)
codeql.yaml                    ← Security code scanning
dependabot-pr.yaml             ← CI for Dependabot PRs after human approval
terraform.yaml                 ← Terraform lint/plan/apply (if terraform/ exists)
terraform-drift-detection.yaml ← Weekly Terraform drift check (if terraform/ exists)
lint-api.yaml                  ← API spec linting on PR (if specs/ changed)
```

### `has_changes` and Conditional Jobs

When terraform plan reports no changes, apply is skipped. Downstream jobs must use this condition to continue past skipped apply (but still fail on actual failures):

```yaml
if: ${{ always() && !cancelled() && !contains(needs.*.result, 'failure') }}
```

### Go CI Differences

Replace the Java test job in `ci.yaml` with:

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

All other jobs (docker-lint, docker-build, docker-scan, docker-push) are identical.

## Security Scanning (`.github/workflows/codeql.yaml`)

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

Java/Kotlin projects -- add configuration:

```yaml
    with:
      java_version: '25'
      java_distribution: 'liberica'
      use_setup_java: true
      codeql_queries: 'security-extended'
```

## PR Verification

Validates PR titles follow conventional commits:

```yaml
name: Verify PR

on:
  pull_request:
    types: [opened, edited, synchronize, reopened]

jobs:
  verify:
    uses: entur/gha-meta/.github/workflows/verify-pr.yml@v1
```

## Releases (Semantic Versioning)

Automated via release-please and conventional commits:

```yaml
name: Release

on:
  push:
    branches: [main]

jobs:
  release:
    uses: entur/gha-meta/.github/workflows/release.yml@v1
    with:
      release_type: simple      # simple | terraform-module | helm | maven | manifest
    permissions:
      contents: write
      pull-requests: write
      issues: write
```

Release types: `simple` (default), `terraform-module`, `helm`, `maven`, `manifest` (uses `release-please-config.json`).

## Firebase Hosting

### Preview on PR

```yaml
jobs:
  firebase-preview:
    uses: entur/gha-firebase/.github/workflows/firebase-hosting-preview.yml@v1
    with:
      gcp_project_id: my-project-dev
      environment: dev
      build_artifact_name: build-output
      build_artifact_path: build
```

### Deploy to Live

```yaml
jobs:
  firebase-deploy:
    uses: entur/gha-firebase/.github/workflows/firebase-hosting-deploy.yml@v1
    with:
      gcp_project_id: my-project-dev
      environment: dev
      build_artifact_name: build-output
      build_artifact_path: build
```

### Full Firebase Flow with Approval

```yaml
jobs:
  firebase-preview:
    uses: entur/gha-firebase/.github/workflows/firebase-hosting-preview.yml@v1
    with:
      gcp_project_id: my-project-dev
      environment: dev

  approve:
    needs: [firebase-preview]
    runs-on: ubuntu-latest
    environment: apr          # GitHub Environment with protection rules
    steps:
      - run: echo "Approved"

  firebase-deploy:
    needs: [approve]
    uses: entur/gha-firebase/.github/workflows/firebase-hosting-deploy.yml@v1
    with:
      gcp_project_id: my-project-dev
      environment: dev
```

## Documentation Publishing

```yaml
jobs:
  publish-docs:
    uses: entur/gha-docs/.github/workflows/publish.yml@v1
    secrets: inherit
    with:
      project: my-application     # Default: repo name
      directory: docs             # Default: docs
```

## Slack Notifications

Prerequisite: In your Slack channel, run `/invite @GitHub Actions Slack send`

```yaml
jobs:
  notify:
    uses: entur/gha-slack/.github/workflows/post.yml@v2
    with:
      channel_id: "C01ABCDEFGH"
      message: "Deployment to prd completed successfully"
    secrets: inherit
```

## Artifactory Publishing (Maven/Gradle)

Requires `gradle.properties` at repo root with a `version` field in semver format.

```yaml
jobs:
  update-version:
    permissions:
      contents: write
    uses: entur/gha-artifactory/.github/actions/update-version@v1

  maven-publish:
    needs: [update-version]
    uses: entur/gha-artifactory/.github/actions/maven-publish@v1
```

## Workflow Details

Individual workflow steps with optional parameters.

### Docker

```yaml
# Lint (optional: ignore specific hadolint rules)
docker-lint:
  uses: entur/gha-docker/.github/workflows/lint.yml@v1
  with:
    ignore: "DL3008,DL3015"

# Build (outputs: image_artifact)
docker-build:
  uses: entur/gha-docker/.github/workflows/build.yml@v1
  with:
    dockerfile: Dockerfile      # Default
    context: "."                # Default

# Push (outputs: image_name, image_tag, image_and_tag; tag format: branch_name.date-SHA)
docker-push:
  uses: entur/gha-docker/.github/workflows/push.yml@v1
```

### Helm

```yaml
# Lint
helm-lint:
  uses: entur/gha-helm/.github/workflows/lint.yml@v1
  with:
    environment: dev

# Unit test
helm-unittest:
  uses: entur/gha-helm/.github/workflows/unittest.yml@v1

# Deploy
helm-deploy:
  uses: entur/gha-helm/.github/workflows/deploy.yml@v1
  with:
    environment: dev
    image: ${{ needs.docker-push.outputs.image_and_tag }}
```

### Terraform

```yaml
# Lint
terraform-lint:
  uses: entur/gha-terraform/.github/workflows/lint.yml@v2

# Plan (outputs: has_changes, plan_summary)
terraform-plan:
  uses: entur/gha-terraform/.github/workflows/plan.yml@v2
  with:
    environment: dev

# Apply
terraform-apply:
  uses: entur/gha-terraform/.github/workflows/apply.yml@v2
  with:
    environment: dev
    has_changes: ${{ needs.terraform-plan.outputs.has_changes }}
```

## Environments

| Environment | Description |
|-------------|-------------|
| `dev` | Development |
| `tst` | Testing / staging |
| `prd` | Production |

All workflows accept the `environment` input.

## Preferred CI/CD Structure

Split into focused workflow files instead of a monolithic pipeline:

| File | Trigger | Purpose |
|------|---------|---------|
| `ci.yaml` | `workflow_call` | Reusable CI build (lint, test, Docker build/scan/push) |
| `build.yaml` | `pull_request` to main | PR build trigger (calls ci.yaml) |
| `cd.yaml` | Push to main, `workflow_dispatch` | Deploy: resolve PR-built image, deploy dev → tst → prd |
| `pr.yaml` | `pull_request` to main | PR verification (title/body validation) |
| `codeql.yaml` | PR, push to main, schedule | Security code scanning (CodeQL) |
| `dependabot-pr.yaml` | `pull_request_review` submitted | CI for Dependabot PRs after human approval |
| `terraform.yaml` | PR/push changes to `terraform/` | Terraform lint/plan/apply |
| `terraform-drift-detection.yaml` | Weekly schedule, `workflow_dispatch` | Terraform drift detection with issue creation |
| `lint-api.yaml` | PR changes to `specs/` | API spec linting |

### ci.yaml (Reusable Build Workflow)

Called by `build.yaml` (PRs) and `cd.yaml` (workflow_dispatch). Helm linting runs inside this workflow so charts are validated on every build, not only when `helm/**` files change.

```yaml
name: CI
on:
  workflow_call:
    outputs:
      image_and_tag:
        description: Fully qualified image reference
        value: ${{ jobs.docker-push.outputs.image_and_tag }}

jobs:
  docker-lint:
    uses: entur/gha-docker/.github/workflows/lint.yml@v1

  helm-lint:
    uses: entur/gha-helm/.github/workflows/lint.yml@v1
    strategy:
      matrix:
        environment: [dev, tst, prd]
    with:
      chart: helm/{repoName}
      environment: ${{ matrix.environment }}
      values: values-kub-ent-${{ matrix.environment }}.yaml

  test:
    if: github.actor != 'dependabot[bot]'
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      checks: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: liberica
          java-version: "25"
      - uses: gradle/actions/setup-gradle@50e97c2cd7a37755bbfafc9c5b7cafaece252f6e # v6.1.0
        with:
          cache-provider: basic
      - name: Build and test
        run: ./gradlew build
        env:
          ARTIFACTORY_AUTH_USER: ${{ secrets.ARTIFACTORY_AUTH_USER }}
          ARTIFACTORY_AUTH_TOKEN: ${{ secrets.ARTIFACTORY_AUTH_TOKEN }}
      - uses: dorny/test-reporter@v2
        if: always()
        with:
          name: test-results
          path: app/build/test-results/test/*.xml
          reporter: java-junit
      - uses: actions/upload-artifact@v4
        with:
          name: build
          path: app/build/distributions/app.tar
          retention-days: 4

  docker:
    needs: [test, docker-lint, helm-lint]
    uses: entur/gha-docker/.github/workflows/build.yml@v1
    with:
      build_artifact_name: build
      build_artifact_path: app/build/distributions

  docker-scan:
    needs: [docker]
    uses: entur/gha-security/.github/workflows/docker-scan.yml@v2
    secrets: inherit
    with:
      image_artifact: ${{ needs.docker.outputs.image_artifact }}

  docker-push:
    needs: [docker]
    uses: entur/gha-docker/.github/workflows/push.yml@v1
    secrets: inherit
```

### build.yaml (PR Build Trigger)

Runs CI on every pull request. The built image is pushed so it can be resolved by `cd.yaml` after merge.

```yaml
name: Build
on:
  pull_request:
    branches: [main]
    paths-ignore:
      - '**/*.md'
      - 'terraform/**'
concurrency:
  group: build-${{ github.ref }}
  cancel-in-progress: true

jobs:
  build:
    uses: ./.github/workflows/ci.yaml
    secrets: inherit
```

### pr.yaml (PR Verification)

Validates PR metadata against Entur conventions using the reusable verification workflow:

```yaml
name: PR Verification
on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened, edited]
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  verify-pr:
    name: Verify PR
    if: ${{ github.event_name == 'pull_request' }}
    uses: entur/gha-meta/.github/workflows/verify-pr.yml@v1
```

### dependabot-pr.yaml (Dependabot Approval Gate)

Dependabot PRs do not receive repository secrets by default. This workflow runs CI only after a human has approved the PR:

```yaml
name: on-pull_request_review-submitted
on:
  pull_request_review:
    types: [submitted]
permissions:
  contents: read

jobs:
  ci:
    if: github.event.review.state == 'approved' && github.event.pull_request.user.login == 'dependabot[bot]'
    uses: ./.github/workflows/ci.yaml
    secrets: inherit
```

### cd.yaml (Continuous Deployment)

Uses image promotion: on merge to main, resolves the image built during the PR (via git tag) and deploys it. Also supports manual deployment via `workflow_dispatch`.

```yaml
name: Deploy
on:
  push:
    branches: [main]
    paths-ignore:
      - '**/*.md'
      - 'terraform/**'
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        description: Environment to deploy to
        default: dev
        options: [dev, tst, prd]
      image:
        description: Docker image tag (image_name:image_tag). Leave empty to build new.
        required: false
        type: string
concurrency:
  group: cd-${{ github.ref }}
  cancel-in-progress: false

jobs:
  build:
    if: github.event_name == 'workflow_dispatch' && inputs.image == ''
    uses: ./.github/workflows/ci.yaml
    secrets: inherit

  resolve-image:
    if: github.event_name == 'push'
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      pull-requests: read
    outputs:
      image: ${{ steps.resolve.outputs.image }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          fetch-tags: true
      - name: Resolve image from merged PR git tag
        id: resolve
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          # ... resolves PR branch name, finds matching git tag, outputs image reference

  deploy-dev:
    needs: [build, resolve-image]
    if: >-
      always() && !cancelled() && !contains(needs.*.result, 'failure')
      && (needs.resolve-image.outputs.image != '' || inputs.image != '' || needs.build.outputs.image_and_tag != '')
    uses: entur/gha-helm/.github/workflows/deploy.yml@v1
    concurrency:
      group: cd-deploy-dev
      cancel-in-progress: false
    with:
      environment: dev
      image: ${{ needs.resolve-image.outputs.image || inputs.image || needs.build.outputs.image_and_tag }}
    secrets: inherit

  deploy-tst:
    needs: [build, resolve-image, deploy-dev]
    # ... same pattern, depends on deploy-dev

  deploy-prd:
    needs: [build, resolve-image, deploy-tst]
    # ... same pattern, depends on deploy-tst
```

### lint-api.yaml (API Spec Linting)

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

### Dependabot Configuration

```yaml
# .github/dependabot.yml
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
    ignore:
      - dependency-name: "bellsoft/liberica-runtime-container"
        versions: [">=26.0.0"]
  - package-ecosystem: "gradle"
    directory: "/"
    schedule:
      interval: "weekly"
    groups:
      flyway:
        applies-to: version-updates
        patterns: ["org.flywaydb*"]
```

## Best Practices

1. Use `secrets: inherit` for security scanning and docs workflows
2. Pin workflow versions to major tags: `@v1`, `@v2`
3. Use `has_changes` output from terraform-plan to skip unnecessary applies
4. Name the CodeQL workflow `codeql.yaml`
5. Use GitHub Environments with protection rules for deployment approvals
6. Split workflows into focused files -- separate CI, deploy, linting, and security
7. Use matrix strategy for multi-namespace/environment deploys
8. Use concurrency groups with `cancel-in-progress: true` on PR/feature workflows
9. Pass build secrets via `BUILD_SECRETS` for multi-stage Docker builds with Artifactory
10. Use `paths-ignore` on build and cd workflows to skip Terraform-only and docs-only changes
11. Use `paths` triggers on Terraform workflows to only run on infrastructure changes
12. Upload test reports using `dorny/test-reporter` for PR check visibility
13. Upload build artifacts for Docker builds that need compiled output (Java/Kotlin)
14. Image promotion model: PRs build and push images; merges resolve the PR-built image via git tag -- never rebuild on merge
15. Deploy concurrency uses `cancel-in-progress: false` -- never cancel an in-progress deploy
16. Dependabot PRs need human approval before CI runs with secrets (`dependabot-pr.yaml`)
