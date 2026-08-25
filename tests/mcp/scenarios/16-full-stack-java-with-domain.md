# Scenario: Full-stack Java service on GKE with Postgres, Redis, Kafka, and a custom domain

## Description

The most-asked compound question by new teams: how to stand up a Java Spring
Boot service on Entur's GKE platform with the three most-common managed
dependencies (Cloud SQL Postgres, Memorystore Redis, Aiven Kafka), AND how to
expose it on a custom `*.entur.io` hostname.

This is intentionally broad. Claude may make several MCP searches to cover
each sub-topic. Assertions are split into:

- **Bootstrap + infra** (well-documented): self-service manifest, Entur
  Terraform modules, common Helm chart, Kafka starter.
- **Domain setup** (thin coverage today): `common.ingress.host`,
  `*.entur.io` pattern. There is no dedicated "set up a custom domain"
  playbook in the indexed docs as of writing -- if this scenario flags a
  gap there, that is the signal to create one.

## Prompt

Search the Entur knowledge base to answer the following compound question. You may call mcp__entur-kompass__search_entur_kb up to FIVE times with different queries to cover all sub-topics. Do not call any other tool. Do not use prior knowledge of Entur.

Q: I am setting up a brand-new Java Spring Boot service at Entur that will run on Google Kubernetes Engine. It needs a managed PostgreSQL database, a Memorystore Redis instance, and Kafka topics for event streaming. It also needs to be reachable on a custom hostname under `entur.io`. Walk me through:

1. How GCP projects are created (manifest kind and filename, naming pattern).
2. Which Entur Terraform modules provision Postgres and Redis.
3. How Kafka is provided and what Spring Boot dependency to add.
4. Which Helm chart to use to deploy the service.
5. How to expose the service on a custom domain (Helm value to set, hostname pattern, and any platform-team action required).

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>, <id4>, <id5>
answer_bootstrap: <2-3 sentences on items 1-2 above>
answer_runtime: <2-3 sentences on items 3-4 above>
answer_domain: <2-3 sentences on item 5; if the KB does not explain step-by-step how to request and wire up a custom domain, say so explicitly>

Each `id` must be copied verbatim from the MCP responses (e.g. `guides_platform_self-service_md`, `AGENTS_md`). Include the most relevant IDs across all searches you ran. If a sub-question cannot be answered from the snippets, write `insufficient` for that line -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "googlecloudapplication",
    "terraform-google-sql-db",
    "terraform-google-memorystore",
    "aiven",
    "common",
    "ingress"
  ],
  "must_not_contain": [
    "amazon rds",
    "elasticache",
    "confluent cloud",
    "msk",
    "route53",
    "cloudflare",
    "self-host the kafka cluster"
  ],
  "must_match": [
    "guides_platform_self-service_md|self-service",
    "guides_playbooks_add-postgres_md|add-postgres|terraform-google-sql-db",
    "guides_playbooks_add-redis_md|add-redis|terraform-google-memorystore",
    "guides_playbooks_add-kafka_md|add-kafka|entur-kafka-(spring-boot-)?starter",
    "guides_platform_common-helm_md|common-helm|common helm chart",
    "common\\.ingress\\.host|ingress:\\s*\\n\\s*host:|ingress\\.host",
    "entur\\.io|\\.entur\\.io"
  ]
}
```

## Budget

0.20
