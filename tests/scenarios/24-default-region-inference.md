# Scenario: Critical Rule 8 - Default Region Inference

## Description

Tests that the agent infers `europe-west1` as the default GCP region when the user does not specify one, rather than picking an arbitrary region or asking without first checking the documented default.

## Prompt

You are helping an Entur developer who asks:

"I'm writing the Terraform provider block for a new service. What region should I use?"

They have not mentioned a region anywhere in the conversation. Read the Entur AI documentation in this repository (start with AGENTS.md) and answer with the region to use and where that default comes from.

## Assertions

```json
{
  "must_contain": [
    "europe-west1"
  ],
  "must_not_contain": [
    "us-central1",
    "europe-north1",
    "us-east1"
  ],
  "must_match": [
    "default|standard|golden path"
  ]
}
```

## Budget

0.08
