# Scenario: Pre-approved IAM Roles

## Description

There is a curated allowlist of IAM roles at `guides/platform/iam-roles.md`.
Granting roles outside it requires platform team approval. The MCP must
surface this allowlist on a "which IAM roles can I use" query.

## Prompt

Search the Entur knowledge base to answer:

Q: Which Google Cloud IAM roles can I grant in my Terraform without asking the platform team for approval? Where is the canonical list documented?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "guides_platform_iam-roles_md",
    "iam"
  ],
  "must_not_contain": [
    "roles/owner",
    "roles/editor",
    "any role is fine",
    "no allowlist"
  ],
  "must_match": [
    "allowlist|approved list|allowed (list|roles)",
    "#talk-utviklerplattform|platform team"
  ]
}
```

## Budget

0.10
