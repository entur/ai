# Scenario: Critical Rule 9 - Conventional Commits

## Description

Tests that the agent formats a commit message using Conventional Commits (`type(scope): description`), not a free-text summary, and knows why it matters (automated semver via release-please).

## Prompt

You are helping an Entur developer who asks:

"I just fixed a bug where expired OAuth tokens weren't being refreshed in the auth client. What commit message should I use?"

Read the Entur AI documentation in this repository (start with AGENTS.md) and provide the exact commit message, plus a one-sentence explanation of why the format matters at Entur.

## Assertions

```json
{
  "must_contain": [
    "fix"
  ],
  "must_not_contain": [
    "Fixed a bug",
    "Fixed bug",
    "Bug fix:"
  ],
  "must_match": [
    "fix(\\([a-z0-9_-]+\\))?:\\s",
    "semver|semantic version|release-please"
  ]
}
```

## Budget

0.08
