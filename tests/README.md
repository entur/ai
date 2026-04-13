# AI Documentation Comprehension Tests

Automated tests that verify AI agents correctly understand Entur's platform documentation. Each test scenario sends a prompt to Claude, lets it read the docs, and validates the response against expected patterns.

## Quick Start

```bash
# From the repository root:

# Dry run -- validate scenarios parse correctly, no API calls
go run ./tests --dry-run

# Run all tests (~$0.35, ~90 seconds)
go run ./tests --verbose

# Run a specific scenario
go run ./tests --scenario "05-*" --verbose

# Use a different model
go run ./tests --model sonnet
```

## What It Tests

| Scenario | Tests |
|----------|-------|
| 01-kotlin-api | Full identity chain: metadata.id to GCP projects, Helm, Terraform, Docker |
| 02-go-service | Go-specific conventions: health paths, distroless image, metrics path |
| 03-data-project | Data project naming: `ent-data-{id}-{int\|ext}-{env}` pattern |
| 04-firebase-app | Firebase naming: standard `ent-{id}-{env}`, not special prefix |
| 05-derive-from-manifest | Core test: distinguish metadata.id from metadata.name |
| 06-critical-rules | Never create GCP projects via Terraform |
| 07-cicd-file-structure | CI/CD pipeline file naming and architecture (all 8+ workflow files) |
| 08-cicd-ci-workflow | ci.yaml reusable workflow: lint, test (Java 25/Gradle), Docker build/scan/push |
| 09-cicd-build-workflow | build.yaml PR trigger and pr.yaml verification workflow |
| 10-cicd-cd-workflow | cd.yaml image promotion model: resolve-image via git tag, deploy chain |
| 11-cicd-security-workflows | codeql.yaml (CodeQL + Java config) and dependabot-pr.yaml (approval gate) |
| 12-cicd-terraform-workflows | terraform.yaml (lint/plan/apply) and drift detection (weekly schedule) |
| 13-cicd-dependabot-config | dependabot.yml: correct ecosystems for Kotlin (github-actions, docker, gradle) |
| 14-cicd-go-variant | Go CI: setup-go, go test, no artifact upload, no test-reporter, gomod dependabot |
| 15-cicd-no-helm | No Helm: ci.yaml omits helm-lint, cd.yaml omits deploy jobs |
| 16-cicd-python-variant | Python CI: setup-python, pytest, no artifact upload, pip dependabot |

## CLI Options

```text
go run ./tests [OPTIONS]

  --scenario PATTERN   Glob pattern to filter scenarios (e.g. "01-*")
  --model MODEL        Claude model: haiku (default), sonnet, opus
  --budget FLOAT       Total budget cap in USD (default: 1.00)
  --junit FILE         Write JUnit XML report to FILE
  --verbose            Print full Claude responses for failed scenarios
  --dry-run            Parse scenarios, print commands, no API calls
  --strict             Require 100% assertion pass rate (default: 80%)
  --no-retry           Disable automatic retry on failure
  --dir PATH           Scenario directory (default: auto-detected)
```

## How It Works

1. The runner reads `.md` scenario files from `tests/scenarios/`
2. For each scenario, it invokes `claude -p "<prompt>" --output-format json --model haiku`
3. Claude reads the repo docs via `Read`, `Grep`, `Glob` tools (read-only)
4. The runner parses Claude's JSON response and validates against assertions
5. Failed scenarios are retried once (non-deterministic AI output)

### Pass Criteria

- **Normal mode** (default): all `must_not_contain` pass + 80% of positive assertions pass
- **Strict mode** (`--strict`): 100% of all assertions must pass

## Adding a New Scenario

Create a markdown file in `tests/scenarios/` following this format:

````markdown
# Scenario: My Test Name

## Description

What this test verifies (human-readable, not used by runner).

## Prompt

The exact prompt sent to Claude. Tell it to read the docs
and ask specific, answerable questions.

## Assertions

```json
{
  "must_contain": ["exact strings that must appear"],
  "must_not_contain": ["strings that must NOT appear"],
  "must_match": ["regex.*patterns"]
}
```

## Budget

0.10
````

### Assertion Types

| Type | Matching | Use For |
|------|----------|---------|
| `must_contain` | Case-insensitive substring | Exact identifiers: GCP project names, bucket names |
| `must_not_contain` | Case-insensitive substring | Common mistakes: wrong naming patterns, forbidden resources |
| `must_match` | Regex with `(?i)` flag | YAML values with quoting variations: `shortname.*jpapi` |

### Tips

- Keep prompts focused -- ask numbered questions, request "ONLY the answers"
- Use `must_not_contain` to catch the specific mistakes you've seen AI agents make
- Use `must_match` regex for values that may appear in different YAML formats
- Set `## Budget` to `0.10` for most scenarios (default is `0.08`)
- Test your scenario with `--scenario "your-*" --verbose` first

## Cost

Each scenario costs ~$0.03-0.08 with Haiku. Full suite of 16 scenarios: ~$1.00.
With retries (worst case): ~$2.00. Increase budget cap with `--budget 2.50` if running all.

## Unit Tests

The scenario parser and assertion logic have their own Go tests:

```bash
cd tests && go test -v
```
