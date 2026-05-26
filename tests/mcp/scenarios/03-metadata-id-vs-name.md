# Scenario: metadata.id vs metadata.name Disambiguation

## Description

The single most common source of AI agent confusion at Entur. `metadata.id`
drives GCP projects, Helm `shortname`, Terraform `app_id`, state bucket.
`metadata.name` becomes the Kubernetes namespace and typically matches the repo
name. If the KB cannot teach this difference, every other derivation breaks.

## Prompt

Search the Entur knowledge base to answer:

Q: In an Entur self-service manifest, what is the difference between `metadata.id` and `metadata.name`? Which one drives GCP project naming, and which one becomes the Kubernetes namespace?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "metadata.id",
    "metadata.name",
    "namespace",
    "gcp"
  ],
  "must_not_contain": [
    "metadata.name drives gcp",
    "metadata.id becomes the kubernetes namespace"
  ],
  "must_match": [
    "metadata\\.id.*(gcp|project)",
    "metadata\\.name.*(namespace|kubernetes)"
  ]
}
```

## Budget

0.10
