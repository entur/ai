# Deprecate or Delete a Service

Gracefully retire an application from the platform.

## Goal

The application stops serving traffic, releases its resources, and is removed cleanly without leaving orphaned cloud resources or broken consumer integrations.

## Prerequisites

- All consumers of the service have been notified and migrated (or the service has had zero traffic for an agreed window).
- The team has agreed which date the service goes dark.

## Steps (Deprecation -- reversible)

1. **Confirm zero traffic.** Check the Grafana "traffic-pr-service" dashboard for the namespace. See [observability.md](../reference/observability.md#grafana-dashboards).

2. **Scale to zero replicas in GCP Console (soft delete).** This is reversible -- you can scale back up if needed. The GCP projects, Terraform state, and self-service manifest remain.

3. **Remove Apigee proxies.** Ask `#talk-utviklerplattform` to clean up. Internal-only APIs without Apigee skip this.

4. **Request domain name removal** if the service has a custom domain.

## Steps (Deletion -- destructive, mostly irreversible)

1. **Delete the `.entur/` manifests.** Remove `.entur/<appid>.yaml` (and any data-project manifests). Open a PR, run `entur apply`. The Platform Orchestrator schedules GCP project teardown. See [self-service.md](../platform/self-service.md#aborting-or-rolling-back).

2. **Archive the GitHub repository.** Settings → Danger Zone → Archive. This preserves history but prevents pushes.

3. **Clean up container images.** Delete the image repository from Google Artifact Registry once you're sure no rollback is needed.

4. **(Within 30 days) projects can be restored** via `#talk-utviklerplattform`. After 30 days the GCP project deletion becomes permanent and state is lost.

## Verify

- `kubectl get all -n <app>` in each cluster returns no resources.
- The application's GCP project disappears from the GCP Console after the orchestrator finishes its teardown (typically minutes to hours).
- Apigee Developer Portal no longer lists the API spec.
- Repository is shown as **Archived** in GitHub.

## See also

- Self-service mechanics: [self-service.md](../platform/self-service.md)
- Application lifecycle context: [architecture.md](../reference/architecture.md#application-lifecycle)
