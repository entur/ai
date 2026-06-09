# Incident Response

Handle critical incidents from detection through resolution and post mortem. This playbook covers the "Kritisk Hendelseshåndtering" process owned by Team Kvalitet.

## Goal

Restore service as fast as possible, communicate clearly to operators and stakeholders, and capture learnings to prevent recurrence.

## Roles

| Role | Responsibility |
|------|----------------|
| **Utvikler / Utviklervakt** | Analyses and remediates the fault. Updates progress on `#talk-incidents`. Contacts the relevant team for fixes. Documents and writes the Post Mortem. |
| **Incident Manager (IM) / Bakvakt** | Leads the critical incident process. Drives communication via status.entur.org. Documents the incident timeline. Creates the Post Mortem page. Creates a ticket in the KHH project. |
| **Support** | Sends first-line notifications to operators (Vy Drops, SJ OPS, GoA Ops). Notifies the KS shift leader. Answers all inbound queries related to the incident. Keeps OPS centres informed via phone or email. |
| **Beredskapsledergruppe (BLG)** | Crisis leadership group. Activated by Administrerende direktør (or ELG) for incidents involving økonomiske tap, omdømmetap, or GDPR-brudd. Shall be involved after 6 hours per Beredskapsplanen, but assessed earlier for GDPR, payment card, and security incidents. |

## What Is a Critical Incident

The definition is established in the operator agreements (see Avtaleregisteret for all active agreements).

A **B-feil** (Nivå B -- alvorlig feil) is also treated as a critical incident when a **sales channel** is affected (e.g. TVM or MT). These must follow the full critical incident process.

Incidents that do not directly affect sales but affect **downstream systems** (e.g. Avregning, E-journal, Cleos/Eos) are also critical incidents. These can cause reporting violations -- for example, missing MVA data causes incorrect settlement and may breach legal requirements. Such incidents must be reported and handled using the critical incident process. Reporting for these types of faults should use a different notification method than SMS, decided in collaboration with Entur's leadership group.

## Categorization of Operational Incidents

| Nivå | Kategori | Description | Examples |
|------|----------|-------------|----------|
| **A** | Kritisk feil | Whole or significant parts of the service are unavailable -- all service types and/or components are down. Also used for security incidents. If Entur has implemented functional redundancy and the alternative is a fully adequate replacement, the incident is not category A. | All or very many travelers cannot buy tickets; travelers receive incorrect information in channels; data loss; a service type is unavailable; transfers to other critical systems fail; network problems prevent communication (excluding distributed locations for automater and stasjonssalg) |
| **B** | Alvorlig feil | Individual functions do not work or have significantly degraded response times. A service or component has reduced availability, making work harder for the operator, users, or the supplier. | A single channel is unavailable or has poor response time; measured response times fall within "poor" category |
| **C** | Mindre alvorlig feil | Faults causing individual functions to not work as agreed, but the operator or customer can work around them relatively easily. | A single location has an outage in one or more channels or services; documentation is incomplete or imprecise |

**Service types** (tjenestetype): ruteinformasjonstjenester, salgstjenester, billetteringstjenester, betalingstjenester. This list is not exhaustive.

**Components**: salgskanaler, klienter, grensesnitt, integrasjon, grunndata, and any new components established later.

High-volume periods: incidents affecting single services during high-volume periods (e.g. campaigns or a new service/function launch) are escalated to **Nivå A** if the consequence is significant.

## SLA -- Operational Incidents (Operator Agreements)

| Avtale | Nivå | Communication requirements |
|--------|------|---------------------------|
| 4.3.5 | **A -- Kritisk** | Notify reporter within **30 minutes**. Work continuously until resolved or a satisfactory workaround is established. Affected parties informed immediately and kept updated. Status feedback to reporter within **2 hours** after case is closed or downgraded. |
| 4.3.6 | **B -- Alvorlig** | Feedback to reporter within **60 minutes** that work has started. Status feedback when resolved. |
| 4.3.7 | **C -- Mindre alvorlig** | **98.5%** of cases acknowledged within **2 business days**. Status feedback when resolved. |


## Detection

Incidents are typically detected through:

- **PagerDuty alerts** triggered by Prometheus AlertManager rules (error rate, latency, pod restarts, health check failures). See [observability.md](../reference/observability.md#alerting) for recommended alert thresholds.
- **Grafana dashboards** at `grafana.entur.org` -- anomalies in traffic, latency, or error rate. Prometheus metrics are stored long-term in Thanos for historical comparison.
- **Structured logs** in Google Cloud Logging -- error spikes, exception patterns. See [logging.md](../reference/logging.md) for correlation via `traceId` and `requestId`.
- **External reports** -- operator notifications, support channels for Jernbane and OMS, customer complaints.

## Process: Informere, Utføre, Etterarbeid

### 1. INFORMERE -- incident detected

**Utvikler / Utviklervakt:**

- Ring Incident Manager or Bakvakt
- Start a thread in Slack

**Incident Manager / Bakvakt:**

- Ring Support om varsling av Ops
- SMS operators and Entur leadership

**Support:**

- Ring Incident Manager or Bakvakt (if Support detected the incident)
- Ring Vy Drops, SJ OPS, and GoA Ops
- Varsle skiftleder KS

### 2. UTFØRE -- incident in progress

**Utvikler / Utviklervakt:**

- Analyse and handle the fault
- Update progress on fault correction in `#talk-incidents`
- Contact relevant team/developer for fixes

**Incident Manager / Bakvakt:**

- Lead the critical incident process
- Assist Utvikler/Utviklervakt with information
- Update incident handling status in `#talk-incidents`
- Document the incident timeline
- Put copies of SMS notifications in the Slack channel
- WPChat with the KS shift leader to ensure good two-way communication

**Support:**

- Answer all inbound queries related to the incident
- Keep OPS centres informed continuously via phone or email

**Investigation tooling:**

- Use distributed tracing in Cloud Trace to follow the failing request path. See [tracing.md](../reference/tracing.md).
- Check Prometheus metrics via Grafana for anomalies: error rate, latency percentiles, resource saturation. Thanos provides historical context for comparison.
- Correlate logs in Cloud Logging using `traceId` and `spanId`. See [logging.md](../reference/logging.md#correlation).
- Use Cloud Profiler to check for CPU or memory regression if the issue is performance-related. See [profiler.md](../reference/profiler.md).
- Apply the fix or rollback. All versions must support rollback to the previous version. See [architecture.md](../reference/architecture.md#microservice-principles).

### 3. ETTERARBEID -- incident resolved

**Utvikler / Utviklervakt:**

- Update and document the Post Mortem

**Incident Manager / Bakvakt:**

- Update status.entur.org
- Create a ticket in the **KHH project**
- Create the Post Mortem page

**Support:**

- Inform OPS centres and KS shift leader that the fault is resolved

### 4. Maintain the incident log

Throughout the incident, the Incident Manager maintains a timestamped log:

```text
[HH:MM] <author> -- <what happened / what was decided / what action was taken>
```

This log is the primary input for the post mortem. Use the "Mal - Logg under kritisk hendelse" template from the internal wiki.

### 5. Post mortem

Post mortems are blameless. The goal is systemic improvement, not individual fault.

The post mortem must include:

1. **Timeline** -- derived from the incident log
2. **Impact** -- users affected, duration, data implications
3. **Root cause** -- what failed and why
4. **Contributing factors** -- what made detection or resolution slower
5. **Action items** -- concrete, assigned, with deadlines. Examples:
   - Add a missing alert rule
   - Improve runbook for this failure mode
   - Fix the underlying bug
   - Add a circuit breaker or retry mechanism
6. **Lessons learned** -- what went well, what to improve in the process

Store post mortems and incident reports following the "Post Mortems og hendelsesrapporter" routines documented in the internal wiki.

## Escalation to Crisis Management (BLG)

Per Entur's Beredskapsplan: "Dersom krisestab etableres, skal BLG ha nødvendig ansvar og myndighet til å utøve god kriseledelse innad i og på vegne av Entur. Administrerende direktør, alternativt andre deltagere i ELG, beslutter om en krise/mulig krise defineres slik at det er nødvendig å etablere BLG."

### When to consider BLG

Each incident must be assessed individually. Incidents in these areas may lead to crisis management:

- **Økonomiske tap** -- significant financial loss
- **Omdømmetap** -- reputational damage
- **GDPR-brudd** -- personal data breach

Time of day matters: an incident at 03:00 is assessed differently from one at 09:00. BLG shall be involved and activated after **6 hours** per Entur's Beredskapsplan, but must be assessed on a case-by-case basis -- especially for GDPR, payment card, and security incidents where earlier activation may be warranted.

### How to activate BLG

1. IM/Bakvakt contacts **leder Digital** for guidance on next steps and whether BLG should be contacted
2. IM/Bakvakt sends SMS to the **Beredskapsledelse** group (see Kontaktliste Entur)

## Escalation to Virksomhetsstyring / Personvernleder

Incidents involving the following areas must **always** be escalated:

| Area | Escalate to | Notes |
|------|-------------|-------|
| **Persondata (GDPR)** | Personvernleder | Any breach of personal data |
| **Kortdata (betalingskort)** | Virksomhetsstyring | Separate procedures and notification deadlines to card acquirers apply. See SIP 009 in Sikkerhetsportalen. |
| **Sikkerhet** | Team Sikkerhet i Digital | Security incidents |

For all these incident types:

- **Continuously assess** whether a war room should be opened and whether to escalate to BLG (Beredskapsledelse)
- BLG shall be involved after 6 hours per the Beredskapsplan, but should be assessed earlier for these types
- **War room should be considered earlier** to quickly get to the core of the problem

## Escalation Summary

```text
Utvikler / Utviklervakt
  └─ rings ──> IM / Bakvakt (leads the incident)
       │
       ├─ GDPR breach ──────────> Personvernleder
       ├─ Payment card data ────> Virksomhetsstyring (see SIP 009)
       ├─ Security incident ────> Team Sikkerhet i Digital
       │
       └─ After 6 hours, or økonomiske tap / omdømmetap / GDPR-brudd:
            └─ IM/Bakvakt contacts Leder Digital
                 └─ IM/Bakvakt sends SMS to Beredskapsledelse ──> BLG
                      └─ Adm. dir. / ELG decides whether to establish krisestab
```

## Verify

- Reporter notified within 30 min (Nivå A) or 60 min (Nivå B) per SLA.
- `#talk-incidents` has status updates throughout the incident.
- Incident log captures the full timeline.
- Grafana dashboards confirm metrics returned to baseline after resolution.
- status.entur.org were updated at incident start, every 30 minutes and at resolution ("feil løst").
- GDPR, payment card, or security incidents were escalated to the correct party (Personvernleder / Virksomhetsstyring / Team Sikkerhet).
- BLG escalation was assessed at 6 hours, or earlier for GDPR/payment/security incidents.
- Post Mortem page created and ticket filed in the KHH project.
- Action items from the post mortem are tracked to completion.

## See also

- Observability standards: [observability.md](../reference/observability.md)
- Structured logging: [logging.md](../reference/logging.md)
- Distributed tracing: [tracing.md](../reference/tracing.md)
- Cloud Profiler: [profiler.md](../reference/profiler.md)
- Production hardening: [deploy-to-prd.md](deploy-to-prd.md)
