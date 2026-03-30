# Scenario: Firebase Application

## Description

Verifies that Firebase applications use GoogleCloudFirebaseApplication kind and follow the standard ent-{id}-{env} naming (not a special prefix like data projects).

## Prompt

You are setting up a new Firebase web application at Entur.

Details:

- Team: team-partner
- App ID (metadata.id): prtnrprt
- Kind: GoogleCloudFirebaseApplication
- Firebase region: europe-west1
- Environments: dev, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read self-service.md) to answer.
Output each answer in `key: value` format on its own line:

- gcp_project_dev: <GCP project ID for dev>
- gcp_project_prd: <GCP project ID for prd>

Then show the complete self-service application manifest YAML.

## Assertions

```json
{
  "must_contain": [
    "ent-prtnrprt-dev",
    "ent-prtnrprt-prd",
    "GoogleCloudFirebaseApplication",
    "orchestrator.entur.io/apps/v1",
    "id: prtnrprt",
    "europe-west1"
  ],
  "must_not_contain": [
    "ent-data-prtnrprt",
    "ent-firebase-prtnrprt",
    "GoogleCloudApplication",
    "GoogleCloudDataProject"
  ],
  "must_match": [
    "firebase[\\s\\S]*region"
  ]
}
```

## Budget

0.10
