# Entur Composite Actions

Reference: [entur/gha-meta](https://github.com/entur/gha-meta)

Composite actions for authentication and common tasks. Used internally by reusable workflows but available for custom steps.

## Available Actions

| Action | Purpose |
|--------|---------|
| `entur/gha-meta/.github/actions/cloud-auth` | Authenticate with GCP |
| `entur/gha-meta/.github/actions/k8s-auth` | Authenticate with GKE |
| `entur/gha-meta/.github/actions/docker-auth` | Authenticate with Google Artifact Registry |
| `entur/gha-meta/.github/actions/posthog` | Send workflow analytics event to PostHog |

## Cloud Authentication

```yaml
steps:
  - uses: entur/gha-meta/.github/actions/cloud-auth@v1
    with:
      environment: dev
```

Sets up Workload Identity Federation for keyless authentication.

## Kubernetes Authentication

```yaml
steps:
  - uses: entur/gha-meta/.github/actions/k8s-auth@v1
    with:
      environment: dev
```

## Docker Registry Authentication

```yaml
steps:
  - uses: entur/gha-meta/.github/actions/docker-auth@v1
```

## PostHog Analytics

> **Audience:** Authors of Entur reusable workflows (e.g. `gha-helm`, `gha-docker`). Application teams do not need this action.

Sends a usage analytics event to PostHog whenever a reusable workflow runs. Use this to track adoption and inputs of your reusable workflow.

| Input | Required | Description |
|-------|----------|-------------|
| `api_key` | Yes | PostHog project API key -- use your team's org variable (e.g. `vars.POSTHOG_API_TOKEN_MYTEAM`) |
| `gha_repository` | Yes | Repo that owns the reusable workflow, e.g. `entur/gha-helm` |
| `workflow_name` | Yes | Event name in PostHog (e.g. `deploy`, `lint`) |
| `workflow_inputs` | No | Workflow inputs as a JSON string -- use `toJSON(inputs)`. Inputs with names matching `token`, `secret`, `key`, `password`, `credential`, or `auth` are stripped automatically. |

Events are sent to the Entur EU PostHog instance (`https://eu.i.posthog.com`). The action runs with `continue-on-error: true` so analytics failures never block a deployment.

Each team must create their own PostHog API token and store it as a GitHub org variable before using this action:

1. Create a PostHog API token for your team's project.
2. Add it as a GitHub org variable named `POSTHOG_API_TOKEN_{TEAMNAME}` (e.g. `POSTHOG_API_TOKEN_PLATTFORM` for the platform team).

Then reference it in the action with `vars.POSTHOG_API_TOKEN_{TEAMNAME}`.

```yaml
# Inside your reusable workflow (e.g. gha-helm/.github/workflows/deploy.yml)
jobs:
  deploy:
    runs-on: ubuntu-24.04
    steps:
      - uses: entur/gha-meta/.github/actions/cloud-auth@v1
        with:
          environment: ${{ inputs.environment }}
      # ... deploy steps ...
      - uses: entur/gha-meta/.github/actions/posthog@v1
        with:
          api_key: ${{ vars.POSTHOG_API_TOKEN_MYTEAM }}
          gha_repository: entur/gha-helm
          workflow_name: deploy
          workflow_inputs: ${{ toJSON(inputs) }}
```

## When to Use Directly

Prefer reusable workflows (`gha-docker`, `gha-helm`, `gha-terraform`) for standard operations. Use composite actions directly only when:

- You need custom steps not covered by reusable workflows
- You need auth combined with custom commands in a single job
- You're building a new reusable workflow

Example -- custom kubectl deployment:

```yaml
jobs:
  custom-deploy:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: entur/gha-meta/.github/actions/cloud-auth@v1
        with:
          environment: dev
      - uses: entur/gha-meta/.github/actions/k8s-auth@v1
        with:
          environment: dev
      - run: kubectl apply -f k8s/custom-resource.yaml
```
