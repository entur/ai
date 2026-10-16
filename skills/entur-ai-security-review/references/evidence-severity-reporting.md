# Evidence, Severity, and Reporting

Use this reference to turn investigation leads into defensible findings and a
concise report. Severity is a risk judgement based on demonstrated impact and
exploitability, not a keyword, data type, or vote.

## Candidate evidence record

Record the following for every candidate before refutation:

- **ID and claim:** one precise statement of the violated control or attack.
- **Location:** exact file and line for the source, enforcement point, and sink.
- **Trace:** how identity, authority, or attacker-controlled data reaches the
  sensitive operation or disclosure.
- **Attacker and prerequisites:** anonymous, authenticated, tenant member,
  contributor, operator, compromised dependency, or another realistic actor.
- **Status:** reachable, conditionally reachable, latent, or unverified, with
  wiring and deployment evidence.
- **Impact:** affected confidentiality, integrity, availability, subjects,
  systems, data volume, and recoverability.
- **Mitigations checked:** validation, authorisation, framework defaults,
  gateway, IAM, feature flags, environment isolation, and operational controls.
- **Basis:** applicable Entur requirement or established security invariant.
- **Proposed severity and confidence:** separate risk from evidentiary certainty.
- **Smallest realistic fix and verification:** a stack-appropriate remediation
  and a test or observation that proves it works.

Never include a full credential, token, unnecessary personal data, or a weaponised
proof of concept in the record.

## Refutation outcomes

Use one of these outcomes:

- `confirmed`: evidence supports the path, impact, and severity.
- `downgraded`: the issue exists but reachability, prerequisites, or impact are
  narrower than first assessed.
- `rejected`: a mitigation or corrected trace defeats the claim.
- `needs validation`: the evidence required to decide is unavailable or would
  require access outside the authorised review.

Only confirmed and downgraded issues appear in severity totals. Put `needs
validation` items in their own section and omit rejected candidates.

## Severity calibration

Use repository-specific policy if it defines a scale. Otherwise apply this
calibration:

| Severity | Evidence-based meaning |
|----------|------------------------|
| Critical | A practical path with low prerequisites to catastrophic impact, such as broad sensitive-data disclosure, production control, account takeover, remote code execution, or prolonged loss of a critical service. A live credential is Critical only when its privileges and exposure support comparable impact. |
| High | A practical path to serious cross-user or cross-tenant compromise, sensitive-data exposure, privilege escalation, or material service disruption, but with meaningful prerequisites, narrower scope, or stronger recovery than Critical. |
| Medium | A credible, material weakness with limited impact or one that normally needs chaining, special positioning, or another failure to cause serious harm. |
| Low | A narrow hardening or hygiene issue with difficult prerequisites and limited direct impact. |

Do not automatically raise severity because personal data appears. Evaluate its
sensitivity, volume, subjects, exposure, attacker, and operational context. Do
not automatically raise severity because a route is undocumented; documentation
drift changes assurance and may be a separate issue, but technical severity still
follows the attack path.

For latent code, state the severity expected if activated and label the status
`LATENT`; never describe it as currently exploitable. Omit dead code unless it
has a credible path to shipping or the misleading claim is independently harmful.

## Confidence calibration

- **High:** the reviewer traced the complete path and verified deployment or
  effective configuration.
- **Medium:** code evidence is strong, but a runtime, platform, or deployment fact
  remains unverified.
- **Low:** the concern is a hypothesis or depends on unavailable evidence; place
  it under `Needs validation`, not confirmed findings.

## Merge and reporting rules

- Merge locations that share one root cause, trust boundary, exploit path, and
  fix. Keep issues separate when attacker prerequisites, impact, boundary, or
  remediation differ.
- Treat documentation drift as supporting evidence for the code defect. Report it
  separately only when it remains harmful after the implementation is fixed.
- Do not count missing tests, suspicious dependencies, absent optional headers,
  or pattern hits as vulnerabilities without a violated control or credible path.
- The lead reviewer owns severity, de-duplication, and the final wording. Preserve
  unresolved disagreement as uncertainty rather than averaging reviewers' votes.
- Record verified-sound controls only when positively checked, and state their
  scope. Absence of a finding is not evidence that an unreviewed surface is safe.

## Report template

Use findings first. Omit empty sections except the summary and coverage boundary.

```markdown
# AI-Assisted Security Review: <service>

Date: YYYY-MM-DD
Scope: <commit, diff, components, or repository>
Profile: <change-focused, targeted, or repository triage>
Method: <manual traces, static searches, delegated surfaces, refutation>

## Summary

<Confirmed totals, urgent exposure status, and one sentence on the observed posture within scope.>

## Findings

### [HIGH] <Specific outcome an attacker can cause>

- **Where:** `path/file.ext:line`
- **Status:** reachable | conditionally reachable | latent
- **Confidence:** high | medium
- **Attacker and prerequisites:** <who and what they need>
- **Evidence:** <concise source-to-control-to-sink trace>
- **Impact at Entur:** <systems, people, data, or availability affected>
- **Mitigations checked:** <what was ruled out>
- **Fix:** <smallest realistic remediation>
- **Verify:** <test or observation that proves the fix>

## Needs validation

- <Unresolved concern, missing evidence, and how to validate it safely.>

## Verified sound

- <Control checked, evidence, and precise scope.>

## Coverage and residual risk

- **Reviewed:** <entrypoints, trust boundaries, controls, and configuration.>
- **Not reviewed:** <files, environments, scanners, runtime behavior, or formal assessments.>

## Recommended next actions

1. <Urgent containment or release blocker, if any.>
2. <Prioritised remediation or deeper review.>
```

If a live credential or personal-data exposure is suspected, place a short alert
before the normal report. Redact sensitive values, avoid further exploitation,
and recommend immediate handling through established Entur processes. Do not
claim a legal breach determination or contact others without authorisation.
