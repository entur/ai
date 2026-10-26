# Scenario: Security Review Threat-Model Evidence Gates

## Description

Verifies that the `entur-ai-security-review` skill keeps its evidence gates when it models data flows or assesses privacy. Reachable, conditionally reachable, latent, and unverified flows stay distinct; unknown personal data stays unknown; a course payment provider is not invented; and dev evidence does not prove production. Ordinary processing can confirm an evidenced technical retention mismatch, while unknown purpose and unknown rights channels need validation. Effective deletion or upstream authorisation defeats the respective allegation. Personal data does not raise severity automatically, latent code is not currently exploitable, no GDPR conclusion is drawn, and refutation runs locally without unauthorised delegation. An ordinary-processing privacy candidate loads the privacy rules while the full view stays off.

## Prompt

You are applying the Entur security-review skill. Read `skills/entur-ai-security-review/SKILL.md`, then `skills/entur-ai-security-review/references/evidence-severity-reporting.md` and `skills/entur-ai-security-review/references/threat-modeling-output.md`. Use only those rules.

The facts below are SUPPLIED EVIDENCE for this exercise. They are not files in this repository. Treat them as complete for what they state, and as unknown for anything they omit.

Part 1, a requested DFD of one service:

- F1: `api/Routes.kt:12` registers the order route, and `config/dev.yaml:4` enables it. No production configuration was supplied.
- F2: `export/ExportJob.kt:30` writes to the export bucket, wired behind the flag `exportEnabled`, whose effective value is unknown.
- F3: `refund/RefundClient.kt:8` is implemented but not registered; a supplied release plan ships it next quarter.
- F4: `sync/PartnerSync.kt:15` exists; its registration, deployment, and data classes are unknown.
- A supplied course slide shows a payment provider; no code or configuration mentions one.

Part 2, an ordinary change-focused review with no STRIDE, LINDDUN, or DFD request. Delegation to sub-agents has not been authorised.

- Candidate A: a supplied privacy policy states 30 days. `config/retention.yaml:3` keeps identifiable traveller events for 365 days, and no purge job exists in the traced path.
- Candidate B: `analytics/Join.kt:20` joins encrypted trip records across purposes on a stable identifier; the permitted purpose and population are unknown.
- Candidate C: the reviewed diff has no delete endpoint; nothing is known about other rights channels.
- Candidate D: a suspected retention defect, but a supplied effective deletion trace shows the records are purged on schedule.
- Candidate E: an alleged IDOR on `GET /orders/{id}`, but `gateway/Policy.kt:44` enforces object ownership upstream on that route.
- Candidate F: an unregistered admin handler that would be High severity if activated.
- Candidate G: a handler returns traveller names, but its registration and access enforcement cannot be verified in scope.

Output exactly these lines, as plain text, with no bullets, markdown, or code fences. Pick one value from each list:

f1_status: <reachable|conditionally reachable|latent|unverified>
f1_production_proven: <yes|no>
f2_status: <reachable|conditionally reachable|latent|unverified>
f3_status: <reachable|conditionally reachable|latent|unverified>
f4_status: <reachable|conditionally reachable|latent|unverified>
f4_personal_data: <yes|no|unknown>
course_psp_in_model: <yes|no>
a_outcome: <confirmed|needs validation|rejected>
a_actor: <attacker|ordinary processing>
a_claim: <technical retention mismatch|unlawful processing>
b_outcome: <confirmed|needs validation|rejected>
c_outcome: <confirmed|needs validation|rejected>
d_outcome: <confirmed|needs validation|rejected>
e_outcome: <confirmed|needs validation|rejected>
f_label: <LATENT|reachable>
f_currently_exploitable: <yes|no>
g_auto_high: <yes|no>
g_outcome: <confirmed|needs validation|rejected>
legal_conclusion: <none|GDPR breach>
refutation: <local separate pass|delegated|skipped>
privacy_rules_loaded: <yes|no>
full_view: <on|off>

## Assertions

```json
{
  "must_not_contain": [
    "f4_personal_data: no",
    "course_psp_in_model: yes",
    "f1_production_proven: yes",
    "legal_conclusion: GDPR breach",
    "a_actor: attacker"
  ],
  "must_match": [
    "(?m)^f1_status: reachable$",
    "(?m)^f1_production_proven: no$",
    "(?m)^f2_status: conditionally reachable$",
    "(?m)^f3_status: latent$",
    "(?m)^f4_status: unverified$",
    "(?m)^f4_personal_data: unknown$",
    "(?m)^course_psp_in_model: no$",
    "(?m)^a_outcome: confirmed$",
    "(?m)^a_actor: ordinary processing$",
    "(?m)^a_claim: technical retention mismatch$",
    "(?m)^b_outcome: needs validation$",
    "(?m)^c_outcome: needs validation$",
    "(?m)^d_outcome: rejected$",
    "(?m)^e_outcome: rejected$",
    "(?m)^f_label: LATENT$",
    "(?m)^f_currently_exploitable: no$",
    "(?m)^g_auto_high: no$",
    "(?m)^g_outcome: needs validation$",
    "(?m)^legal_conclusion: none$",
    "(?m)^refutation: local separate pass$",
    "(?m)^privacy_rules_loaded: yes$",
    "(?m)^full_view: off$"
  ]
}
```

## Budget

0.10
