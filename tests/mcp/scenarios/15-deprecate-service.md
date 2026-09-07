# Scenario: Deprecating a Service

## Description

Tests an end-of-lifecycle question: when a team retires a service, what is
the documented teardown path? Surfaces the deprecate-service playbook.

## Prompt

Search the Entur knowledge base to answer:

Q: My team needs to deprecate and shut down an old Entur service. What steps does the platform expect us to follow, and what is the recommended order (Helm release, GCP project, repository)?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "deprecate"
  ],
  "must_not_contain": [
    "just delete the github repository",
    "rm -rf is sufficient"
  ],
  "must_match": [
    "guides_playbooks_deprecate-service_md|deprecate-service",
    "helm|kubernetes|gcp project|self-service"
  ]
}
```

## Budget

0.10
