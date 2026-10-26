# Scenario: Cloud Trace Project Routing for Kubernetes Workloads

## Description

Verifies that the agent picks the shared cluster host project (`ent-kub-<env>`) for Cloud Trace reads on a Kubernetes workload, matching where Cloud Logging writes the workload's k8s logs -- **not** the application's per-env project (`ent-<app>-<env>`). This is the most-confused routing rule: Cloud Trace and Cloud Logging both route Kubernetes signals to the shared host project (which is what lets a trace be correlated with its logs), while Cloud Profiler is the exception that reports to the application's own project regardless of runtime. Agents often assume all telemetry is app-scoped like Cloud Profiler, or that trace and log routing differ from each other.

## Prompt

You are helping an Entur engineer debug a slow request in a Kubernetes-runtime service.

Details:

- Repository: `entur/some-repo`
- App ID (metadata.id): `someapp`
- Runtime: Kubernetes (deployed via the common Helm chart)
- Environment: prd
- Cluster host project: `ent-kub-prd`

The engineer wants to query Cloud Trace for slow traces from the last hour.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the tracing and observability references) and answer in `key: value` format on its own line:

- trace_project_id: <the GCP project ID to query Cloud Trace in>
- logs_project_id: <the GCP project ID to query Cloud Logging in for the workload's stdout>
- one_line_reason: <one sentence explaining why traces and k8s logs land in the same project>

## Assertions

```json
{
  "must_contain": [
    "trace_project_id: ent-kub-prd",
    "logs_project_id: ent-kub-prd"
  ],
  "must_not_contain": [
    "trace_project_id: ent-someapp-prd",
    "logs_project_id: ent-someapp-prd"
  ],
  "must_match": [
    "(correlat|same project|shared|host project).*(trace|log)|(trace|log).*(correlat|same project|shared|host project)"
  ]
}
```

## Budget

0.10
