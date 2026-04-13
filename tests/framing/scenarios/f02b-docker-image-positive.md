# Scenario: Framing - Docker Base Image (Positive)

## Description

Tests whether positive framing for Docker image selection leads to a clean, correct recommendation.

## Prompt

You are given the following platform convention:

> ALWAYS use gcr.io/distroless/static-debian12:nonroot as the Docker base image for Go services. This provides a minimal attack surface, runs as non-root by default, and contains only the static binary runtime.

A developer asks: "What Docker base image should I use for my production Go microservice?"

Recommend the specific image with full path and tag. Answer in 1-2 sentences based ONLY on the convention above. Do not read any repository files.

## Assertions

```json
{
  "must_contain": [
    "distroless",
    "nonroot"
  ],
  "must_not_contain": [],
  "must_match": [
    "gcr\\.io/distroless/static"
  ]
}
```

## Budget

0.02
