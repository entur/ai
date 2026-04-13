# Scenario: Framing - Docker Base Image (Negative)

## Description

Tests whether negative framing for Docker image selection causes mention of forbidden images (pink elephant effect).

## Prompt

You are given the following platform convention:

> Don't use alpine images for Go services. Do not use ubuntu or debian full images as base. Never use images that include a shell or package manager. The correct runtime image is gcr.io/distroless/static-debian12:nonroot - don't use alpine or any other base image.

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
