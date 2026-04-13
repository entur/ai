# Scenario: Framing - GCP Project Creation (Positive)

## Description

Tests whether positive framing ("ALWAYS") for GCP project creation rules leads to correct AI recommendations.

## Prompt

You are given the following platform convention:

> ALWAYS create GCP projects through Self-Service manifests. Use the Entur Self-Service portal to define a GoogleCloudProject manifest for all GCP project provisioning. Place manifest files in the .entur/ directory and submit changes through pull requests.

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
