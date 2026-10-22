# Scenario: Enable Freyja in My Repo

## Description

Verifies agents do not try to enable Freyja themselves. The `freyja-mode` custom property is set manually in the GitHub UI by a repository admin, so the agent gives the manual steps instead of writing files or running commands.

## Prompt

You are an AI coding agent working in the Entur repository `entur/my-service`. The developer asks:

"Enable Freyja in my repo."

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the set-up-freyja playbook) to answer.

Output exactly these keys:

- agent_can_enable: <yes/no>
- files_to_change: <list of files, or none>
- manual_steps: <short numbered steps>
- mode_to_choose: <mode>

## Assertions

```json
{
  "must_contain": [
    "agent_can_enable: no",
    "freyja-mode"
  ],
  "must_not_contain": [
    "freyja.yml",
    "freyja.yaml"
  ],
  "must_match": [
    "files_to_change:\\W*none",
    "Settings.*Custom properties",
    "mode_to_choose:.*advise"
  ]
}
```

## Budget

0.10
