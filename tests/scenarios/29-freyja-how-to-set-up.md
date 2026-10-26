# Scenario: How Do I Set Up Freyja

## Description

Verifies agents explain how to set up Freyja: through the `freyja-mode` custom property, starting in `advise`, and going through the auto-merge checklist before choosing `auto-merge`. Freyja is not configured through a file in the repository.

## Prompt

You are helping an Entur developer who asks:

"How do I set up Freyja for Dependabot PRs?"

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the set-up-freyja playbook) to answer.

Output exactly these keys:

- setting_name: <name of the setting>
- setting_location: <where in GitHub the setting is changed>
- who_can_change_it: <role>
- first_mode: <mode to start with>
- before_auto_merge: <one sentence>
- config_file_in_repo: <yes/no>

## Assertions

```json
{
  "must_contain": [
    "config_file_in_repo: no"
  ],
  "must_not_contain": [
    "freyja.yml",
    "freyja.yaml"
  ],
  "must_match": [
    "setting_name:\\W*freyja-mode",
    "first_mode:\\W*advise",
    "setting_location:.*custom propert",
    "who_can_change_it:.*admin",
    "before_auto_merge:.*checklist"
  ]
}
```

## Budget

0.10
