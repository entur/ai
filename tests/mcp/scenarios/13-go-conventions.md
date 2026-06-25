# Scenario: Go Service Conventions

## Description

Tests Go-specific conventions: distroless runtime image, standard health
endpoints, structured logging.

## Prompt

Search the Entur knowledge base to answer:

Q: I am writing a new Go HTTP service at Entur. What Docker runtime base image should I use, what health endpoint paths are expected, and what logging style is required?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "distroless",
    "8080"
  ],
  "must_not_contain": [
    "alpine runtime",
    "ubuntu runtime",
    "println(",
    "fmt.println(\"log:"
  ],
  "must_match": [
    "guides_reference_go_md|reference/go",
    "gcr\\.io/distroless|distroless/static",
    "structured (logging|logs)|json log|slog"
  ]
}
```

## Budget

0.10
