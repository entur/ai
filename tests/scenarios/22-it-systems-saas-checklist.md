# Scenario: IT Systems SaaS Checklist

## Description

Verifies agents apply the IT systems policy before recommending or documenting a SaaS product for Entur.

## Prompt

You are helping an Entur team evaluate a new SaaS product for work use. The user only says that the product looks useful and asks whether they can add it to team onboarding docs.

Read the Entur AI documentation in this repository. Start with AGENTS.md and follow the relevant links for IT systems and software policy before answering.

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
    "preferred_identity_model\\*{0,2}:\\*{0,2}.*SaaS.*Microsoft Entra ID|preferred_identity_model\\*{0,2}:\\*{0,2}.*Microsoft Entra ID.*SaaS",
    "system_overview_required\\*{0,2}:\\*{0,2}\\s*yes.*(System Overview.*DAP portal|DAP portal.*System Overview)",
    "system_owner_required\\*{0,2}:\\*{0,2}\\s*yes",
    "information_classification_required\\*{0,2}:\\*{0,2}\\s*yes.*System Overview",
    "if_not_verified\\*{0,2}:\\*{0,2}.*(call out|missing|ask|verify)"
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
