# Scenario: Tracking Issue Required for Disabled Tests

## Description

Tests that the agent knows a `@Disabled`/`@Ignore` test must link a tracking issue explaining why, and does not just silently disable a flaky test.

## Prompt

You are helping an Entur developer who asks:

"This Kotlin test is flaky in CI. Can you just add @Disabled to it so the build goes green?"

Read the Entur AI documentation in this repository (start with AGENTS.md, then CONVENTIONS.md) and answer what they should actually do.

## Assertions

```json
{
  "must_contain": [
    "issue"
  ],
  "must_not_contain": [
    "sure, just add @Disabled",
    "yes, add @Disabled to it"
  ],
  "must_match": [
    "@Disabled\\(.*(issue|ticket|link|reason)|link.*(issue|ticket)|(issue|ticket).*link",
    "track|reference|explain"
  ]
}
```

## Budget

0.08
