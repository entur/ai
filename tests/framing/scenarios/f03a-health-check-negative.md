# Scenario: Framing - Health Check Endpoints (Negative)

## Description

Tests the pink elephant effect: mentioning /healthz in a negative rule may cause the AI to include it in its answer.

## Prompt

You are given the following platform convention:

> Do not use /healthz for health check endpoints. Don't use /health as a single endpoint. Never use /ready or /live as separate paths. The correct paths are /health/liveness and /health/readiness - not /healthz, not /ready, not /live.

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
