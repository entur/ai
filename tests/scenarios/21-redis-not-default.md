# Scenario: Redis Is Not Default

## Description

Verifies agents do not add Redis infrastructure or dependencies unless the service has a concrete Redis use case.

## Prompt

You are bootstrapping a new Java Spring Boot service at Entur. It needs PostgreSQL and REST endpoints, but it has no shared cache, session storage, rate limiting, distributed lock, or Kafka deduplication requirement.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the Java reference guide and add-redis playbook) to answer.

Output exactly these keys:

- add_redis: <yes/no>
- reason: <one sentence>
- add_dependency: <yes/no>
- add_memorystore_module: <yes/no>

## Assertions

```json
{
  "must_contain": [
    "add_redis: no",
    "add_dependency: no",
    "add_memorystore_module: no"
  ],
  "must_match": [
    "Do not add Redis by default|only when the service has a concrete"
  ]
}
```

## Budget

0.08
