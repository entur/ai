# Assignable IAM Roles

These are the IAM roles that CD service accounts are allowed to **grant to other identities**
(e.g. service accounts used by your application at runtime) via `google_project_iam_member`
or `google_project_iam_binding` in your Terraform code.

If a role you need is not listed here, request it to be added in the #talk-utviklerplattform channel on Slack.

## Allowed roles

| Role                                   | Service                  |
| -------------------------------------- | ------------------------ |
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
| `roles/cloudsql.admin`                 | Cloud SQL                |
| `roles/cloudsql.client`                | Cloud SQL                |
| `roles/dataform.editor`                | Dataform                 |
| `roles/eventarc.eventReceiver`         | Eventarc                 |
| `roles/firebase.developAdmin`          | Firebase                 |
| `roles/firebase.developViewer`         | Firebase                 |
| `roles/firebaseauth.admin`             | Firebase Auth            |
| `roles/firebasecloudmessaging.admin`   | Firebase Cloud Messaging |
| `roles/firebasehosting.admin`          | Firebase Hosting         |
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
