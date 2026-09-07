# Scenario: Running the Service Locally

## Description

Tests that the local-dev playbook is reachable and that the KB explains the
basic local-run loop (mise tools, docker-compose dependencies if any).

## Prompt

Search the Entur knowledge base to answer:

Q: How do I run an Entur Spring Boot service on my laptop during development -- what tool manages the JDK and other dependencies, and how do I bring up its Postgres dependency?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "mise"
  ],
  "must_not_contain": [
    "sdkman is required",
    "brew install java is the only option"
  ],
  "must_match": [
    "guides_playbooks_local-dev_md|local-dev",
    "docker[- ]?compose|testcontainers|cloud sql proxy"
  ]
}
```

## Budget

0.10
