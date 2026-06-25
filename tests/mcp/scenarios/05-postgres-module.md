# Scenario: Postgres Provisioning Module

## Description

Verifies the KB surfaces the `terraform-google-sql-db` module + the Cloud SQL
proxy sidecar pattern, and connects the answer to the add-postgres playbook.

## Prompt

Search the Entur knowledge base to answer:

Q: How do I provision a managed PostgreSQL database for an Entur service? Which Terraform module should I use, and how does the application connect to the database?

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "terraform-google-sql-db",
    "cloud sql",
    "proxy",
    "localhost:5432"
  ],
  "must_not_contain": [
    "google_sql_database_instance",
    "alloydb",
    "amazon rds",
    "public ip"
  ],
  "must_match": [
    "guides_playbooks_add-postgres_md|guides_platform_terraform-modules_md",
    "common\\.postgres\\.enabled|cloud sql proxy|cloud sql auth proxy"
  ]
}
```

## Budget

0.10
