# Scenario: Tracing IAM Role Is Write-Only

## Description

Verifies the agent grants exactly roles/cloudtrace.agent to the workload service
account, not a broader trace role.

## Prompt

What IAM role should be bound to the workload service account so it can export
traces to Cloud Trace, per Entur's tracing golden path?

Read `skills/setup-tracing/SKILL.md` in this repository and answer in `key: value`
format on its own line:

- role: <role>

## Assertions

```json
{
  "must_contain": [
    "role: roles/cloudtrace.agent"
  ],
  "must_not_contain": [
    "role: roles/cloudtrace.user",
    "role: roles/cloudtrace.admin",
    "role: roles/cloudtrace.editor"
  ]
}
```

## Budget

0.08
