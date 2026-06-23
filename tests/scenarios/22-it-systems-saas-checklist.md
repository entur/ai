# Scenario: IT Systems SaaS Checklist

## Description

Verifies agents apply the IT systems policy before recommending or documenting a SaaS product for Entur.

## Prompt

You are helping an Entur team evaluate a new SaaS product for work use. The user only says that the product looks useful and asks whether they can add it to team onboarding docs.

Read the Entur AI documentation in this repository (start with AGENTS.md, then read it-systems-policy.md) to answer.

Output exactly these keys:

- preferred_identity_model: <one sentence>
- system_overview_required: <yes/no and where>
- system_owner_required: <yes/no>
- information_classification_required: <yes/no and where>
- if_not_verified: <what the agent should say or do>

## Assertions

```json
{
  "must_contain": [
    "Microsoft Entra ID",
    "System Overview",
    "DAP portal",
    "System Owner",
    "classified"
  ],
  "must_match": [
    "preferred_identity_model:.*SaaS.*Microsoft Entra ID|preferred_identity_model:.*Microsoft Entra ID.*SaaS",
    "system_overview_required:\\s*yes.*System Overview.*DAP portal",
    "system_owner_required:\\s*yes",
    "information_classification_required:\\s*yes.*System Overview",
    "if_not_verified:.*(call out|missing|ask|verify)"
  ],
  "must_not_contain": [
    "approved by default",
    "no registration needed",
    "skip classification"
  ]
}
```

## Budget

0.08
