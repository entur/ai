<!--
metadata:
  source:
    repo: entur/tf-gcp-apps
    path: terraform/modules/modules/app_gcp_base/variables.tf
    variable: assignable_iam_roles
    url: https://github.com/entur/tf-gcp-apps/blob/main/terraform/modules/modules/app_gcp_base/variables.tf
    note: >
      The platform policy guard enforces this upstream variable, not the table
      below. The table is a hand-maintained mirror for AI agents reading these
      docs; the upstream is the source of truth and what `terraform plan`
      checks against.
  sync_procedure:
    - >
      Open a PR against entur/tf-gcp-apps editing `assignable_iam_roles` in
      `terraform/modules/modules/app_gcp_base/variables.tf`. Keep the list
      alphabetically sorted by full role string.
    - >
      Request the narrowest GCP-predefined role that grants the required
      permission. Prefer write-only agent roles (e.g. `roles/cloudtrace.agent`,
      `roles/cloudprofiler.agent`) over `roles/*.admin` or `roles/*.editor`.
    - >
      After the upstream PR merges and a tagged release ships, open a follow-up
      PR here mirroring the new row into the table below, in the same sort
      order.
    - >
      Announce the addition in #talk-utviklerplattform so consuming teams know
      it is available without polling the tag stream.
  edit_warning: >
    Do not edit only this page. A role added here without the upstream change
    fails at `terraform plan` for every consumer. A role removed upstream
    without removing it here silently misleads agents into picking a role the
    policy guard now rejects.
-->

# Assignable IAM Roles

This page is the **authoritative allowlist** of IAM roles that CD service accounts may **grant to other identities** (e.g. the service account your application runs as) via `google_project_iam_member` or `google_project_iam_binding` in your Terraform code. Roles outside this list will be rejected by the platform's policy guard.

If a role you need is not on this allowlist, request it to be added in the #talk-utviklerplattform channel on Slack.

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
