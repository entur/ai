# Entur AI

> **Audience:** Entur employees using or contributing to this repo. AI coding agents should start at [AGENTS.md](AGENTS.md).

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

**Plugins** wrap skills and other tools and can be installed in one command via the Claude Code or Codex CLI plugin marketplaces. See [Install via plugin marketplace](#install-via-plugin-marketplace) below.

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

That's it. `AGENTS.md` is read automatically by GitHub Copilot and many other agents. Claude Code reads `CLAUDE.md` instead -- see [Agent Compatibility](#agent-compatibility) for details. The agent will fetch the linked URL to get the full platform standards.

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

## Install via plugin marketplace

If you use Claude Code or the Codex CLI, you can install skills as plugins from the `entur/ai` marketplace instead of copying files manually.

### Claude Code

```shell
claude plugin marketplace add entur/ai
claude # then run /plugin to browse
```

### Codex CLI

```shell
codex plugin marketplace add entur/ai
codex # then run /plugins to browse
```

## Install individual skills (any agent)

For agents without a plugin marketplace, the `gh skill` extension for the `gh` CLI can pull skills directly from this repo:

```shell
# Install the extension once via `gh extension install`, then:
gh skill install entur/ai
```

It walks the repo for `SKILL.md` files and lets you pick which to install into your local agent skill folder. See [`skills/README.md`](skills/README.md) for the full install matrix.

## Agent Compatibility

| Agent          | Reads `AGENTS.md`       | Can fetch URLs | Notes                                                                        |
| -------------- | ----------------------- | -------------- | ---------------------------------------------------------------------------- |
| Claude Code    | No (reads `CLAUDE.md`)  | Yes            | Natively reads `CLAUDE.md`; create a symlink or copy for Claude Code support |
| GitHub Copilot | Yes                     | Limited        | Reads `AGENTS.md`; may not fetch URLs in all modes                           |
| opencode       | Yes                     | Unknown        | Reads `AGENTS.md` natively                                                   |

`AGENTS.md` is supported by a large ecosystem of AI coding agents including Codex, Gemini CLI, Jules, Windsurf, Aider, and many more.

Claude Code reads `CLAUDE.md`, not `AGENTS.md`. To support Claude Code alongside other agents, create a symlink: `ln -s AGENTS.md CLAUDE.md`.

For agents that cannot fetch URLs, the most important rules are already inline in your project's `AGENTS.md`. For deeper coverage, you can copy key sections from this repo into your project's instructions.

## Documentation Structure

The `guides/` folder is split into three layers along the internal developer platform value chain:

- **`platform/`** -- what the platform provides (self-service, common Helm chart, reusable workflows, Terraform modules, Permission Store, Kafka starter).
- **`playbooks/`** -- end-to-end tasks (bootstrap, add a database, deploy to prd, etc). These are the agent's recommended entry point for multi-step work.
- **`reference/`** -- language and topic standards (Java, Kotlin, Go, Docker, observability, tracing, profiler, security, code review, docs).

```text
AGENTS.md                              # Top-level agent routing and critical rules
CONVENTIONS.md                         # Cross-language coding conventions
guides/
  platform/
    self-service.md                    # GitOps provisioning, manifests, onboarding
    common-helm.md                     # Entur common Helm chart reference
    gha-workflows.md                   # Reusable GitHub Actions workflows
    gha-actions.md                     # Composite GitHub Actions
    terraform-modules.md               # Terraform modules (init, SQL, Redis, GCS)
    iam-roles.md                       # Approved IAM roles
    permission-store.md                # Permission Store + Permission Client
    entur-kafka-starter.md             # Entur Kafka Spring starter (Aiven)
  playbooks/
    bootstrap-service.md               # New service on the platform
    add-postgres.md                    # Managed PostgreSQL
    add-redis.md                       # Memorystore Redis
    add-kafka.md                       # Aiven Kafka producer/consumer
    set-up-auth.md                     # OIDC + Permission Store authorization
    add-custom-domain.md               # *.entur.{no,io,org} hostname + managed TLS
    deploy-to-prd.md                   # Promote to production
    deprecate-service.md               # Retire an application
    local-dev.md                       # Run the service locally
  reference/
    java.md                            # Java standards (Spring Boot, Gradle)
    kotlin.md                          # Kotlin standards
    go.md                              # Go standards
    docker.md                          # Containerization
    api-design.md                      # REST and gRPC API design
    architecture.md                    # Architecture principles, lifecycle
    logging.md                         # Structured logging
    observability.md                   # Health checks, metrics
    tracing.md                         # OpenTelemetry + Cloud Trace
    profiler.md                        # Cloud Profiler (CPU, heap)
    security.md                        # Secrets, scanning, IAM
    code-review.md                     # Review checklist
    markdown.md                        # Markdown standards and linting
    documentation.md                   # Writing user-facing docs
skills/
  README.md                            # Skill catalogue, usage guide, how to contribute
  entur-project-bootstrap/             # Bootstrap a new app (self-service, Helm, TF, Docker, CI/CD)
  setup-cicd-workflows/                # Generate CI/CD workflows by language
tests/
  README.md                            # Test usage guide and how to add scenarios
  main.go                              # Test runner (Go, stdlib only)
  scenarios/                           # Test scenarios (one .md file per test)
```

AI agents read `AGENTS.md` first, which routes them by goal: a playbook for a multi-step task, a platform doc for a specific capability, or a reference doc for language and topic standards.

## Shared Tooling Referenced

| Category              | Tools                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Terraform modules** | [terraform-google-init](https://github.com/entur/terraform-google-init), [terraform-google-sql-db](https://github.com/entur/terraform-google-sql-db), [terraform-google-memorystore](https://github.com/entur/terraform-google-memorystore), [terraform-google-cloud-storage](https://github.com/entur/terraform-google-cloud-storage)                                                                                                                                              |
| **Helm charts**       | [helm-charts](https://github.com/entur/helm-charts) (common chart)                                                                                                                                                                                                                                                                                                                                                                                                                  |
| **CI/CD workflows**   | [gha-docker](https://github.com/entur/gha-docker), [gha-helm](https://github.com/entur/gha-helm), [gha-terraform](https://github.com/entur/gha-terraform), [gha-security](https://github.com/entur/gha-security), [gha-meta](https://github.com/entur/gha-meta), [gha-firebase](https://github.com/entur/gha-firebase), [gha-docs](https://github.com/entur/gha-docs), [gha-slack](https://github.com/entur/gha-slack), [gha-artifactory](https://github.com/entur/gha-artifactory) |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute, how to write good AI-facing documentation, and how to run the comprehension tests that every doc change must pass.

For questions, ideas, or just to say hi, find us in `#talk-utviklerplattform` on Slack.

## License

[EUPL v1.2](https://eupl.eu/1.2/en/) - Entur AS
