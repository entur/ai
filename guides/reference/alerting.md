# Alerting

How to set up alerts for an Entur service using Grafana, route them through PagerDuty, and write PromQL queries that work well against the shared Prometheus/Thanos stack.

- **Target audience**: developers adding alerting to a new or existing service.
- **Intent**: the team creates Grafana alert rules that fire on real problems, route through the shared PagerDuty integration, and reach the right on-call responder without manual wiring.
- **Scope**: Grafana alert rules (golden path), PromQL for alerting, PagerDuty routing, notification policies. GCP Cloud Monitoring alerting is mentioned as an alternative. Health probes, metrics instrumentation, and tracing live in [observability.md](observability.md), [tracing.md](tracing.md), and [profiler.md](profiler.md).
- **Prerequisites**: the service exposes Prometheus metrics (see [observability.md](observability.md)) and is deployed to at least the `dev` environment.

## What Makes a Good Alert

Alert on **symptoms, not causes**. A good alert is actionable -- when it fires, someone can and should do something about it. A bad alert trains the team to ignore alerts.

- **Alert on user-visible impact**: high error rate, elevated latency, broken health checks. These tell you something is wrong from the user's perspective.
- **Avoid alerting on transient signals**: a single pod restart, a brief CPU spike, or a momentary increase in queue depth. Use `for` durations to filter noise.
- **Every alert should have a clear response**: if the on-call person cannot act on it, the alert should not exist.
- **Prefer rates and ratios over raw counts**: `rate(http_server_requests_seconds_count{status=~"5.."}[5m]) / rate(http_server_requests_seconds_count[5m])` is more meaningful than a raw error count that scales with traffic.
- **Account for traffic patterns**: set thresholds that tolerate normal daily variation. Alerting on an absolute request rate will fire every rush hour and teach the team to ignore the alert.

## Grafana Alert Rules (Golden Path)

Create and manage alert rules in [grafana.entur.org](https://grafana.entur.org). This is the recommended path for all Entur services.

### Organize Alert Rules by Team

Each team owns a **Grafana folder** for their alert rules. Create one if it does not exist. Name it after the team, not after individual services -- a single folder holds alert rules for all services the team operates.

### Create an Alert Rule

1. Open **Alerting > Alert rules** in Grafana.
2. Click **New alert rule**.
3. Select the Prometheus data source.
4. Write the PromQL query (see [PromQL for alerting](#promql-for-alerting) below).
5. Set the **condition** (threshold, math expression, or reduce function).
6. Set the **evaluation interval** and `for` duration. There is no hard rule, but reasonable defaults are 1m evaluation with 5m pending for warnings and 3m pending for critical alerts.
7. Add **labels** -- these drive notification routing:
   - `severity`: `critical` (-> High urgency) or `warning` (-> Low urgency)
   - `service_name`: the **PagerDuty service name**, matched verbatim (see [Labels That Drive Routing](#labels-that-drive-routing)). This is *not* automatically the Kubernetes namespace.
   - `env`: the environment (`dev`, `tst`, `prd`)
8. Add **annotations** for context in the alert notification:
   - `summary`: one-line description of the problem
   - `description`: details, thresholds, and a link to the relevant Grafana dashboard if one exists
9. Save the rule into your team's folder.

### Labels That Drive Routing

The labels `service_name` and `env` are sent as `custom_details` to PagerDuty; Global Event Orchestration uses them to route the incident to the right team's PagerDuty service. Set them on every alert rule. Use `severity` to control urgency (`critical` -> High, `warning` -> Low).

- **`service_name` must equal the PagerDuty service's display name exactly** -- same spelling, casing, and spacing as the service's `spec.name` in its [Service manifest](https://github.com/entur/so-incident-management/blob/main/docs/SERVICE.md). Event Orchestration matches on this exact string; any mismatch (typo, wrong case, extra space) means the alert is **not routed** and falls through. It is *not* automatically the Kubernetes namespace, and the display name (e.g. `My Service`) is usually different from the kebab-case namespace (`my-service`) you use in PromQL selectors.
- **`env`** scopes routing to the environment: `dev`, `tst`, or `prd`.

> **The PagerDuty service must already exist.** Teams, services, escalation policies, and Slack connections are provisioned declaratively from manifests in `entur/so-incident-management` -- not created by hand in the PagerDuty UI. If no service exists yet for your app, add a [Service manifest](https://github.com/entur/so-incident-management/blob/main/docs/SERVICE.md) (open a PR, `entur apply`, merge) before pointing `service_name` at it. See the [PagerDuty Integration](#pagerduty-integration) section below.

## PromQL for Alerting

Entur runs self-hosted Prometheus with Thanos for long-term storage and cross-cluster queries. All PromQL queries in Grafana run against Thanos.

### Select the environment

Always include `prometheus_group` or `cluster_environment` in alert queries to scope to a specific cluster or environment.

| `cluster_environment` | `prometheus_group` example |
|-------------|--------------------------|
| dev         | `kub-ent-dev-001`        |
| tst         | `kub-ent-tst-001`        |
| prd         | `kub-ent-prd-001`        |

Production alerts should always filter on `cluster_environment="prd"`.

### Standard Labels

These labels are available on most metrics from Kubernetes workloads:

| Label | Description | Example |
|-------|-------------|---------|
| `app` | Application name | `my-service` |
| `kubernetes_namespace` | Namespace the pod runs in | `my-service` |
| `cluster_environment` | Environment | `dev`, `tst`, `prd` |
| `prometheus_group` | Cluster identifier | `kub-ent-prd-001` |

### Metric Sources

| Source | Prefix / pattern | Provides |
|--------|-----------------|----------|
| cAdvisor / kubelet | `container_*` | Container CPU, memory, network, filesystem |
| kube-state-metrics | `kube_*` | Pod status, replica counts, resource requests/limits |
| Kafka | `kafka_` | Kafka details |
| Http | `http_(server\|client)` | HTTP details |
| gRPC | `grpc_(server\|client)` | gRPC details |
| Micrometer (Spring Boot) | `jvm_*`, `process_*` | Application-level HTTP, JVM, and process metrics |
| Hikari connnection pool | `hikaricp_` | Connection pool details |
| Logging (logback) | `logback_` | Log details |
| Custom application metrics | Varies | Business-specific counters, gauges, histograms |

### PromQL Examples for Common Alerts

All examples below target production. Adjust `prometheus_group`, `kubernetes_namespace`, and thresholds for your service.

#### High 5xx Error Rate

Fires when more than 5% of HTTP requests return 5xx for 5 minutes:

```promql
(
  sum(rate(http_server_requests_seconds_count{
    prometheus_group="kub-ent-prd-001",
    kubernetes_namespace="my-service",
    status=~"5.."
  }[5m]))
/
  sum(rate(http_server_requests_seconds_count{
    prometheus_group="kub-ent-prd-001",
    kubernetes_namespace="my-service"
  }[5m]))
) > 0.05
```

#### High p99 Latency

Fires when the 99th percentile request duration exceeds 5 seconds. Use the pre-calculated recording rule `entur_service_latency_seconds:p99` instead of computing `histogram_quantile` over raw buckets -- it is cheaper to evaluate and consistent across dashboards and alerts:

```promql
entur_service_latency_seconds:p99{
  prometheus_group="kub-ent-prd-001",
  kubernetes_namespace="my-service"
} > 5
```

#### Pod Restarts

Fires when a container restarts more than 3 times in 15 minutes:

```promql
increase(kube_pod_container_status_restarts_total{
  prometheus_group="kub-ent-prd-001",
  namespace="my-service"
}[15m]) > 3
```

#### Memory Close to Limit

Fires when container memory usage exceeds 85% of its limit:

```promql
(
  sum by (pod) (container_memory_working_set_bytes{
    prometheus_group="kub-ent-prd-001",
    namespace="my-service",
    container!=""
  })
/
  sum by (pod) (kube_pod_container_resource_limits{
    prometheus_group="kub-ent-prd-001",
    namespace="my-service",
    resource="memory"
  })
) > 0.85
```

#### CPU Throttling

Fires when containers experience sustained CPU throttling:

```promql
(
  sum by (pod) (rate(container_cpu_cfs_throttled_periods_total{
    prometheus_group="kub-ent-prd-001",
    namespace="my-service",
    container!=""
  }[5m]))
/
  sum by (pod) (rate(container_cpu_cfs_periods_total{
    prometheus_group="kub-ent-prd-001",
    namespace="my-service",
    container!=""
  }[5m]))
) > 0.5
```

#### Readiness Probe Failing

Fires when a pod is not ready for 3 minutes:

```promql
kube_pod_status_ready{
  prometheus_group="kub-ent-prd-001",
  namespace="my-service",
  condition="true"
} == 0
```

### PromQL Reference

For learning PromQL syntax, functions, and operators:

- [Prometheus querying basics](https://prometheus.io/docs/prometheus/latest/querying/basics/) -- selectors, ranges, offsets
- [Prometheus querying operators](https://prometheus.io/docs/prometheus/latest/querying/operators/) -- arithmetic, comparison, logical, vector matching
- [Prometheus querying functions](https://prometheus.io/docs/prometheus/latest/querying/functions/) -- `rate`, `increase`, `histogram_quantile`, aggregations
- [Awesome Prometheus alerts](https://samber.github.io/awesome-prometheus-alerts/) -- curated collection of reusable alert rules by category

## PagerDuty Integration

Entur uses a shared PagerDuty Global Event Orchestration. All alerts flow through a single integration point -- PagerDuty routes to the correct team based on alert metadata.

### How It Works

1. A Grafana alert fires and sends a notification to the **`entur-pagerduty`** contact point.
2. PagerDuty receives the event with `service_name` and `env` as custom details.
3. PagerDuty's Global Event Orchestration routes the event to the correct team's PagerDuty service based on these fields (no per-service integration key needed).
4. PagerDuty handles on-call scheduling and escalation for that service.

You do not wire up PagerDuty from Grafana beyond the `entur-pagerduty` contact point. But the team's PagerDuty resources -- the **service**, on-call schedule, **escalation policy**, and Slack connections -- are managed declaratively through the Incident Management Sub-Orchestrator, **not** by clicking around in the PagerDuty UI (manual changes are reverted on the next apply). Provision them from manifests in [`entur/so-incident-management`](https://github.com/entur/so-incident-management):

- [Service manifest](https://github.com/entur/so-incident-management/blob/main/docs/SERVICE.md) -- the service your `service_name` must match, plus its escalation-policy assignment.
- [Team manifest](https://github.com/entur/so-incident-management/blob/main/docs/TEAM.md) -- team members, on-call schedule, escalation policy, and the team's Slack channel.

### Escalation Model

Escalation is driven by **urgency**, which the alert's `severity` sets, and the recipient chain comes from the team's escalation policy (provisioned from the [Team manifest](https://github.com/entur/so-incident-management/blob/main/docs/TEAM.md), integrated with the [Entur incident-management process](https://entur.atlassian.net/wiki/spaces/EKH/pages/5488541894/Varslingsrutine+for+Kritiske+hendelser)):

| Urgency | `severity` | Recipient |
|---------|-----------|-----------|
| **High** | `critical` | On-call responder paged via SMS / automated phone call, 24/7. |
| **Low** | `warning` | On-call responder notified by e-mail and/or Slack only. |

Change escalation via the Team manifest (or override per service with `spec.escalationPolicy`) -- not in the PagerDuty UI. The team's responsibility on the alerting side is to create alerts that fire on real, actionable problems and carry the right `severity`.

## Grafana Notification Policies

Notification policies in Grafana control which contact point receives an alert. The routing tree matches on labels.

For most teams, the setup is:

- A notification policy that matches on `service_name` (or `team` if the team uses a team-level label) and routes to the **`entur-pagerduty`** contact point.
- `severity=critical` alerts should have shorter group wait/interval to reach on-call faster.
- `severity=warning` alerts can use longer intervals.

> **Slack routing is handled by PagerDuty, not Grafana.** Once an incident reaches PagerDuty it is posted to Slack via the team's orchestrator-managed Slack connections (the team channel from the [Team manifest](https://github.com/entur/so-incident-management/blob/main/docs/TEAM.md), plus a shared platform channel). Prefer this over adding a separate Grafana Slack contact point, so on-call notifications and Slack posts stay consistent and don't drift from PagerDuty's view of the incident.

## SLO-Based Alerting

Entur has a centralized **SLODash** setup that tracks SLA/SLO compliance for services that are covered by service-level agreements. SLO-based alerts are managed outside individual teams. If your service is under an SLA, the SLO alerts may already exist -- check with your team lead or `#talk-utviklerplattform` before duplicating them.

## GCP Cloud Monitoring

Some teams use Google Cloud Monitoring alerting policies as an alternative or supplement. This is supported but not the recommended path. The Grafana-based workflow gives teams a single pane of glass for dashboards and alerts, backed by the same Prometheus/Thanos data source.

If you use Cloud Monitoring alerting, configure notification channels to route to PagerDuty or the team's preferred channel independently from the Grafana setup.

## Recommended Starting Alerts

Every production service should have at minimum:

| Alert | Condition | Severity | `for` |
|-------|-----------|----------|-------|
| High error rate | 5xx rate > 5% | Critical | 5m |
| High latency | p99 > 5s | Warning | 10m |
| Pod restarts | > 3 restarts in 15m | Warning | 0m |
| Memory near limit | > 85% of limit | Critical | 5m |
| Readiness probe failing | Pod not ready | Critical | 3m |

Adjust thresholds to your service's traffic profile and SLO targets. These are starting points, not universal rules.

## Further Reading

- [Service manifest](https://github.com/entur/so-incident-management/blob/main/docs/SERVICE.md) / [Team manifest](https://github.com/entur/so-incident-management/blob/main/docs/TEAM.md) -- the authority for provisioning PagerDuty services, escalation policies, and Slack connections. Your `service_name` label must match a service's `spec.name` here.
- [Sub-Orchestrators overview](https://entur.atlassian.net/wiki/spaces/ESP/pages/5421170741/Sub-Orchestrators) -- how incident-management resources are applied.
- [observability.md](observability.md) -- health probes, Prometheus metrics setup, Grafana dashboards.
- [tracing.md](tracing.md) -- distributed tracing for correlating alerts with request traces.
- [logging.md](logging.md) -- structured logging for investigating fired alerts.
- [profiler.md](profiler.md) -- CPU and heap profiling for diagnosing performance alerts.
