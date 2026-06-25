# Run a Service on Cloud Run

Off-golden-path runtime for stateless services that do not need Kubernetes. The golden path is still GKE + the [common Helm chart](../platform/common-helm.md) — see [bootstrap-service.md](bootstrap-service.md) first. Pick this playbook when the service is stateless, request-bound, and the team accepts that Cloud Run has no in-cluster service mesh, no session affinity, and a multi-second cold start unless `minScale >= 1` is set.

## Goal

A new `ent-<appid>-prd` project hosts a Cloud Run service deployed by GitHub Actions from `cloudrun.yaml`, fronted by a global HTTPS load balancer with a Google-managed TLS cert. The public hostname is wired up by team-plattform against the load balancer's static IP.

## Prerequisites

- A GitHub repository in the `entur` org and an **App ID** picked (3–10 lowercase alphanumeric chars, unique across Entur). See [self-service.md](../platform/self-service.md#gcp-project-naming).
- The service is **stateless**: no in-memory session, no local disk, no leader election. Cloud Run scales to zero, churns instances, and offers no session affinity.
- The service is **not** the right fit for GKE. If in doubt, start with [bootstrap-service.md](bootstrap-service.md) — that is the golden path.

## Steps

1. **Provision the GCP project with Kubernetes disabled.** Add `.entur/cicd.yaml` (`GitHubActions`) and `.entur/<appid>.yaml` (`GoogleCloudApplication`). Set `spec.kubernetes.enabled: false` and `spec.terraform.createBackend: true`. The Platform Orchestrator creates `ent-<appid>-prd`, the runtime service-account placeholder, and the Terraform state bucket `ent-gcs-tfa-<appid>`. See [self-service.md](../platform/self-service.md) for the manifest fields and the apply flow.

   ```yaml
   # .entur/<appid>.yaml
   apiVersion: orchestrator.entur.io/apps/v1
   kind: GoogleCloudApplication
   metadata:
     id: <appid>
     displayName: <Human Name>
     name: <repo-name>
     owner: team-<your-team>
   spec:
     environments: [prd]
     repositories: [<repo-name>]
     kubernetes:
       enabled: false
     terraform:
       createBackend: true
   ```

2. **Pin the Terraform provider to the new project.** Without an explicit `project`, plans default to the operator's gcloud ADC quota project and PR plans fail on IAM reads. In `terraform/providers.tf`:

   ```hcl
   provider "google" {
     project = "ent-${var.app_id}-${var.environment}"
     region  = var.region
   }
   ```

   Do not set `billing_project` or `user_project_override` — both require `roles/serviceusage.serviceUsageConsumer`, which the platform CI service account is not granted via the [`app_gcp_base`](../platform/terraform-modules.md) module. The GCS state backend lives in `ent-gcs-tfa-<appid>`, provisioned by step 1.

3. **Enable the required GCP APIs.** The Cloud Run Admin API (`run.googleapis.com`) is **mandatory** — without it, the deploy step in step 6 fails. Enable it on the project alongside the other APIs the LB and registry need:

   ```hcl
   resource "google_project_service" "apis" {
     for_each = toset([
       "run.googleapis.com",                  # Cloud Run Admin
       "compute.googleapis.com",              # HTTPS load balancer + serverless NEG + managed cert
       "artifactregistry.googleapis.com",     # container image registry
       "iam.googleapis.com",                  # runtime service account
       "secretmanager.googleapis.com",        # secrets injected into cloudrun.yaml
       "logging.googleapis.com",
       "monitoring.googleapis.com",
       "cloudresourcemanager.googleapis.com",
     ])
     project            = "ent-${var.app_id}-${var.environment}"
     service            = each.value
     disable_on_destroy = false
   }
   ```

4. **Write `cloudrun.yaml` at the repo root.** The Knative manifest is the source of truth for runtime config: env vars, scaling, probes, secrets, sidecars. Use `image: placeholder` — the CD workflow in step 6 substitutes the freshly-built tag.

   ```yaml
   # cloudrun.yaml
   apiVersion: serving.knative.dev/v1
   kind: Service
   metadata:
     name: <repo-name>
     annotations:
       run.googleapis.com/ingress: all
   spec:
     template:
       metadata:
         annotations:
           autoscaling.knative.dev/minScale: "1"   # one warm instance, no cold-start tax
           autoscaling.knative.dev/maxScale: "10"
           run.googleapis.com/cpu-throttling: "true"
       spec:
         serviceAccountName: <repo-name>@ent-<appid>-prd.iam.gserviceaccount.com
         timeoutSeconds: 60
         containers:
           - name: app
             image: placeholder
             ports:
               - containerPort: 8080
             resources:
               limits:
                 cpu: "1"
                 memory: 512Mi
             startupProbe:
               httpGet: { path: /health/liveness, port: 8080 }
               periodSeconds: 3
               failureThreshold: 10
             livenessProbe:
               httpGet: { path: /health/liveness, port: 8080 }
               periodSeconds: 10
     traffic:
       - percent: 100
         latestRevision: true
   ```

5. **Create the runtime service account and grant the minimum IAM.** `roles/logging.logWriter` is the floor for Cloud Logging; add `roles/monitoring.metricWriter` if the service writes Cloud Monitoring metrics (required for the Managed Prometheus sidecar). App-specific roles (Secret Manager accessor, Discovery Engine, …) go on top.

   ```hcl
   resource "google_service_account" "runtime" {
     project    = "ent-${var.app_id}-${var.environment}"
     account_id = "<repo-name>"
   }

   resource "google_project_iam_member" "log_writer" {
     project = "ent-${var.app_id}-${var.environment}"
     role    = "roles/logging.logWriter"
     member  = "serviceAccount:${google_service_account.runtime.email}"
   }

   resource "google_project_iam_member" "metric_writer" {
     project = "ent-${var.app_id}-${var.environment}"
     role    = "roles/monitoring.metricWriter"
     member  = "serviceAccount:${google_service_account.runtime.email}"
   }
   ```

   See [iam-roles.md](../platform/iam-roles.md) for the Entur-assignable role list. Do not grant blanket `roles/editor` or project-level `roles/run.admin`.

6. **Ask team-plattform for Artifact Registry Reader on the central registry.** The Cloud Run service pulls images from `eu.gcr.io/entur-system-1287/`, which lives in the central `entur-system-1287` project — a different project from yours, with IAM owned by team-plattform. Two identities in `ent-<appid>-prd` need `roles/artifactregistry.reader` there before the first deploy can succeed:

   - The runtime service account created in step 5: `<repo-name>@ent-<appid>-prd.iam.gserviceaccount.com`.
   - The Cloud Run Service Agent of the prd project: `service-<PROJECT_NUMBER>@serverless-robot-prod.iam.gserviceaccount.com`. Resolve `<PROJECT_NUMBER>` with:

     ```bash
     gcloud projects describe ent-<appid>-prd --format='value(projectNumber)'
     ```

   Cross-project IAM on `entur-system-1287` is not grantable from your repo's Terraform. Open a ticket in [#talk-utviklerplattform](https://entur.slack.com/archives/talk-utviklerplattform) with both service-account emails and ask for `roles/artifactregistry.reader` on `entur-system-1287`. Without this binding, the first CD run in step 9 fails with an image-pull error and no Cloud Run revision is created.

7. **Add the CD workflow.** GitHub Actions builds the image, pushes it to the central GAR (`eu.gcr.io/entur-system-1287/<image>`), substitutes the tag into `cloudrun.yaml`, and hands the manifest to `deploy-cloudrun`. Terraform does **not** own the Cloud Run service itself — the workflow does.

   ```yaml
   # .github/workflows/cd.yml — deploy step
   - name: Set image in cloudrun.yaml
     env:
       IMAGE: eu.gcr.io/entur-system-1287/${{ needs.ci.outputs.image_and_tag }}
     run: sed -i "s|image: placeholder|image: ${IMAGE}|" cloudrun.yaml

   - name: Deploy to Cloud Run
     uses: google-github-actions/deploy-cloudrun@2028e2d7d30a78c6910e0632e48dd561b064884d # v3.0.1
     with:
       metadata: cloudrun.yaml
       region: europe-west1
       project_id: ent-<appid>-prd
   ```

   The `metadata:` input triggers `gcloud run services replace`, which ignores the action's `image:` input — that is why the `sed` step is needed. See [gha-workflows.md](../platform/gha-workflows.md) for the surrounding workflow shape (image build via `entur/gha-docker`, workload-identity auth via `entur/gha-meta`).

8. **Front the service with a global HTTPS load balancer.** team-plattform owns the `entur.{no,io,org}` DNS zones and only attaches A records to a stable IP, so the Cloud Run service must sit behind an LB that has one. Cloud Run's built-in `*.run.app` URL and `google_cloud_run_domain_mapping` are **not** valid paths — the latter requires Search Console domain ownership the platform does not grant.

   Provision the full chain in Terraform: a `google_compute_global_address` (the static IPv4 to share), a `google_compute_region_network_endpoint_group` of type `SERVERLESS` pointing at the Cloud Run service by name, a `google_compute_backend_service`, a `google_compute_url_map`, a `google_compute_managed_ssl_certificate` for the hostname, a `google_compute_target_https_proxy`, and a `google_compute_global_forwarding_rule` on port 443. Expose the IP via an output so team-plattform can read it:

   ```hcl
   output "lb_ipv4" {
     value = google_compute_global_address.lb_ipv4.address
   }
   ```

   Managed SSL certificates cannot be edited in place. Wrap the cert's name in a `random_id` suffix with `lifecycle { create_before_destroy = true }` so a domain change can rotate the cert without a window where the HTTPS proxy has no cert attached.

9. **Expect a one-time failure cycle on first deploy.** The dependencies between Cloud Run, the LB, and cross-project Artifact Registry IAM cannot be satisfied in a single pass — the actual order is deploy → fail → terraform → fail → team-plattform → deploy → success. Walk through it:

   1. **First CD run fails** with an image-pull error. Neither the Cloud Run Service Agent nor the runtime SA have `roles/artifactregistry.reader` on `entur-system-1287` yet, so the revision cannot pull the image.
   2. **First `terraform apply` of the LB chain fails** for the same generation. The serverless NEG and the public-invoker binding reference the Cloud Run service by name, but no service exists yet.
   3. **Open (or chase) the team-plattform ticket from step 6** with the runtime SA email and the Cloud Run Service Agent email. Wait for the grant.
   4. **Re-run CD.** It succeeds. The Cloud Run service is created.
   5. **Re-run `terraform apply`.** The LB chain comes up; the managed cert enters `PROVISIONING`.
   6. **Open a second team-plattform ticket** in [#talk-utviklerplattform](https://entur.slack.com/archives/talk-utviklerplattform) with the LB IPv4 (`terraform output lb_ipv4`) and the hostname under `entur.{no,io,org}`. Reference [add-custom-domain.md → Off the golden path](add-custom-domain.md#off-the-golden-path).
   7. **Once team-plattform attaches the A record**, Google validates the managed cert (15–60 min) and it flips to `ACTIVE`.

   Subsequent deploys are one-shot — this cycle only happens on initial bootstrap.

## Verify

- `terraform apply` on the prd workspace is clean.
- The CD workflow finishes green; `gcloud run services describe <repo-name> --region europe-west1 --project ent-<appid>-prd` reports a ready latest revision.
- `curl -s -o /dev/null -w "%{http_code}\n" --resolve "<hostname>:443:$(terraform output -raw lb_ipv4)" https://<hostname>/health/liveness` returns `200`. This works before DNS propagates because `--resolve` bypasses the resolver.
- In the GCP console under **Network services → Load balancing → Certificates** the managed cert status is `ACTIVE`.
- Public DNS resolves the hostname to the LB IP and the URL returns `200` in a browser without a certificate warning.

If the managed cert stays in `FAILED_NOT_VISIBLE` after an hour with the A record in place, force a fresh provisioning attempt: `terraform apply -replace=random_id.cert_suffix`.

## See also

- [bootstrap-service.md](bootstrap-service.md) — golden-path GKE alternative
- [add-custom-domain.md](add-custom-domain.md#off-the-golden-path) — DNS + TLS handoff to team-plattform
- [self-service.md](../platform/self-service.md) — `GoogleCloudApplication` manifest fields
- [gha-workflows.md](../platform/gha-workflows.md) — reusable CI workflows the CD step builds on
- [terraform-modules.md](../platform/terraform-modules.md) — Entur Terraform modules and `app_gcp_base` constraints
- [iam-roles.md](../platform/iam-roles.md) — Entur-assignable role list
- The `entur/ai-portal-mcp` repository is a production worked example of every step above
