# Bootstrap a New Service

End-to-end playbook for standing up a brand-new application on the Entur platform: GCP projects, container image, Helm release, CI/CD.

## Goal

Get a minimal "hello world" service running in `dev` with the standard golden-path layout, deployed via the common Helm chart and a reusable CI/CD pipeline.

## Prerequisites

- A GitHub repository exists in the `entur` org. Repository name = application name (lowercase kebab-case, max 30 chars).
- Your team has read the [DevOps Handbook](https://enturnett.atlassian.net/wiki/spaces/ESP/overview) Plan section.
- You can list the **App ID** you want (3--10 lowercase alphanumeric chars, unique across Entur). See [self-service.md](../platform/self-service.md#gcp-project-naming) for the identity chain.

## Steps

1. **Provision GCP projects via self-service.** Create `.entur/cicd.yaml` (`GitHubActions`) and `.entur/<appid>.yaml` (`GoogleCloudApplication`). Open a PR, comment `entur apply`, merge. See [self-service.md](../platform/self-service.md) for manifest fields and the apply flow.

2. **Lay out the repository per the golden path.** Repository name = application name = Docker image name = Kubernetes namespace = Helm release name. See [CONVENTIONS.md](../../CONVENTIONS.md#project-structure) for the full directory layout.

3. **Add a Dockerfile.** Multi-stage build, non-root user, pinned base image, port 8080, `/actuator/health/liveness` and `/actuator/health/readiness` (or equivalent) endpoints. See [docker.md](../reference/docker.md) for examples per language and [java.md](../reference/java.md) / [kotlin.md](../reference/kotlin.md) / [go.md](../reference/go.md) for language specifics.

4. **Add the Helm chart.** `helm/<repo-name>/` with `Chart.yaml` depending on the Entur `common` chart, `values.yaml`, and `env/values-kub-ent-{dev,tst,prd}.yaml`. Set `common.shortname` to your **App ID** (must match `metadata.id` from step 1). See [common-helm.md](../platform/common-helm.md) for required values.

5. **Add CI/CD workflows.** Split into focused files: `ci.yaml` (reusable build), `build.yaml` (PR), `cd.yaml` (deploy on merge), `pr.yaml`, `codeql.yaml`, `dependabot-pr.yaml`. Use Entur reusable workflows for every step. See [gha-workflows.md](../platform/gha-workflows.md) for the canonical templates.

6. **Add `AGENTS.md` and a symlink for Claude.** Tell agents where the Entur standards live so they generate platform-compliant code. See [README.md](../../README.md#quick-start) for the recommended template, then `ln -s AGENTS.md CLAUDE.md`.

7. **Add `CODEOWNERS`, `.github/dependabot.yml`, and a PR template.** See [CONVENTIONS.md](../../CONVENTIONS.md) for the standard contents.

## Verify

- `helm lint helm/<repo-name>/ -f helm/<repo-name>/env/values-kub-ent-dev.yaml` passes locally.
- After merging to `main`, the `cd.yaml` workflow deploys a pod to the `dev` namespace.
- `kubectl get pods -n <repo-name>` in `kub-ent-dev` shows the pod `Running` with both probes passing.
- The application appears in `grafana.entur.org` dashboards.

## See also

- Add a database: [add-postgres.md](add-postgres.md)
- Add authorization: [set-up-auth.md](set-up-auth.md)
- Promote to prd: [deploy-to-prd.md](deploy-to-prd.md)
- Local dev loop: [local-dev.md](local-dev.md)
- Pre-built bootstrap skill: [`entur-project-bootstrap`](../../skills/entur-project-bootstrap/SKILL.md)
