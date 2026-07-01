# Scenario: Java Agent Is the Default for Java/Kotlin, Not Go

## Description

Verifies the agent defaults a Kotlin/Java service to the OpenTelemetry Java Agent
without a specific reason to hand-instrument, AND correctly does NOT recommend the
Java Agent for a Go service in the same prompt -- confirming it understands the
language scope of the rule rather than pattern-matching "just get tracing working"
to one fixed answer.

## Prompt

Two engineers ask for tracing, no specific reason for manual instrumentation given
by either:

1. "We need tracing on our Spring Boot service, nothing fancy, just get it working."
2. "We need tracing on our Go service, nothing fancy, just get it working."

Read `skills/setup-tracing/SKILL.md` in this repository and answer in `key: value`
format on its own line:

- spring_boot_approach: <java agent/manual sdk>
- go_approach: <java agent/manual sdk>

## Assertions

```json
{
  "must_contain": [
    "spring_boot_approach: java agent",
    "go_approach: manual sdk"
  ],
  "must_not_contain": [
    "spring_boot_approach: manual sdk",
    "go_approach: java agent"
  ]
}
```

## Budget

0.08
