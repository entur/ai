# Scenario: Data Project Naming

## Description

Verifies the special naming convention for GoogleCloudDataProject with external data access: ent-data-{id}-{int|ext}-{env}.

## Prompt

You are setting up a new data project at Entur for the analytics team.

Details:

- Team: team-analyse
- App ID (metadata.id): akt
- Kind: GoogleCloudDataProject
- External data sharing: yes (spec.dataAccess.external: true)
- Environments: dev, tst, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read self-service.md) to answer.
Output each answer in `key: value` format on its own line, then show the full manifest YAML:

- gcp_project_dev: <GCP project ID for dev>
- gcp_project_tst: <GCP project ID for tst>
- gcp_project_prd: <GCP project ID for prd>
- apiVersion: <apiVersion value>

Then show the complete self-service manifest YAML.

## Assertions

```json
{
  "must_contain": [
    "ent-data-akt-ext-dev",
    "ent-data-akt-ext-prd",
    "orchestrator.entur.io/apps/v1",
    "GoogleCloudDataProject"
  ],
  "must_not_contain": [
    "ent-akt-dev",
    "ent-akt-prd",
    "ent-data-akt-dev",
    "GoogleCloudApplication"
  ],
  "must_match": [
    "metadata[.\\s]*id.*akt",
    "external.*true"
  ]
}
```

## Budget

0.10
