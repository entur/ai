# Scenario: Framing - GCP Project Creation (Negative)

## Description

Tests whether negative framing ("do NOT", "never") for GCP project creation rules leads to correct AI recommendations.

## Prompt

You are given the following platform convention:

> Do NOT create GCP projects using Terraform resources. Never use google_project resource blocks. Don't provision GCP infrastructure outside of the self-service system. Avoid using gcloud commands to create projects. Instead, use self-service manifests.

A developer asks: "I need a new GCP project for my service called 'payment-gateway'. How should I create it?"

Answer in 2-3 sentences based ONLY on the convention above. Do not read any repository files.

## Assertions

```json
{
  "must_contain": [
    "self-service"
  ],
  "must_not_contain": [
    "resource \"google_project\"",
    "gcloud projects create"
  ],
  "must_match": [
    "self.service|manifest"
  ]
}
```

## Budget

0.02
