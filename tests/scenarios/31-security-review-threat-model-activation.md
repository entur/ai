# Scenario: Security Review Threat-Model Activation

## Description

Verifies that the `entur-ai-security-review` skill activates threat-model output only as far as the request needs. Ordinary reviews emit light threat-category tags with `DFD not modelled` and no full view; an unexamined privacy dimension is `not assessed`, not `none justified`. Repository triage enables the full view without another permission pause. STRIDE-only and DFD-only requests stay scoped, and a formal threat-model request gets the truthful limitation. A personal-data disclosure is one finding with both tags, a separate retention root cause stays separate, and a logging conflict records both requirements without inventing a retention period or disabling required audit.

Changed expectation: the earlier design expected `taxonomy_output: omitted` for ordinary reviews. Ordinary reviews now carry light tags, but still no full view.

## Prompt

You are applying the Entur security-review skill. Read `skills/entur-ai-security-review/SKILL.md`, then `skills/entur-ai-security-review/references/evidence-severity-reporting.md` and `skills/entur-ai-security-review/references/threat-modeling-output.md`. Use only those rules.

The facts below are SUPPLIED EVIDENCE for this exercise. They are not files in this repository. Treat them as already verified.

- Case A: a change-focused review of one diff, with no STRIDE, LINDDUN, or DFD request. A confirmed missing query bound in `search/Handler.kt:40` lets any caller exhaust the journey-search service. Nobody traced whether its request logs keep traveller identifiers.
- Case B: the user asks for repository triage of a whole service.
- Case C: the user asks for "STRIDE only for `OrderHandler`".
- Case D: the user asks for "a DFD only for the changed ingress; no privacy review".
- Case E: the user asks for "a formal threat model to sign off the release".
- Case F: one confirmed object-authorisation defect returns other travellers' order data; one fix closes it.
- Case G: separately, a ticket store keeps identifiable records past its supplied retention policy; that has its own fix.
- Case H: partner refund actions lack audit events, and the proposed audit payload would copy traveller movement data.

Output exactly these lines, as plain text, with no bullets, markdown, or code fences. Pick one value from each list:

a_full_view: <on|off>
a_dfd: <not modelled|modelled>
a_stride: <Denial of service|none justified|not assessed>
a_linddun: <Data disclosure|none justified|not assessed|n/a: evidenced non-personal>
b_full_view: <on|off>
b_permission_pause: <yes|no>
c_linddun_pass: <yes|no>
d_taxonomy_pass: <yes|no>
e_response: <formal threat model|supporting triage with limitation>
f_findings: <1|2>
f_stride: <Information disclosure|none justified>
f_linddun: <Data disclosure|none justified>
g_retention: <merged with F|separate finding>
h_tension: <record both requirements|drop audit|log full payload>
h_retention_period: <invented|team to decide>
h_disable_audit: <yes|no>

## Assertions

```json
{
  "must_not_contain": [
    "a_linddun: none justified",
    "a_full_view: on",
    "f_findings: 2",
    "h_disable_audit: yes",
    "a_dfd: modelled",
    "a_stride: none justified",
    "a_stride: not assessed",
    "a_linddun: Data disclosure",
    "a_linddun: n/a: evidenced non-personal",
    "b_full_view: off",
    "b_permission_pause: yes",
    "c_linddun_pass: yes",
    "d_taxonomy_pass: yes",
    "e_response: formal threat model",
    "f_stride: none justified",
    "f_linddun: none justified",
    "g_retention: merged with F",
    "h_tension: drop audit",
    "h_tension: log full payload",
    "h_retention_period: invented"
  ],
  "must_match": [
    "(?m)^a_full_view: off$",
    "(?m)^a_dfd: not modelled$",
    "(?m)^a_stride: Denial of service$",
    "(?m)^a_linddun: not assessed$",
    "(?m)^b_full_view: on$",
    "(?m)^b_permission_pause: no$",
    "(?m)^c_linddun_pass: no$",
    "(?m)^d_taxonomy_pass: no$",
    "(?m)^e_response: supporting triage with limitation$",
    "(?m)^f_findings: 1$",
    "(?m)^f_stride: Information disclosure$",
    "(?m)^f_linddun: Data disclosure$",
    "(?m)^g_retention: separate finding$",
    "(?m)^h_tension: record both requirements$",
    "(?m)^h_retention_period: team to decide$",
    "(?m)^h_disable_audit: no$"
  ]
}
```

## Budget

0.10
