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
   - `severity`: `critical` or `warning`
   - `service_name`: the application name (Kubernetes namespace)
   - `env`: the environment (`dev`, `tst`, `prd`)
8. Add **annotations** for context in the alert notification:
   - `summary`: one-line description of the problem
   - `description`: details, thresholds, and a link to the relevant Grafana dashboard if one exists
9. Save the rule into your team's folder.

### Labels That Drive Routing

The labels `service_name` and `env` are sent as `custom_details` to PagerDuty and determine routing and escalation. Set them on every alert rule. Use `severity` to control urgency.

## PromQL for Alerting

Entur runs self-hosted Prometheus with Thanos for long-term storage and cross-cluster queries. All PromQL queries in Grafana run against Thanos.

### The `prometheus_group` Selector

Always include `prometheus_group` in alert queries to scope to a specific cluster. This avoids cross-cluster fan-out and significantly reduces query cost.

| Environment | `prometheus_group` value |
|-------------|--------------------------|
| dev         | `kub-ent-dev-001`        |
| tst         | `kub-ent-tst-001`        |
| prd         | `kub-ent-prd-001`        |

Production alerts should always filter on `prometheus_group="kub-ent-prd-001"`.

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
| Micrometer (Spring Boot) | `http_server_requests_*`, `jvm_*`, `process_*` | Application-level HTTP, JVM, and process metrics |
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

Fires when the 99th percentile request duration exceeds 5 seconds:

```promql
histogram_quantile(0.99,
  sum by (le) (rate(http_server_requests_seconds_bucket{
    prometheus_group="kub-ent-prd-001",
    kubernetes_namespace="my-service"
  }[5m]))
) > 5
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
3. PagerDuty's Global Event Orchestration routes the event to the correct team's PagerDuty service based on these fields.
4. PagerDuty handles on-call scheduling and escalation.

Teams do not need to configure PagerDuty directly. Set the correct `service_name` and `env` labels on every alert rule and PagerDuty takes care of routing and escalation.

### Escalation Model

- During work hours: the owning team's on-call responds.
- Outside work hours: escalation to **utviklervakt** (developer on-call).
- Further escalation: **incident manager**.

PagerDuty manages the schedule and escalation chain. The team's responsibility is to create alerts that fire on real, actionable problems.

## Grafana Notification Policies

Notification policies in Grafana control which contact point receives an alert. The routing tree matches on labels.

For most teams, the setup is:

- A notification policy that matches on `service_name` (or `team` if the team uses a team-level label) and routes to the **`entur-pagerduty`** contact point.
- `severity=critical` alerts should have shorter group wait/interval to reach on-call faster.
- `severity=warning` alerts can use longer intervals or route to a less urgent channel (e.g. a Slack contact point) depending on the team's preference.

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

- [observability.md](observability.md) -- health probes, Prometheus metrics setup, Grafana dashboards.
- [tracing.md](tracing.md) -- distributed tracing for correlating alerts with request traces.
- [logging.md](logging.md) -- structured logging for investigating fired alerts.
- [profiler.md](profiler.md) -- CPU and heap profiling for diagnosing performance alerts.
