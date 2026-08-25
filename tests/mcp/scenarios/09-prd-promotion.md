# Scenario: Promoting a Build to Production

## Description

The cd.yml image-promotion model: PR-built Docker image is tagged in git,
promoted to tst and prd via `resolve-image`. Tests that the KB surfaces the
deploy-to-prd playbook and explains the tag-based promotion (not rebuild).

## Prompt

Search the Entur knowledge base to answer:

Q: How do I promote a build that has been running successfully in `dev` to the `tst` and `prd` environments at Entur? Does the image get rebuilt, or is the same image reused?

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "git tag",
    "prd"
  ],
  "must_not_contain": [
    "rebuild the image for each environment",
    "rebuild in prd",
    "kubectl apply by hand",
    "manual docker push"
  ],
  "must_match": [
    "guides_playbooks_deploy-to-prd_md|deploy-to-prd",
    "promote|resolve-image|same image|reused"
  ]
}
```

## Budget

0.10
