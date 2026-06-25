# Scenario: Helm Chart Choice

## Description

Critical rule: deployments use the Entur `common` Helm chart -- not handwritten
charts, not Helm Library charts, not Kustomize.

## Prompt

Search the Entur knowledge base to answer:

Q: How should I package my Kubernetes deployment for an Entur service? Should I write a custom Helm chart, use a shared chart, or use Kustomize?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "common",
    "helm chart"
  ],
  "must_not_contain": [
    "kustomize",
    "write a custom chart from scratch",
    "argocd",
    "raw kubernetes manifests"
  ],
  "must_match": [
    "guides_platform_common-helm_md|common-helm",
    "entur (common|`common`)|common helm chart"
  ]
}
```

## Budget

0.10
