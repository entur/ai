# Scenario: Data Project Naming (int vs ext)

## Description

Data projects use a different naming pattern from application projects:
`ent-data-{id}-{int|ext}-{env}`. Tests that the KB surfaces this nuance and
that the `int`/`ext` distinction (driven by `spec.dataAccess.external`) is
explained.

## Prompt

Search the Entur knowledge base to answer:

Q: For a `GoogleCloudDataProject` with `metadata.id: routes` that needs to share data externally, what GCP project IDs are created? How does the naming differ from an internal-only data project?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "guides_platform_self-service_md",
    "ent-data-routes",
    "ext"
  ],
  "must_not_contain": [
    "ent-routes-dev",
    "ent-routes-prd",
    "ent-data-routes-dev"
  ],
  "must_match": [
    "ent-data-routes-ext-(dev|tst|prd)",
    "dataaccess.*external|spec\\.dataaccess"
  ]
}
```

## Budget

0.10
