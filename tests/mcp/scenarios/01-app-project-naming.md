# Scenario: GCP Project Naming for a Standard Application

## Description

Probes the most-asked question at Entur: what GCP project name do I get from
my self-service manifest? Validates that the MCP surfaces `self-service.md` and
that it actually documents the `ent-{appid}-{env}` pattern.

## Prompt

Search the Entur knowledge base to answer:

Q: A team sets `metadata.id: products` in a `GoogleCloudApplication` self-service manifest. What GCP project IDs will the Platform Orchestrator create for the `dev`, `tst`, and `prd` environments?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "guides_platform_self-service_md",
    "ent-products-dev",
    "ent-products-tst",
    "ent-products-prd"
  ],
  "must_not_contain": [
    "ent-products-prod",
    "ent-products-test",
    "ent-products-staging",
    "products-dev.entur.io"
  ],
  "must_match": [
    "ent-products-(dev|tst|prd)"
  ]
}
```

## Budget

0.10
