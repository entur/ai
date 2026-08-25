# Scenario: Permission Store and Permission Client

## Description

Entur has a centralized Permission Store for fine-grained authorization,
accessed via the Permission Client. Tests that the KB surfaces this concept
clearly when an engineer asks about role/permission management.

## Prompt

Search the Entur knowledge base to answer:

Q: I need fine-grained authorization in my Entur service (per-user, per-resource permissions). What is the Permission Store, and what is the Permission Client used for?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "permission store",
    "permission client"
  ],
  "must_not_contain": [
    "spring security only",
    "iam roles in gcp",
    "auth0 rbac is sufficient"
  ],
  "must_match": [
    "guides_platform_permission-store_md|permission-store",
    "authorization|permission|access control"
  ]
}
```

## Budget

0.10
