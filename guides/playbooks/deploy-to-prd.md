# Promote a Service to Production

Promote an application from `dev`/`tst` to `prd` using the image-promotion deployment model.

## Goal

The same Docker image that was built and tested in `dev` is deployed to `tst` and `prd`. Production hardening is in place (replicas, PDB, HPA, anti-affinity, resource limits). Observability is wired up.

## Prerequisites

- The service is running in `dev` and `tst`.
- Self-service manifest lists `prd` in `spec.environments` and has been applied. See [self-service.md](../platform/self-service.md).
- The GitHub `prd` environment has protection rules (required reviewers) configured.

## Steps

1. **Verify the production environment is provisioned.** Confirm `ent-{appid}-prd` exists in GCP and the Terraform state bucket has a `prd` workspace.

2. **Set production Helm values.** In `helm/<app>/env/values-kub-ent-prd.yaml`: production hostname, larger HPA `maxReplicas`, `pdb.minAvailable: 50%` (or higher), production log level, production Spring profile. See [common-helm.md](../platform/common-helm.md#multi-namespace-deployments).

3. **Confirm production hardening.** See [architecture.md](../reference/architecture.md#production-hardening) for the full table; the common chart enforces most of it automatically:
   - `replicas > 1` (multi-pod for availability)
   - HPA scales on CPU 80% (auto-enabled when replicas > 1)
   - PDB `minAvailable: 50%` (watch the percentage-rounding gotcha)
   - Pod and zone anti-affinity for distribution across nodes/AZs
   - CPU request set, **no CPU limit** (compressible -- allow bursting)
   - Memory request **equal to** memory limit (incompressible -- OOM kill protection)
   - VPA enabled (recommendations stabilize over weeks)

4. **Confirm CD pipeline uses image promotion.** `cd.yaml` on push to `main` resolves the PR-built image via git tag and deploys `dev → tst → prd` with `cancel-in-progress: false` on the deploy concurrency group. Manual `workflow_dispatch` for re-deploying specific tags. See [gha-workflows.md](../platform/gha-workflows.md#cdyaml-continuous-deployment).

5. **Confirm observability is wired.** Liveness verifies the process only; readiness only checks **private** dependencies. Prometheus endpoint enabled. Tracing sampling lowered for prd (often 0.1--1%). See [observability.md](../reference/observability.md).

6. **Confirm secrets and IAM are environment-scoped.** All prd secrets live in `ent-{appid}-prd`'s Secret Manager. IAM roles granted at the group level, never to individual users. See [security.md](../reference/security.md).

7. **Confirm rollback works.** Versions must be backward-compatible with the previous version, and Flyway migrations must allow rolling deploys. See [architecture.md](../reference/architecture.md#microservice-principles).

## Verify

- After merging to `main`, the CD pipeline progresses through `deploy-dev` → `deploy-tst` → `deploy-prd` with the approval gate (if configured) between stages.
- Production pods are `Running` across multiple nodes and AZs (`kubectl get pods -o wide -n <app>`).
- `grafana.entur.org` dashboards show traffic, latency, and saturation.
- VPA dashboard begins producing recommendations after ~1--2 weeks.

## See also

- Bootstrap a new service: [bootstrap-service.md](bootstrap-service.md)
- Workflow templates: [gha-workflows.md](../platform/gha-workflows.md)
- Helm chart reference: [common-helm.md](../platform/common-helm.md)
