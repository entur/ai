# Assignable IAM Roles

This page is the **authoritative allowlist** of IAM roles that CD service accounts may **grant to other identities** (e.g. the service account your application runs as) via `google_project_iam_member` or `google_project_iam_binding` in your Terraform code. Roles outside this list will be rejected by the platform's policy guard.

If a role you need is not on this allowlist, request it to be added in the #talk-utviklerplattform channel on Slack.

The canonical list is the `assignable_iam_roles` Terraform variable in [entur/tf-gcp-apps](https://github.com/entur/tf-gcp-apps/blob/main/terraform/modules/modules/app_gcp_base/variables.tf) -- the table below mirrors it. See [Updating the allowlist](#updating-the-allowlist) before editing.

## Allowed roles (allowlist)

| Role                                   | Service                  |
| -------------------------------------- | ------------------------ |
| `roles/appengine.appViewer`            | App Engine               |
| `roles/bigquery.admin`                 | BigQuery                 |
| `roles/bigquery.connectionAdmin`       | BigQuery                 |
| `roles/bigquery.dataEditor`            | BigQuery                 |
| `roles/bigquery.dataOwner`             | BigQuery                 |
| `roles/bigquery.dataViewer`            | BigQuery                 |
| `roles/bigquery.jobUser`               | BigQuery                 |
| `roles/bigquery.metadataViewer`        | BigQuery                 |
| `roles/bigquery.readSessionUser`       | BigQuery                 |
| `roles/bigquery.user`                  | BigQuery                 |
| `roles/cloudfunctions.invoker`         | Cloud Functions          |
| `roles/cloudfunctions.viewer`          | Cloud Functions          |
| `roles/cloudsql.admin`                 | Cloud SQL                |
| `roles/cloudsql.client`                | Cloud SQL                |
| `roles/dataform.editor`                | Dataform                 |
| `roles/eventarc.eventReceiver`         | Eventarc                 |
| `roles/firebase.developAdmin`          | Firebase                 |
| `roles/firebase.developViewer`         | Firebase                 |
| `roles/firebaseauth.admin`             | Firebase Auth            |
| `roles/firebasecloudmessaging.admin`   | Firebase Cloud Messaging |
| `roles/firebasehosting.admin`          | Firebase Hosting         |
| `roles/firebasehosting.viewer`         | Firebase Hosting         |
| `roles/iam.serviceAccountTokenCreator` | IAM                      |
| `roles/iam.serviceAccountViewer`       | IAM                      |
| `roles/logging.logWriter`              | Cloud Logging            |
| `roles/logging.viewer`                 | Cloud Logging            |
| `roles/monitoring.viewer`              | Cloud Monitoring         |
| `roles/pubsub.publisher`               | Pub/Sub                  |
| `roles/pubsub.subscriber`              | Pub/Sub                  |
| `roles/pubsub.viewer`                  | Pub/Sub                  |
| `roles/run.invoker`                    | Cloud Run                |
| `roles/run.viewer`                     | Cloud Run                |
| `roles/secretmanager.secretAccessor`   | Secret Manager           |
| `roles/secretmanager.viewer`           | Secret Manager           |
| `roles/serviceusage.apiKeysViewer`     | Service Usage            |
| `roles/storage.bucketViewer`           | Cloud Storage            |
| `roles/storage.objectAdmin`            | Cloud Storage            |
| `roles/storage.objectCreator`          | Cloud Storage            |
| `roles/storage.objectViewer`           | Cloud Storage            |
| `roles/workflows.invoker`              | Workflows                |

## Updating the allowlist

The platform policy guard enforces the list in [entur/tf-gcp-apps](https://github.com/entur/tf-gcp-apps/blob/main/terraform/modules/modules/app_gcp_base/variables.tf) (`assignable_iam_roles` -- the second `variable` block in that file, distinct from `addable_iam_roles` above it). This page is a mirror for AI agents; the upstream variable is what actually gates Terraform plans.

To add or remove a role:

1. Open a PR against [entur/tf-gcp-apps](https://github.com/entur/tf-gcp-apps) editing `assignable_iam_roles` in `terraform/modules/modules/app_gcp_base/variables.tf`. Keep the list alphabetically sorted -- the existing entries are sorted by full role string (e.g. `roles/cloudprofiler.agent` before `roles/cloudsql.admin`).
2. State the requesting team, the consuming repository, and the narrowest GCP-predefined role that grants the required permission. Do **not** request `roles/*.admin` or `roles/*.editor` when a write-only agent role (e.g. `roles/cloudtrace.agent`, `roles/cloudprofiler.agent`) suffices.
3. After the upstream PR merges and a tagged release ships, open a follow-up PR here updating both the table above and this section if the canonical commit reference moves. Add a row in the order produced by `sort` on the role string -- the upstream is the source of truth for ordering.
4. Announce the addition in `#talk-utviklerplattform` so consuming teams know it is available without polling the tag stream.

Do **not** edit only this page. A role added here without the upstream change will fail at `terraform plan` for every consumer; a role removed upstream without removing it here will silently mislead agents into picking a role the policy guard now rejects.

If you spot drift between this table and the upstream variable, open an issue (or a small PR mirroring the upstream) -- the table is regenerated by hand today, so periodic drift is expected.
