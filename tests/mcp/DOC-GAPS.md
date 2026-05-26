# MCP Suite -- Documentation Improvement Backlog

Findings from the most-recent full run of `tests/mcp/run.sh`. Each item lists
the scenario(s) that surfaced it, classifies it as a **content gap** (the doc
is missing or incomplete) or a **retrievability gap** (the doc has the answer
but the MCP snippet does not capture it), names the file to edit, and
suggests a concrete fix.

Status legend: `[ ]` open, `[x]` fixed, `[~]` partial.

## High priority (real content gaps)

### [ ] H1. Custom domain setup playbook is missing

- **Surfaced by:** scenario 16 (`16-full-stack-java-with-domain.md`),
  `answer_domain: insufficient`.
- **Symptom:** Engineers can find `common.ingress.host` and example
  `*.entur.io` hostnames in `guides/platform/common-helm.md`, but there is
  no end-to-end playbook for the human side of standing up a custom domain.
- **What is missing:**
  - Who to request the domain from (Slack channel, ticket flow).
  - When `trafficType: public` vs `api` applies.
  - Hostname pattern per environment (`<app>.dev.entur.io`,
    `<app>.entur.io`).
  - TLS certificate handling (managed certs? platform-team action?).
  - DNS record provisioning (does the platform do it, or the team?).
  - How to verify the domain end-to-end (`curl`, browser, etc).
- **Fix:** Create `guides/playbooks/add-custom-domain.md` modelled on
  `add-postgres.md` / `add-redis.md`. Link it from
  `CLAUDE.md`'s "By goal (start here)" table and from the
  `## Networking → Ingress` section of `common-helm.md`.

## Medium priority (retrievability -- restructure so MCP snippets pick up key terms)

The MCP returns the correct doc but the extracted snippet often misses the
specific phrasing engineers (and our assertions) look for. Fixing these is
usually a 1-3 line edit in the lead paragraph or a heading rename.

### [ ] M1. `security.md` -- "never hardcode" doesn't reach the snippet

- **Surfaced by:** scenario 06 (`06-secret-management.md`),
  `must_match: /never hardcode|do not hardcode|must not hardcode/` failed.
- **Root cause:** `guides/reference/security.md:13` literally says
  "**Never hardcode secrets**", but the MCP snippet for the "where do I
  store secrets" query picks up downstream sections (`### Creating
  Secrets (Terraform)`, etc) rather than the rules at the top.
- **Fix:** Lift the "Never hardcode" rule out of a bullet and make it the
  first sentence of `## Secret Management` (currently the H2 heading is
  followed by a sub-heading `### Rules` and then a bullet list -- the
  snippet extractor often drops the rules in favour of the next code
  block). One-line lead like:
  > Never hardcode secrets in source code, config files, Dockerfiles, or
  > CI workflows -- always use Google Secret Manager + ExternalSecrets.

### [ ] M2. `self-service.md` -- `.entur/` directory not in snippet

- **Surfaced by:** scenario 12 (`12-self-service-not-terraform.md`),
  `must_contain: ".entur/"` failed.
- **Root cause:** `guides/platform/self-service.md:3` says "Define YAML
  manifests in `.entur/` and apply through a GitOps PR workflow", but for
  the "should I use Terraform or something else?" query the MCP returned
  snippets from the project-bootstrap skill and from later sections
  instead.
- **Fix:** In `self-service.md`, repeat the `.entur/` path in the
  `## How It Works` section's first numbered step (e.g.
  `Create/modify YAML manifests in .entur/<appid>.yaml.`). Add an explicit
  "NOT Terraform" sentence near the top so a single snippet captures both
  the right answer and the contrast.

### [ ] M3. `iam-roles.md` -- "allowlist" terminology not in headings

- **Surfaced by:** scenario 07 (`07-iam-roles-allowlist.md`),
  `must_match: /allowlist|approved list|allowed (list|roles)/` failed.
- **Root cause:** The H2 heading is `## Allowed roles` but the lead
  paragraph says "These are the IAM roles that CD service accounts are
  allowed to grant..." -- the MCP snippet picks up other phrasings first.
  Engineers and assertions use the word **allowlist** which never appears
  in the doc.
- **Fix:** Rename `## Allowed roles` to `## Allowed roles (allowlist)` and
  add "This page is the authoritative allowlist of IAM roles..." as the
  first sentence of the doc.

### [ ] M4. `local-dev.md` -- "docker compose" / "testcontainers" terms thin

- **Surfaced by:** scenario 14 (`14-local-dev.md`),
  `must_match: /docker[- ]?compose|testcontainers|cloud sql proxy/` failed.
- **Root cause:** `guides/playbooks/local-dev.md` says "Docker Compose for
  dependencies" in the lead, but the answer Claude produced from the
  snippets says "`compose.yaml`" without ever mentioning the tool by name.
  The doc itself never mentions `testcontainers` (which is what most teams
  actually use for integration tests).
- **Fix:** Two small additions:
  1. In step 2, change "Add a `compose.yaml`..." to "Add a `compose.yaml`
     (run with `docker compose up`)...".
  2. Add a one-line note: "For integration tests, prefer Testcontainers
     (link to `guides/reference/java.md` / `go.md` sections)" so the term
     is retrievable.

## Low priority (small precision wins)

### [ ] L1. `go.md` -- standard port 8080 not in Go-specific health section

- **Surfaced by:** scenario 13 (`13-go-conventions.md`),
  `must_contain: "8080"` failed for a question that asked about base image,
  health paths, and logging style.
- **Note:** `guides/reference/go.md:50` already has `EXPOSE 8080` in the
  Dockerfile example, and line 96 has `envDefault:"8080"`. But neither
  surfaces when the question is about conventions rather than Dockerfile
  content.
- **Fix:** Add a one-liner in `## Health Checks`: "Services bind to port
  8080 (Entur convention; matches the common Helm chart default)."

### [ ] L2. `common-helm.md` ingress section lacks a runnable end-to-end example near the host setting

- **Surfaced by:** scenario 16's `answer_domain` (the bit that did get
  picked up was the snippet from later in the file). The ingress section
  (around line 181) shows `trafficType` but the `host` setting only
  appears much later in the "Complete Example" section.
- **Fix:** Add a `host:` example directly under `### Ingress` showing
  `host: <app>.dev.entur.io` and `host: <app>.entur.io`, so the snippet
  that surfaces for ingress queries also surfaces the hostname pattern.
  Cross-link to the new `add-custom-domain.md` playbook from H1 once it
  exists.

## Not a doc gap (test-side fixes)

These also showed up as failures but are assertion precision, not doc
problems. Track in the suite, not here:

- Scenario 08 (`08-helm-common-chart.md`):
  `must_not_contain: "kustomize"` is wrong -- the question itself mentions
  Kustomize, so Claude correctly says "do not use Kustomize". The
  assertion should be removed; doc is fine.
- Scenarios 06, 07, 12, 13, 14 also have minor regex tightness that
  contributes to the failures. Once M1-M4 land, those assertions will
  match the doc snippets naturally and can stay strict.

## How to re-verify

After fixing any item, run the affected scenario(s) to confirm the
snippet now surfaces the desired phrasing:

```bash
cd tests
./mcp/run.sh --scenario "06-*" --verbose --no-retry      # one scenario
./mcp/run.sh --verbose --no-retry --parallel 4           # full suite
```

The retrievability fixes (M1-M4) should each move an assertion from FAIL
to PASS. The content gap (H1) needs a follow-up scenario added once the
playbook exists.
