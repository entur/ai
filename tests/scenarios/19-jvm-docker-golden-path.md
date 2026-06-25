# Scenario: JVM Docker Golden Path

## Description

Verifies Java/Kotlin services use the distroless Java runtime image as the Golden Path, not Liberica.

## Prompt

You are creating a Dockerfile for a new Kotlin Spring Boot service at Entur.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the Docker and Java reference guides) to answer.

Output exactly these keys:

- runtime_base_image: <recommended runtime image>
- alternative_runtime_image: <allowed alternative if shell/debugging tools are needed>
- not_golden_path: <image family that is not the Golden Path>

## Assertions

```json
{
  "must_contain": [
    "gcr.io/distroless/java25-debian13:nonroot",
    "eclipse-temurin:25-jre-alpine"
  ],
  "must_match": [
    "not_golden_path.*Liberica|not_golden_path.*liberica"
  ],
  "must_not_contain": [
    "bellsoft/liberica-runtime-container:jre-25-cds-slim-musl"
  ]
}
```

## Budget

0.08
