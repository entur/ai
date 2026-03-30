# Scenario: Go Service Conventions

## Description

Verifies that the agent uses Go-specific conventions (health paths, Docker images, metrics) and does NOT use Spring Boot patterns for a Go service.

## Prompt

You are setting up a new Go service at Entur.

Details:

- Repository name: stop-lookup
- Team: team-data-platform
- App ID (metadata.id): stoplkup
- Language: Go (not Java/Kotlin)
- No database needed

Read the Entur AI documentation in this repository (start with AGENTS.md, then read go.md and docker.md) to answer.
Output each answer in `key: value` format on its own line:

- liveness_path: <liveness probe path>
- readiness_path: <readiness probe path>
- docker_base_image: <Docker runtime base image>
- prometheus_path: <Prometheus metrics path>
- gcp_project_dev: <GCP project ID for dev>
- gcp_project_prd: <GCP project ID for prd>
- helm_shortname: <Helm shortname value>

## Assertions

```json
{
  "must_contain": [
    "/health/liveness",
    "/health/readiness",
    "ent-stoplkup-dev",
    "ent-stoplkup-prd"
  ],
  "must_not_contain": [
    "/actuator/health/liveness",
    "/actuator/health/readiness",
    "/actuator/prometheus",
    "spring",
    "temurin",
    "liberica"
  ],
  "must_match": [
    "distroless.*static|static.*distroless",
    "prometheus_path.*/metrics",
    "helm_shortname.*stoplkup"
  ]
}
```

## Budget

0.10
