---
name: setup-cicd-workflows
description: >
  Generate Entur-standard CI/CD GitHub Actions workflows for a project.
  Detects language (Kotlin/Java, Go, Python) and generates all required workflow
  files: ci.yaml, build.yaml, cd.yaml, pr.yaml, codeql.yaml, dependabot-pr.yaml,
  terraform.yaml, terraform-drift-detection.yaml, and dependabot.yml.
  Use this skill when the user says "set up CI/CD", "create pipelines",
  "add GitHub Actions", "configure deployment", or needs CI/CD workflows for a new
  or existing Entur project. Always uses Entur reusable workflows -- never custom steps.
---

# CI/CD Workflow Setup

Generate the complete set of GitHub Actions workflows for an Entur project using Entur reusable workflows. Uses the image promotion model: PRs build and push images, merges resolve and deploy the PR-built image.

## Pipeline Architecture

```text
.github/workflows/
  ci.yaml                        ← Reusable CI (lint, test, Docker build/scan/push)
  build.yaml                     ← PR: calls ci.yaml
  cd.yaml                        ← Deploy: resolve PR-built image, deploy dev → tst → prd
  pr.yaml                        ← PR verification (title/body validation)
  codeql.yaml                    ← Security code scanning
  dependabot-pr.yaml             ← CI for Dependabot PRs after human approval
  terraform.yaml                 ← Terraform lint/plan/apply (if terraform/ exists)
  terraform-drift-detection.yaml ← Weekly Terraform drift check (if terraform/ exists)
.github/
  dependabot.yml                 ← Dependabot dependency update config
```

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
| **Repo name** | from git remote or directory name | ask |
| **Slack channel ID** | ask user (optional) | omit |

## Step 2: Generate `.github/workflows/ci.yaml`

Reusable build workflow called by both `build.yaml` (PRs) and `cd.yaml` (workflow_dispatch without pre-built image).

```yaml
# Continuous Integration Workflow
#
# This is a reusable workflow (workflow_call) -- it is NOT triggered directly by events.
# It is called by build.yaml (on PRs) and cd.yaml (on workflow_dispatch without a pre-built image).
#
# Pipeline:
#   1. Lint (parallel)    -- Validate Dockerfile {and Helm charts for all environments}
#   2. Test              -- Build and run tests, upload artifacts
#   3. Docker build      -- Build Docker image (runs after all lints + tests pass)
#   4. Docker scan/push  -- Scan image for vulnerabilities and push to the container registry (parallel)
#
# The output `image_and_tag` is passed back to the caller (cd.yaml) so it knows
# which image to deploy. The docker push also creates a git tag used later to
# promote the same image to higher environments after merge.

name: CI

on:
  workflow_call:
    outputs:
      image_and_tag:
        description: Fully qualified image reference
        value: ${{ jobs.docker-push.outputs.image_and_tag }}

jobs:
  # --- Linting (runs in parallel with test) ---

  docker-lint:
    uses: entur/gha-docker/.github/workflows/lint.yml@v1
```

### Add Helm lint job (if Helm chart exists):

Helm linting runs inside `ci.yaml` (not as a separate workflow) so charts are validated on every PR and dispatch build, not only when `helm/**` files change.

```yaml
  # Helm charts are linted per environment to catch env-specific value errors
  helm-lint:
    uses: entur/gha-helm/.github/workflows/lint.yml@v1
    strategy:
      matrix:
        environment: [{environments}]       # e.g. [dev, tst, prd]
    with:
      chart: helm/{repoName}
      environment: ${{ matrix.environment }}
      values: values-kub-ent-${{ matrix.environment }}.yaml
```

### Add test job by language:

Replace `{module}` with the Gradle subproject name (e.g. `app`). For root-level builds without subprojects, use the repo root paths directly (e.g. `build/test-results/test/*.xml`).

**Kotlin/Java:**

```yaml
  # --- Build and Test ---

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
      # Publish test results as a GitHub check so they are visible directly on the PR
      - uses: dorny/test-reporter@v2
        if: always()
        with:
          name: test-results
          path: "{module}/build/test-results/test/*.xml"
          reporter: java-junit
      # Upload the build artifact for the Docker build step
      - uses: actions/upload-artifact@v4
        with:
          name: build
          path: "{module}/build/distributions/{module}.tar"
          retention-days: 4
```

**Go:**

```yaml
  # --- Build and Test ---

  test:
    if: github.actor != 'dependabot[bot]'
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      checks: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./...
```

**Python:**

```yaml
  # --- Build and Test ---

  test:
    if: github.actor != 'dependabot[bot]'
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      checks: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version-file: '.python-version'
      - run: pip install -r requirements.txt && pytest
```

### Add Docker build, scan, and push jobs:

```yaml
  # --- Docker Build, Scan, and Push ---
  # Only runs after all lints and tests pass to avoid wasting resources on broken builds
```

**Kotlin/Java** (uses build artifact from test job):

```yaml
  docker:
    needs: [test, docker-lint{, helm-lint}]     # include helm-lint if Helm exists
    uses: entur/gha-docker/.github/workflows/build.yml@v1
    with:
      build_artifact_name: build
      build_artifact_path: "{module}/build/distributions"   # match the upload-artifact path from test job
```

If Artifactory deps detected, add build secrets:

```yaml
    secrets:
      BUILD_SECRETS: |
        "ARTIFACTORY_AUTH_USER=${{ secrets.ARTIFACTORY_AUTH_USER }}"
        "ARTIFACTORY_AUTH_TOKEN=${{ secrets.ARTIFACTORY_AUTH_TOKEN }}"
```

**Go / Python** (no build artifact needed):

```yaml
  docker:
    needs: [test, docker-lint{, helm-lint}]     # include helm-lint if Helm exists
    uses: entur/gha-docker/.github/workflows/build.yml@v1
```

### Scan and push (all languages):

```yaml
  # Scan and push run in parallel -- a vulnerability finding does not block the push,
  # but the scan results are available for review on the PR
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

## Step 3: Generate `.github/workflows/build.yaml`

PR build trigger. Calls the reusable `ci.yaml` on every pull request.

```yaml
# Build Workflow
#
# Runs CI (lint, test, Docker build/scan/push) on every pull request.
# The built image is tagged and pushed to the registry so it can be
# resolved and deployed by cd.yaml after merge.

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

## Step 4: Generate `.github/workflows/cd.yaml`

Continuous deployment using image promotion. On merge to main, resolves the image built during the PR (via git tag) and deploys through dev -> tst -> prd. Also supports manual deployment via workflow_dispatch.

```yaml
# Continuous Deployment Workflow -- GitHub Flow
#
# Deploys to all environments after merge to main, using the image built
# during the PR (resolved via git tag). Also supports manual deployments
# via workflow_dispatch.
#
# Triggers:
#   - push (main): After a PR is merged, resolve the image that was built
#     during the PR (via git tag) and deploy it through dev -> tst -> prd.
#     tst and prd require approval via GitHub Environment protection rules.
#   - workflow_dispatch: Manual deployments via the GitHub Actions UI. Allows
#     targeting a specific environment and optionally providing a pre-built image.
#     If no image is provided, a new one is built first.
#
# Concurrency:
#   - Workflow-level: cancel-in-progress: false ensures a running deploy is never
#     interrupted. GitHub automatically drops older queued runs, keeping only the
#     latest -- so you get at most 1 active + 1 queued run per branch.
#   - Deploy jobs use cancel-in-progress: false (never cancel an in-progress deploy)
#
# Image promotion:
#   The PR workflow builds and pushes a Docker image tagged with
#   {branch}.{date}-SHA{short_sha}. The docker push workflow also creates a git
#   tag with that same value. On merge, the resolve-image job looks up the merged
#   PR's branch name, finds the most recent matching git tag, and outputs the
#   image reference -- so all environments deploy the exact same image that was
#   built during the PR.

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
        options:
          - dev
          - tst
          - prd
      image:
        description: Docker image tag to deploy (image_name:image_tag). Leave empty to build a new image.
        required: false
        type: string

concurrency:
  group: cd-${{ github.ref }}
  cancel-in-progress: false

jobs:
  # ============================================================================
  # BUILD (workflow_dispatch only, when no pre-built image is provided)
  # ============================================================================

  build:
    if: github.event_name == 'workflow_dispatch' && inputs.image == ''
    uses: ./.github/workflows/ci.yaml
    secrets: inherit

  # ============================================================================
  # RESOLVE IMAGE (push to main only)
  # ============================================================================
  # After a PR is merged, find the image that was built during the PR by looking
  # up the PR branch name and finding the most recent git tag created by the
  # docker push workflow for that branch/pr.

  resolve-image:
    if: github.event_name == 'push'
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      pull-requests: read
    outputs:
      image: ${{ steps.resolve.outputs.image }}
      pr_number: ${{ steps.resolve.outputs.pr_number }}
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
          REPO_NAME="${GITHUB_REPOSITORY#*/}"

          # Find the PR that was merged at this commit
          PR_NUMBER=$(gh pr list --search "$GITHUB_SHA" --state merged --json number --jq '.[0].number')
          if [ -z "$PR_NUMBER" ]; then
            echo "::error::Could not find merged PR for commit $GITHUB_SHA"
            exit 1
          fi

          # Get the PR branch name and apply the same sanitisation as gha-docker:
          # lowercase, replace /.- with -, truncate to 43 chars, strip trailing -
          BRANCH=$(gh pr view "$PR_NUMBER" --json headRefName --jq '.headRefName')
          BRANCH=$(echo "$BRANCH" | tr '[:upper:]' '[:lower:]' | tr -d 'ÆØÅæøå')
          BRANCH=${BRANCH//\//-}
          BRANCH=${BRANCH//./-}
          BRANCH=${BRANCH//!/-}
          BRANCH=${BRANCH:0:43}
          BRANCH=${BRANCH%-}

          # Find the most recent git tag for this branch (created by the docker push workflow).
          # Tags follow the pattern: {branch}.{YYYYMMDD}-SHA{short_sha}
          IMAGE_TAG=$(git tag -l "${BRANCH}.*" --sort=-creatordate | head -1)
          if [ -z "$IMAGE_TAG" ]; then
            echo "::warning::No git tag matching '${BRANCH}.*' (PR #${PR_NUMBER}) -- no image was built for this PR, skipping deploy"
            echo "pr_number=${PR_NUMBER}" >> "$GITHUB_OUTPUT"
            exit 0
          fi

          echo "Resolved image: ${REPO_NAME}:${IMAGE_TAG} (from PR #${PR_NUMBER})"
          echo "image=${REPO_NAME}:${IMAGE_TAG}" >> "$GITHUB_OUTPUT"
          echo "pr_number=${PR_NUMBER}" >> "$GITHUB_OUTPUT"

  # ============================================================================
  # DEPLOY
  # ============================================================================
  # Each deploy job lists both build and resolve-image in `needs` so it can
  # receive the image from either path. The `always()` condition ensures the job
  # runs even when one of its dependencies is skipped (which is expected -- build
  # is skipped on push, resolve-image is skipped on dispatch).
  # The image fallback chain is: resolve-image (push) || inputs.image (dispatch) || build (dispatch without pre-built image).
```

### Deploy jobs (for each environment):

Generate deploy jobs for all environments. This example shows the standard three-environment setup:

```yaml
  deploy-dev:
    needs: [build, resolve-image]
    if: >-
      always() && !cancelled() && !contains(needs.*.result, 'failure')
      && (github.event_name == 'push'
          || (github.event_name == 'workflow_dispatch' && (inputs.environment == '' || inputs.environment == 'dev')))
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
    if: >-
      always() && !cancelled() && !contains(needs.*.result, 'failure')
      && (github.event_name == 'push'
          || (github.event_name == 'workflow_dispatch' && (inputs.environment == '' || inputs.environment == 'tst')))
      && (needs.resolve-image.outputs.image != '' || inputs.image != '' || needs.build.outputs.image_and_tag != '')
    uses: entur/gha-helm/.github/workflows/deploy.yml@v1
    concurrency:
      group: cd-deploy-tst
      cancel-in-progress: false
    with:
      environment: tst
      image: ${{ needs.resolve-image.outputs.image || inputs.image || needs.build.outputs.image_and_tag }}
    secrets: inherit

  deploy-prd:
    needs: [build, resolve-image, deploy-tst]
    if: >-
      always() && !cancelled() && !contains(needs.*.result, 'failure')
      && (github.event_name == 'push'
          || (github.event_name == 'workflow_dispatch' && (inputs.environment == '' || inputs.environment == 'prd')))
      && (needs.resolve-image.outputs.image != '' || inputs.image != '' || needs.build.outputs.image_and_tag != '')
    uses: entur/gha-helm/.github/workflows/deploy.yml@v1
    concurrency:
      group: cd-deploy-prd
      cancel-in-progress: false
    with:
      environment: prd
      image: ${{ needs.resolve-image.outputs.image || inputs.image || needs.build.outputs.image_and_tag }}
    secrets: inherit
```

### Add Slack notifications (if Slack channel ID provided):

Add `slack_channel_id: {CHANNEL_ID}` to each deploy job's `with:` block.

### Multi-namespace deploy (if detected):

Replace simple deploy jobs with matrix strategy:

```yaml
  deploy-{env}:
    needs: [build, resolve-image{, deploy-previous-env}]
    if: >-
      always() && !cancelled() && !contains(needs.*.result, 'failure')
      && (github.event_name == 'push'
          || (github.event_name == 'workflow_dispatch' && (inputs.environment == '' || inputs.environment == '{env}')))
      && (needs.resolve-image.outputs.image != '' || inputs.image != '' || needs.build.outputs.image_and_tag != '')
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
    concurrency:
      group: cd-deploy-{env}
      cancel-in-progress: false
    with:
      environment: ${{ matrix.environment }}
      chart: helm/{repoName}
      namespace: ${{ matrix.namespace }}
      release_name: ${{ matrix.release_name }}
      image: ${{ needs.resolve-image.outputs.image || inputs.image || needs.build.outputs.image_and_tag }}
      values: ${{ matrix.values }}
    secrets: inherit
```

### Skip Helm (if no `helm/` directory):

Omit all deploy jobs from `cd.yaml`. The workflow only builds and pushes the Docker image.

## Step 5: Generate `.github/workflows/pr.yaml`

PR verification using the Entur reusable PR verification workflow.

```yaml
# PR Verification Workflow
#
# Validates pull request metadata (title, body, labels, etc.) against Entur conventions.
# This runs on every PR event -- including edits to the title/body -- so that developers
# get immediate feedback if their PR doesn't meet the required format.
#
# The concurrency group ensures that only the latest run per PR is active,
# cancelling any in-progress runs when a new push or edit arrives.

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

## Step 6: Generate `.github/workflows/codeql.yaml`

Security code scanning with GitHub CodeQL.

```yaml
# CodeQL Security Analysis Workflow
#
# Runs GitHub's CodeQL static analysis to detect security vulnerabilities and
# coding errors in the source code.
#
# Triggers:
#   - push to main:     Scan after merge to catch anything missed in PR review
#   - pull_request:     Scan on PRs to give feedback before merge
#   - schedule (weekly): Catch newly discovered vulnerability patterns in existing code
#     (CodeQL's rule database is updated regularly, so new issues can appear
#     even without code changes)

name: CodeQL Analysis

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

For Java/Kotlin projects, add configuration:

```yaml
    with:
      java_version: '25'
      java_distribution: 'liberica'
      use_setup_java: true
      codeql_queries: 'security-extended'
```

## Step 7: Generate `.github/workflows/dependabot-pr.yaml`

Dependabot PRs do not receive repository secrets by default. This workflow adds an approval gate so secrets are only exposed after a human has reviewed and approved the PR.

```yaml
# Dependabot pull request workflow
#
# Dependabot PRs do not receive repository secrets by default.
# This workflow adds an approval gate so secrets are only exposed after
# a human has reviewed and approved the PR.

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

## Step 8: Generate `.github/workflows/terraform.yaml` (if `terraform/` exists)

Manages Terraform infrastructure changes across all environments. On PR: lint, plan all environments, apply dev. On merge: lint, plan tst+prd, apply tst, apply prd.

```yaml
# Terraform Infrastructure Workflow -- apply on PR, promote on merge
#
# Manages infrastructure changes across all environments (dev, tst, prd).
#
# 1. Pull Request (terraform/** changes):
#    Lint -> Plan all environments -> Apply dev only.
#    The developer sees the plan output for all environments on the PR, but only
#    dev is applied automatically. This gives fast feedback without risking
#    higher environments.
#
# 2. Push to main (terraform/** changes):
#    Lint -> Plan tst and prd -> Apply tst -> Apply prd.
#    After a PR with terraform changes is merged, tst and prd are applied with
#    environment approval required for each (via GitHub Environment protection).
#
# Concurrency:
#   - Plan jobs use cancel-in-progress: true (only latest plan matters)
#   - Apply jobs use cancel-in-progress: false (never cancel an in-progress apply)

name: Terraform

on:
  push:
    branches: [main]
    paths:
      - 'terraform/**'
  pull_request:
    branches: [main]
    paths:
      - 'terraform/**'
  workflow_dispatch:

jobs:
  # --- Lint ---
  terraform-lint:
    uses: entur/gha-terraform/.github/workflows/lint.yml@v2
    concurrency:
      group: terraform-lint-${{ github.ref }}
      cancel-in-progress: true

  # ============================================================================
  # PLAN -- all environments
  # ============================================================================

  tf-plan-dev:
    needs: [terraform-lint]
    if: github.event_name != 'push'
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    concurrency:
      group: terraform-plan-dev-${{ github.ref }}
      cancel-in-progress: true
    with:
      environment: dev

  tf-plan-tst:
    needs: [terraform-lint]
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    concurrency:
      group: terraform-plan-tst-${{ github.ref }}
      cancel-in-progress: true
    with:
      environment: tst

  tf-plan-prd:
    needs: [terraform-lint]
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    concurrency:
      group: terraform-plan-prd-${{ github.ref }}
      cancel-in-progress: true
    with:
      environment: prd

  # ============================================================================
  # APPLY -- dev (PR only, automatic)
  # ============================================================================

  tf-apply-dev:
    if: github.event_name == 'pull_request' || github.event_name == 'workflow_dispatch'
    needs: [tf-plan-dev]
    uses: entur/gha-terraform/.github/workflows/apply.yml@v2
    concurrency:
      group: terraform-apply-dev
      cancel-in-progress: false
    with:
      environment: dev
      has_changes: ${{ needs.tf-plan-dev.outputs.has_changes }}

  # ============================================================================
  # APPLY -- tst and prd (push to main only, requires environment approval)
  # ============================================================================

  tf-apply-tst:
    if: github.event_name == 'push' || github.event_name == 'workflow_dispatch'
    needs: [tf-plan-tst]
    uses: entur/gha-terraform/.github/workflows/apply.yml@v2
    concurrency:
      group: terraform-apply-tst
      cancel-in-progress: false
    with:
      environment: tst
      has_changes: ${{ needs.tf-plan-tst.outputs.has_changes }}

  tf-apply-prd:
    if: github.event_name == 'push' || github.event_name == 'workflow_dispatch'
    needs: [tf-plan-prd, tf-apply-tst]
    uses: entur/gha-terraform/.github/workflows/apply.yml@v2
    concurrency:
      group: terraform-apply-prd
      cancel-in-progress: false
    with:
      environment: prd
      has_changes: ${{ needs.tf-plan-prd.outputs.has_changes }}
```

## Step 9: Generate `.github/workflows/terraform-drift-detection.yaml` (if `terraform/` exists)

Weekly scheduled Terraform plan to detect infrastructure drift. Creates a GitHub issue and sends a Slack notification if drift is found.

```yaml
# Terraform Drift Detection -- weekly scheduled check
#
# Runs Terraform plan for all environments on a schedule to detect if real
# infrastructure has diverged from Terraform state. If drift is found, a
# GitHub issue is created and a Slack notification is sent.

name: Terraform Drift Detection

on:
  schedule:
    - cron: '0 10 * * 4'       # Thursdays at 10:00 UTC
  workflow_dispatch:

jobs:
  # --- Plan all environments ---
  tf-plan-dev:
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    with:
      environment: dev

  tf-plan-tst:
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    with:
      environment: tst

  tf-plan-prd:
    uses: entur/gha-terraform/.github/workflows/plan.yml@v2
    with:
      environment: prd

  # --- Drift check ---
  drift-check:
    needs: [tf-plan-dev, tf-plan-tst, tf-plan-prd]
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      issues: write
    outputs:
      has_drift: ${{ steps.check.outputs.has_drift }}
      drifted_envs: ${{ steps.check.outputs.drifted_envs }}
      issue_url: ${{ steps.check.outputs.issue_url }}
    steps:
      - name: Check for infrastructure drift
        id: check
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          DEV_CHANGES: ${{ needs.tf-plan-dev.outputs.has_changes }}
          TST_CHANGES: ${{ needs.tf-plan-tst.outputs.has_changes }}
          PRD_CHANGES: ${{ needs.tf-plan-prd.outputs.has_changes }}
        run: |
          DRIFTED_ENVS=""
          if [ "$DEV_CHANGES" = "true" ]; then DRIFTED_ENVS="$DRIFTED_ENVS dev"; fi
          if [ "$TST_CHANGES" = "true" ]; then DRIFTED_ENVS="$DRIFTED_ENVS tst"; fi
          if [ "$PRD_CHANGES" = "true" ]; then DRIFTED_ENVS="$DRIFTED_ENVS prd"; fi

          if [ -n "$DRIFTED_ENVS" ]; then
            echo "has_drift=true" >> "$GITHUB_OUTPUT"
            echo "drifted_envs=$DRIFTED_ENVS" >> "$GITHUB_OUTPUT"
            echo "Drift detected in:$DRIFTED_ENVS"
            ISSUE_URL=$(gh issue create \
              --title "Terraform drift detected in:$DRIFTED_ENVS" \
              --body "$(cat <<BODY
          ## Terraform Drift Detected

          The weekly scheduled Terraform plan found infrastructure drift in the following environments: **$DRIFTED_ENVS**

          ### Action Required

          Review the [workflow run](${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}) for details and apply changes or update Terraform configuration to match actual infrastructure.

          ### Details

          | Environment | Drift Detected |
          |-------------|---------------|
          | dev | $DEV_CHANGES |
          | tst | $TST_CHANGES |
          | prd | $PRD_CHANGES |
          BODY
          )" \
              --repo "${{ github.repository }}")
            echo "issue_url=$ISSUE_URL" >> "$GITHUB_OUTPUT"
          else
            echo "has_drift=false" >> "$GITHUB_OUTPUT"
            echo "No drift detected in any environment"
          fi
```

### Add Slack notification (if Slack channel ID provided):

```yaml
  # --- Slack notification ---
  slack-notify:
    if: needs.drift-check.outputs.has_drift == 'true'
    needs: [drift-check]
    uses: entur/gha-slack/.github/workflows/post.yml@v3
    with:
      channel_id: {CHANNEL_ID}
      message: "Terraform drift detected in:${{ needs.drift-check.outputs.drifted_envs }}. Issue: ${{ needs.drift-check.outputs.issue_url }}"
      blocks: >-
        [{"type":"section","text":{"type":"mrkdwn","text":"Terraform drift detected in:${{ needs.drift-check.outputs.drifted_envs }}. Issue: ${{ needs.drift-check.outputs.issue_url }}"}},{"type":"divider"}]
    secrets: inherit
```

## Step 10: Generate `.github/dependabot.yml`

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

## Step 11: Generate Optional Lint Workflows

### `lint-api.yaml` (if OpenAPI specs exist):

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

## Step 12: Print Summary

List all generated files and their purpose:

```text
Generated CI/CD workflows:
  .github/workflows/ci.yaml                        - Reusable CI (lint, test, Docker build/scan/push)
  .github/workflows/build.yaml                     - PR build trigger (calls ci.yaml)
  .github/workflows/cd.yaml                        - Deploy: resolve PR-built image, deploy dev -> tst -> prd
  .github/workflows/pr.yaml                        - PR verification (title/body validation)
  .github/workflows/codeql.yaml                    - Security code scanning (CodeQL)
  .github/workflows/dependabot-pr.yaml             - CI for Dependabot PRs after approval
  .github/workflows/terraform.yaml                 - Terraform lint/plan/apply (if applicable)
  .github/workflows/terraform-drift-detection.yaml - Weekly Terraform drift detection (if applicable)
  .github/workflows/lint-api.yaml                  - API spec linting (if applicable)
  .github/dependabot.yml                           - Automated dependency updates
```

## Critical Rules

- **ALWAYS use Entur reusable workflows** for all CI/CD steps
- **Pin workflow versions** to major tags: `@v1`, `@v2`
- **Use `secrets: inherit`** for security scanning and deploy workflows
- **Image promotion model**: PRs build and push images; merges ALWAYS resolve the PR-built image via git tag
- **Terraform ALWAYS runs in a separate workflow** from `cd.yaml`
- **Deploy concurrency ALWAYS uses `cancel-in-progress: false`** to protect running deployments
- **Plan concurrency uses `cancel-in-progress: true`** -- only the latest plan matters
- **Use `has_changes` output** from terraform-plan to skip unnecessary applies
- **Use the conditional pattern** for deploy jobs with multiple dependency paths:

  ```yaml
  if: >-
    always() && !cancelled() && !contains(needs.*.result, 'failure')
  ```

- **Dependabot PRs need approval** before CI runs with secrets (`dependabot-pr.yaml`)
- **Use `paths-ignore`** on build and cd workflows to skip Terraform-only and docs-only changes
- **Use `paths` triggers** on Terraform workflows to only run on infrastructure changes
- **Upload test reports** using `dorny/test-reporter` for PR check visibility (Java/Kotlin)
- **Upload build artifacts** for Docker builds that need compiled output (Java/Kotlin)
