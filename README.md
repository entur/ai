# Entur AI

Hey! Welcome to Entur's shared AI resources.

This is where we keep everything that helps AI agents and assistants work better at Entur -- from coding standards that agents use to generate platform-compliant code, to reusable skills that supercharge your day-to-day work.

> **Heads up:** The files in `guides/` are written primarily for AI agents to consume, not humans. They're dense, structured for machine readability, and intentionally light on code examples (the AI figures out the implementation from your project's own codebase). This is still very much a work in progress -- things will change, improve, and expand over time. It works well today, but we're iterating!

## What's in this repo

| Folder    | Purpose                                                        |
|-----------|----------------------------------------------------------------|
| `guides/` | Coding standards and platform conventions for AI coding agents |
| `skills/` | Reusable AI skills for day-to-day work across teams            |
| `tests/`  | Comprehension tests that verify AI agents understand the docs  |

**Coding agent instructions** help agents like Claude Code, GitHub Copilot, and Cursor generate code that follows Entur's platform conventions. Instead of every team maintaining their own copy of "how we do things at Entur", we keep it here and everyone points their agents to it.

**Skills** are reusable instruction sets that give AI assistants specialized capabilities -- like drafting architecture decision records, summarizing Slack threads, or navigating Entur-specific tooling. See [`skills/README.md`](skills/README.md) for the full catalogue and how to use them.

## Quick Start

Create an `AGENTS.md` file in your repository root that points the AI agent to this repository:

```markdown
# My Application

Java 25 / Spring Boot application that provides the route planning API.

## Entur Standards

Read and follow the Entur platform standards at:
https://github.com/entur/ai/blob/main/AGENTS.md

When working on a specific task, also read the relevant guides
linked from that file (e.g. java.md, helm.md, docker.md).

## Project-Specific

- Uses PostgreSQL via Cloud SQL
- Publishes events to Pub/Sub topic `route-updates`
- Custom health indicator for external route provider connectivity
```

That's it. `AGENTS.md` is read automatically by GitHub Copilot and [many other agents](https://agents.md). Claude Code reads `CLAUDE.md` instead -- see [Agent Compatibility](#agent-compatibility) for details. The agent will fetch the linked URL to get the full platform standards.

### Tips for a good `AGENTS.md`

- **Describe your application** in the first few lines -- language, framework, what it does
- **Link to the shared standards** so the agent knows Entur conventions
- **Add project-specific context** -- database, messaging, special patterns, team conventions
- **List any overrides** if your project deviates from the shared standards

### Example for a Go service

```markdown
# My Go Service

Go 1.25 service that processes transit data feeds.

## Entur Standards

Read and follow the Entur platform standards at:
https://github.com/entur/ai/blob/main/AGENTS.md

## Project-Specific

- Uses Cloud Pub/Sub for event processing
- Stores processed data in BigQuery
- No external API -- internal consumer only
```

## Agent Compatibility

| Agent          | Reads `AGENTS.md`       | Can fetch URLs | Notes                                                                        |
| -------------- | ----------------------- | -------------- | ---------------------------------------------------------------------------- |
| Claude Code    | No (reads `CLAUDE.md`)  | Yes            | Natively reads `CLAUDE.md`; create a symlink or copy for Claude Code support |
| GitHub Copilot | Yes                     | Limited        | Reads `AGENTS.md`; may not fetch URLs in all modes                           |
| opencode       | Yes                     | Unknown        | Reads `AGENTS.md` natively                                                   |

`AGENTS.md` is supported by a [large ecosystem of AI coding agents](https://agents.md) including Codex, Gemini CLI, Jules, Windsurf, Aider, and many more.

Claude Code reads `CLAUDE.md`, not `AGENTS.md`. To support Claude Code alongside other agents, create a symlink: `ln -s AGENTS.md CLAUDE.md`.

For agents that cannot fetch URLs, the most important rules are already inline in your project's `AGENTS.md`. For deeper coverage, you can copy key sections from this repo into your project's instructions.

## Documentation Structure

```text
AGENTS.md                       # Top-level agent routing and critical rules
CONVENTIONS.md                  # Cross-language coding conventions
guides/
  java.md                       # Java standards (Spring Boot, Gradle)
  kotlin.md                     # Kotlin standards
  go.md                         # Go standards
  docker.md                     # Containerization with Docker
  api-design.md                 # REST and gRPC API design
  architecture.md               # Service and infrastructure architecture
  logging.md                    # Structured logging
  observability.md              # Health checks, metrics, tracing
  security.md                   # Secrets, scanning, IAM
  code-review.md                # Review checklist
  helm.md                       # Entur common Helm chart reference
  self-service.md               # Self-service provisioning, manifests, onboarding
  markdown.md                   # Markdown standards and linting
  terraform/
    modules.md                  # Terraform modules (init, SQL, Redis, GCS)
    iam-roles.md                # Approved IAM roles
  cicd/
    workflows.md                # CI/CD pipeline reference
    actions.md                  # Composite actions reference
skills/
  README.md                     # Skill catalogue, usage guide, and how to contribute
  entur-project-bootstrap/      # Bootstrap a new app (self-service, Helm, TF, Docker, CI/CD)
  setup-cicd-workflows/         # Generate CI/CD workflows by language
tests/
  README.md                     # Test usage guide and how to add scenarios
  main.go                       # Test runner (Go, stdlib only)
  scenario.go                   # Scenario parser and assertion evaluator
  scenario_test.go              # Unit tests for the parser
  scenarios/                    # Test scenarios (one .md file per test)
```

AI agents read `AGENTS.md` first, which routes them to the relevant sub-documents based on the task.

## Shared Tooling Referenced

| Category              | Tools                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Terraform modules** | [terraform-google-init](https://github.com/entur/terraform-google-init), [terraform-google-sql-db](https://github.com/entur/terraform-google-sql-db), [terraform-google-memorystore](https://github.com/entur/terraform-google-memorystore), [terraform-google-cloud-storage](https://github.com/entur/terraform-google-cloud-storage)                                                                                                                                              |
| **Helm charts**       | [helm-charts](https://github.com/entur/helm-charts) (common chart)                                                                                                                                                                                                                                                                                                                                                                                                                  |
| **CI/CD workflows**   | [gha-docker](https://github.com/entur/gha-docker), [gha-helm](https://github.com/entur/gha-helm), [gha-terraform](https://github.com/entur/gha-terraform), [gha-security](https://github.com/entur/gha-security), [gha-meta](https://github.com/entur/gha-meta), [gha-firebase](https://github.com/entur/gha-firebase), [gha-docs](https://github.com/entur/gha-docs), [gha-slack](https://github.com/entur/gha-slack), [gha-artifactory](https://github.com/entur/gha-artifactory) |

## Contributing

This is a shared resource for all of Entur, and we'd love your help making it better! Every contribution matters -- whether it's fixing a typo, clarifying a confusing section, adding a new skill, or sharing a pattern that works well for your team.

A few ways to contribute:

- **Found something wrong or unclear?** Open an issue or just submit a PR directly
- **Have a pattern or skill that works great for your team?** Share it! Others will benefit
- **Not sure if something belongs here?** Open an issue and let's figure it out together
- **Want to improve the AI output for your stack?** Try tweaking the relevant `guides/` file and see how your agent responds -- that's the fastest feedback loop

When submitting changes:

1. Use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages
2. Keep in mind the audience is AI agents, not humans -- be precise and structured
3. **Run the comprehension tests** before opening a PR (see below)
4. Get a review from the platform team

### Comprehension Tests (required)

The `tests/` directory contains automated tests that verify AI agents correctly understand the documentation. These tests send real prompts to Claude, let it read the docs, and validate that the answers are correct.

**You must run these tests before submitting changes to any guide.** A documentation change that humans can read but AI agents misinterpret is a regression.

```bash
# Prerequisite: Go 1.25+ and claude CLI installed

# Dry run -- validate scenario syntax, no API calls
go run ./tests --dry-run

# Full suite -- ~$0.18, ~90 seconds
go run ./tests --verbose

# Run a single scenario for faster iteration
go run ./tests --scenario "05-*" --verbose
```

The tests cover:

| Scenario | What it verifies |
|----------|-----------------|
| 01-kotlin-api | Identity chain: metadata.id → GCP projects, Helm shortname, Terraform app_id |
| 02-go-service | Go-specific: health paths, distroless image, metrics path |
| 03-data-project | Data project naming: `ent-data-{id}-{int\|ext}-{env}` |
| 04-firebase-app | Firebase uses standard `ent-{id}-{env}`, not a special prefix |
| 05-derive-from-manifest | Distinguishes metadata.id from metadata.name (the #1 confusion) |
| 06-critical-rules | Refuses to create GCP projects via Terraform |

If you change a guide and a test starts failing, either fix the guide or update the test scenario. See [`tests/README.md`](tests/README.md) for how to add new scenarios.

For questions, ideas, or just to say hi, find us in `#talk-utviklerplattform` on Slack.

## License

[EUPL v1.2](https://eupl.eu/1.2/en/) - Entur AS
