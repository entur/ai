# Scenario: Framing - Health Check Endpoints (Positive)

## Description

Tests whether positive framing leads to clean, correct health endpoint recommendations without mentioning wrong paths.

## Prompt

You are given the following platform convention:

> ALWAYS configure Go services with these health check endpoints: /health/liveness for Kubernetes liveness probes and /health/readiness for Kubernetes readiness probes.

A developer asks: "What health check endpoints should my Go service expose for Kubernetes probes?"

List the exact paths. Answer in 1-2 sentences based ONLY on the convention above. Do not read any repository files.

## Assertions

```json
{
  "must_contain": [
    "/health/liveness",
    "/health/readiness"
  ],
  "must_not_contain": [],
  "must_match": [
    "/health/liveness.*readiness|/health/readiness.*liveness"
  ]
}
```

## Budget

0.02
