# MCP Knowledge-Base Quality Tests

Black-box tests for the [`entur-kompass` MCP server](https://github.com/entur/kompass) and the
markdown documents in this repo that feed it. Each scenario sends a natural-language
question through `mcp__entur-kompass__search_entur_kb`, then asserts on (a) which docs
the MCP surfaced and (b) whether Claude can extract the right answer from those
docs alone.

The companion suite in [`../scenarios/`](../scenarios/) tests how AI agents read
the markdown files directly via `Read`/`Grep`/`Glob`. **This suite is the opposite
end of the pipeline**: it lets Claude use *only* the MCP, so failures point at
either retrieval quality (the wrong docs came back) or content gaps (the right
docs came back but did not contain the answer).

## What it tests

- **Retrieval quality** -- does the right document end up in the top results
  for a typical engineer question?
- **Document coverage** -- once the right doc is surfaced, does it actually
  contain the answer the question implies?
- **Doc consistency** -- when two docs cover the same topic, do they agree on
  things like file extensions (`.yaml` vs `.yml`), module refs, or naming patterns?
- **Gap detection** -- a `must_contain` assertion that fails on a well-formed
  query is a documentation gap signal, not a flaky test.

## Quick start

```bash
# From the tests/ directory (this suite reuses the same Go runner as ../scenarios/)
cd tests

# Dry run -- parse scenarios, no API calls
./mcp/run.sh --dry-run

# Full MCP suite
./mcp/run.sh --verbose

# Filter to a single scenario
./mcp/run.sh --scenario "03-*" --verbose
```

`run.sh` is a thin wrapper that invokes `go run .` with the right flags:

```text
--dir mcp/scenarios
--allowed-tools "mcp__entur-kompass__search_entur_kb,mcp__entur-kompass__report_feedback"
--system-prompt "<MCP-only prompt -- forbids Read/Grep/Glob and prior knowledge>"
```

You can pass any of the runner's flags through to the wrapper
(`--model`, `--budget`, `--strict`, `--junit FILE`, etc -- see
[`../README.md`](../README.md) for the full list).

## Prerequisites

1. `claude` CLI installed (`npm install -g @anthropic-ai/claude-code`).
2. The `entur-kompass` MCP added to your Claude CLI config:

   ```bash
   claude mcp add --scope user --transport http entur-kompass https://ki.entur.io/mcp
   ```

   First invocation opens a browser for Google sign-in; cached after that.
3. Go 1.25+ (`cd tests && go run .`).

## Scenario format

Identical to the [parent suite](../README.md#adding-a-new-scenario):

```markdown
# Scenario: Short Title

## Description
What this query is probing. Note any known doc gaps or inconsistencies.

## Prompt
A natural-language question, plus the boilerplate that forces structured output
(see existing scenarios for the canonical template).

## Assertions
{
  "must_contain": ["doc id or key fact"],
  "must_not_contain": ["common wrong answer"],
  "must_match": ["regex"]
}

## Budget
0.10
```

### Prompt template

Each scenario uses the same scaffold so assertions can be precise:

```text
Search the Entur knowledge base to answer:

Q: <natural-language question>

Use the mcp__entur-kompass__search_entur_kb tool to research this. Do not call any
other tool. Do not use prior knowledge of Entur.

Then output EXACTLY this format and nothing else:

top_results: <id1>, <id2>, <id3>
answer: <1-3 sentence answer drawn ONLY from the snippets / extractive text>

Each `id` must be copied verbatim from the MCP response
(e.g. `guides_platform_self-service_md`, `AGENTS_md`). If the MCP results
don't contain enough info, write `answer: insufficient` -- do not fabricate.
```

The `top_results:` line lets assertions target document IDs like
`guides_platform_self-service_md`. The `answer:` line lets assertions target
key facts (`ent-{appid}-{env}`, `terraform-google-sql-db`, etc).

The template intentionally does **not** ask Claude to call
`mcp__entur-kompass__report_feedback`. An earlier version did, and the
behavior was that Claude would call the feedback tool as its last action and
then end the turn with an empty `result` field (no final assistant text),
which made every scenario "fail" with empty output. Tests should not pollute
the production feedback log anyway.

### Doc ID convention

The MCP returns IDs derived from the file path: `/` and `.` become `_`. Examples:

| File | Doc ID |
|------|--------|
| `AGENTS.md` | `AGENTS_md` |
| `CONVENTIONS.md` | `CONVENTIONS_md` |
| `guides/platform/self-service.md` | `guides_platform_self-service_md` |
| `guides/playbooks/add-postgres.md` | `guides_playbooks_add-postgres_md` |
| `skills/setup-cicd-workflows/SKILL.md` | `skills_setup-cicd-workflows_SKILL_md` |

## Reading non-failures

A `PASS` does not always mean the docs fully cover the topic. Scenario 16
(`16-full-stack-java-with-domain.md`) is a deliberate kitchen-sink question
that asks for app + Postgres + Redis + Kafka + custom-domain setup. The
assertions pass when the well-covered sub-topics are present, but the prompt
also asks Claude to split its answer into `answer_bootstrap`, `answer_runtime`,
and `answer_domain` and to write `insufficient` for any sub-topic the KB does
not cover. Read the raw output (use `--verbose` and inspect with a manual
`claude -p`, or temporarily relax the runner to print on pass) to surface
sub-topic gaps that won't show up as red.

As of writing, `answer_domain` reliably reports the gap: `common.ingress.host`
and the `*.entur.io` hostname pattern are visible in a `common-helm.md`
snippet, but no playbook explains the human side (request flow, DNS, TLS
certificate, platform-team handoff). That is the documentation gap this test
exists to keep in front of us.

## Reading failures

| Failure shape | Likely cause |
|---------------|--------------|
| Doc ID missing from `top_results:` | Retrieval-quality problem: the query wording does not match the doc well. Either reword the query in the test (if it is unrepresentative) or improve the doc's lead paragraph / headings. |
| Doc ID present but answer-level `must_contain` fails | Documentation gap: the right doc was surfaced but it does not actually contain the fact the question implies. Add the fact to the surfaced doc. |
| `must_not_contain` fails | The MCP surfaced a doc that contradicts another doc, or Claude pulled an old example out of snippet text. Resolve the inconsistency in the source markdown. |
| Top result is a `skills/*` doc that contradicts a `guides/*` doc | Same topic covered in two places with divergent details. Make the skill defer to the guide. |

## Cost

Each MCP scenario costs roughly the same as a parent-suite scenario (~$0.03--0.08
with Haiku, since the MCP call adds modest tokens beyond a single Read). The
suite of ~15 scenarios runs in ~$0.80 with retries enabled. Cap with
`--budget 1.50` if running cold.
