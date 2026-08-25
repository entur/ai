# Scenario: IT Systems SSO and Licences

## Description

Verifies agents do not recommend local user administration, shared accounts, or extra licence purchases when the IT systems policy requires SSO and licence reuse checks.

## Prompt

An Entur team is adopting an internal tool. The tool can run with local username/password accounts, and the team says it is faster to buy five new licences than to check whether existing licences are unused.

Read the Entur AI documentation in this repository. Start with AGENTS.md and follow the relevant links for IT systems and software policy before answering.

Output exactly these keys:

- local_user_admin_default: <yes/no>
- required_identity_provider: <identity provider>
- no_sso_action: <who must be contacted>
- buy_new_licences_first: <yes/no>
- licence_action: <what to check or do first>

## Assertions

```json
{
  "must_contain": [
    "local_user_admin_default",
    "Microsoft Entra ID",
    "System Owner",
    "DAP",
    "buy_new_licences_first",
    "inactive licences"
  ],
  "must_match": [
    "local_user_admin_default\\*{0,2}:\\*{0,2}\\s*no",
    "required_identity_provider\\*{0,2}:\\*{0,2}.*Microsoft Entra ID",
    "no_sso_action\\*{0,2}:\\*{0,2}.*System Owner.*DAP|no_sso_action\\*{0,2}:\\*{0,2}.*DAP.*System Owner",
    "buy_new_licences_first\\*{0,2}:\\*{0,2}\\s*no",
    "licence_action\\*{0,2}:\\*{0,2}.*((remove|reassign|check).*inactive licences|inactive licences.*(remove|reassign|check))"
  ],
  "must_not_contain": [
    "shared accounts are fine",
    "local users by default",
    "buy new licences first"
  ]
}
```

## Budget

0.08
