# Scenario: CI/CD Workflow File List and Extension

## Description

The setup-cicd-workflows skill canonically uses `.yml` for GitHub Actions
workflow files. This scenario probes for the right list and extension; failure
here can flag drift between the workflow skill and reference docs.

## Prompt

Search the Entur knowledge base to answer:

Q: For a new Kotlin Spring Boot service that has a Helm chart and Terraform, list the GitHub Actions workflow files that must be created under `.github/workflows/`. Give the exact file names.

Use the mcp__entur-kb__search_entur_kb tool to research this. Do not call any other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <list the file names, drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response (e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results don't contain enough info, write `answer: insufficient` -- do not fabricate.

## Assertions

```json
{
  "must_contain": [
    "ci.yml",
    "cd.yml",
    "build.yml",
    "pr.yml",
    "codeql.yml",
    "dependabot-pr.yml",
    "terraform.yml",
    "terraform-drift-detection.yml"
  ],
  "must_not_contain": [
    "ci.yaml",
    "cd.yaml",
    "build.yaml",
    "pr.yaml",
    "codeql.yaml",
    "dependabot-pr.yaml",
    "terraform.yaml",
    "terraform-drift-detection.yaml",
    "deploy.yml",
    "deploy.yaml"
  ],
  "must_match": [
    "skills_setup-cicd-workflows_SKILL_md|guides_platform_gha-workflows_md"
  ]
}
```

## Budget

0.10
