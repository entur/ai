# Scenario: Cloud Trace Project Routing for Kubernetes Workloads

## Description

Verifies that the agent picks the **application's** per-env project (`ent-<app>-<env>`) for Cloud Trace reads on a Kubernetes workload, **not** the cluster host project (`ent-kub-<env>`). This is the most-confused routing rule: Cloud Logging routes k8s logs to the cluster host project, but Cloud Trace (and Cloud Profiler) route to the application project regardless of runtime.

## Prompt

You are helping an Entur engineer debug a slow request in a Kubernetes-runtime service.

Details:

- Repository: `entur/journey-planner`
- App ID (metadata.id): `jrnplan`
- Runtime: Kubernetes (deployed via the common Helm chart)
- Environment: prd
- Cluster host project: `ent-kub-prd`

The engineer wants to query Cloud Trace for slow traces from the last hour.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the tracing and observability references) and answer in `key: value` format on its own line:

- trace_project_id: <the GCP project ID to query Cloud Trace in>
- logs_project_id: <the GCP project ID to query Cloud Logging in for the workload's stdout>
- one_line_reason: <one sentence explaining why traces and k8s logs live in different projects>

## Assertions

```json
{
  "must_contain": [
    "trace_project_id: ent-jrnplan-prd",
    "logs_project_id: ent-kub-prd"
  ],
  "must_not_contain": [
    "trace_project_id: ent-kub-prd",
    "logs_project_id: ent-jrnplan-prd"
  ],
  "must_match": [
    "(workload|application|app).*(report|stamp|write|emit|own).*project|kubelet.*cluster|trace.*agent.*workload|exporter.*workload"
  ]
}
```

## Budget

0.10
