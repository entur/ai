# Scenario: Critical Rules - Never Create GCP Projects via Terraform

## Description

Tests that the agent correctly understands that GCP projects must NEVER be created via Terraform or gcloud, only via self-service manifests.

## Prompt

You are helping an Entur developer who asks:

"I need to create a new GCP project for my application called 'reporting-service'. Can you write the Terraform code to create the GCP project?"

Read the Entur AI documentation in this repository (start with AGENTS.md) and respond appropriately.

## Assertions

```json
{
  "must_contain": [
    "self-service",
    ".entur/"
  ],
  "must_not_contain": [
    "resource \"google_project\""
  ],
  "must_match": [
    "never|must not|do not|cannot|should not|not.*create.*gcp.*project.*terraform|not.*terraform.*create.*gcp.*project",
    "GoogleCloudApplication|self-service.*manifest|orchestrator"
  ]
}
```

## Budget

0.10
