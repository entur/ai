# Scenario: GCP Projects Must Come from Self-Service, Not Terraform

## Description

Critical rule #11: never create GCP projects with `google_project` Terraform
resource or `gcloud projects create`. Always via self-service YAML in `.entur/`.
This is one of the most-violated rules by naive AI agents.

## Prompt

Search the Entur knowledge base to answer:

Q: I want to create a new GCP project at Entur. Should I write a Terraform `google_project` resource, run `gcloud projects create`, or use something else?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "self-service",
    ".entur/",
    "googlecloudapplication"
  ],
  "must_not_contain": [
    "use google_project",
    "use the terraform google_project resource",
    "run gcloud projects create"
  ],
  "must_match": [
    "guides_platform_self-service_md|self-service",
    "never (use terraform|use google_project|create.*via terraform|run gcloud)"
  ]
}
```

## Budget

0.10
