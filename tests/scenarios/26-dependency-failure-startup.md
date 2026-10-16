# Scenario: Graceful Startup When a Dependency Is Unavailable

## Description

Tests that the agent knows applications must start gracefully even when a downstream dependency (e.g. a database, Redis, or another service) is temporarily unreachable, rather than crash-looping or blocking indefinitely.

## Prompt

You are helping an Entur developer who asks:

"My service fails to start with a fatal error if Redis isn't reachable yet. Is that the right behavior?"

Read the Entur AI documentation in this repository (start with AGENTS.md, then CONVENTIONS.md) and answer whether that is correct, and what the application should do instead.

## Assertions

```json
{
  "must_contain": [
    "gracefully"
  ],
  "must_not_contain": [
    "yes, that is correct",
    "yes, that's correct",
    "that's the right behavior"
  ],
  "must_match": [
    "not.*correct|should not|shouldn't|isn't right|not the right|incorrect|wrong",
    "start.*(gracefully|even (if|when))|retry|reconnect|readiness"
  ]
}
```

## Budget

0.08
