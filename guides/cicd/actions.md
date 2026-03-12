# Entur Composite Actions

Reference: [entur/gha-meta](https://github.com/entur/gha-meta)

Composite actions for authentication and common tasks. Used internally by reusable workflows but available for custom steps.

## Available Actions

| Action | Purpose |
|--------|---------|
| `entur/gha-meta/.github/actions/cloud-auth` | Authenticate with GCP |
| `entur/gha-meta/.github/actions/k8s-auth` | Authenticate with GKE |
| `entur/gha-meta/.github/actions/docker-auth` | Authenticate with Google Artifact Registry |

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
