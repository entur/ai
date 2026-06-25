# Scenario: Kafka via Aiven and the Entur Starter

## Description

Entur runs Kafka on Aiven and provides a Spring Boot starter
(`entur-kafka-spring-boot-starter`). Tests that the KB surfaces the Kafka
playbook and the starter, not generic Apache Kafka guidance.

## Prompt

Search the Entur knowledge base to answer:

Q: I need a Spring Boot service to produce and consume Kafka events at Entur. What is the recommended setup -- which Kafka cluster, and is there a Spring starter I should use?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "aiven",
    "kafka"
  ],
  "must_not_contain": [
    "confluent cloud",
    "msk",
    "self-hosted kafka",
    "redpanda"
  ],
  "must_match": [
    "guides_platform_entur-kafka-starter_md|entur-kafka-starter|guides_playbooks_add-kafka_md",
    "starter|spring-boot-starter|entur-kafka"
  ]
}
```

## Budget

0.10
