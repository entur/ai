# Scenario: Where to Store Application Secrets

## Description

Critical rule #1 from CLAUDE.md: secrets live in Google Secret Manager,
surfaced via ExternalSecrets in Helm. Never hardcoded, never in repo files.

## Prompt

Search the Entur knowledge base to answer:

Q: Where should I store database passwords and API keys for an Entur service? How are they made available to the running pod?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "secret manager",
    "externalsecret"
  ],
  "must_not_contain": [
    "github secrets",
    "hashicorp vault",
    "aws secrets manager",
    "hardcode",
    "checked into the repo",
    ".env file in the repo"
  ],
  "must_match": [
    "google secret manager|gcp secret manager|secretmanager",
    "never hardcode|do not hardcode|must not hardcode"
  ]
}
```

## Budget

0.10
