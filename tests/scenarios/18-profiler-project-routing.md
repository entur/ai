# Scenario: Cloud Profiler Project Routing for Kubernetes Workloads

## Description

Verifies that the agent picks the **application's** per-env project (`ent-<app>-<env>`) for Cloud Profiler reads on a Kubernetes workload, **not** the cluster host project (`ent-kub-<env>`). Symmetric with scenario `17-trace-project-routing`: the profiler and trace agents both report under the workload's SA, but Cloud Logging routes k8s logs to the cluster host project. Agents commonly conflate the three.

## Prompt

You are helping an Entur engineer investigate a heap leak in a Kubernetes-runtime service.

Details:

- Repository: `entur/some-repo`
- App ID (metadata.id): `someapp`
- Runtime: Kubernetes (deployed via the common Helm chart)
- Environment: prd
- Cluster host project: `ent-kub-prd`

The engineer wants to inspect Cloud Profiler heap profiles from the last 24 hours.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the profiler reference) and answer in `key: value` format on its own line:

- profiler_project_id: <the GCP project ID to query Cloud Profiler in>
- logs_project_id: <the GCP project ID to query Cloud Logging in for the workload's stdout>
- service_version_source: <what value to set as the Profiler agent's ServiceVersion on Kubernetes>
- one_line_reason: <one sentence explaining why profiles and k8s logs live in different projects>

## Assertions

```json
{
  "must_contain": [
    "profiler_project_id: ent-someapp-prd",
    "logs_project_id: ent-kub-prd"
  ],
  "must_not_contain": [
    "profiler_project_id: ent-kub-prd",
    "logs_project_id: ent-someapp-prd"
  ],
  "must_match": [
    "service_version_source:.*(Helm release revision|Helm-injected label|image tag|helm-release-revision)",
    "(workload|application|app|agent).*(report|stamp|write|emit|upload|own).*project|kubelet.*cluster"
  ]
}
```

## Budget

0.10
