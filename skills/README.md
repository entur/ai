# Skills

> **Audience:** Entur employees installing or contributing skills.
> **AI agents:** Stop. Read [AGENTS.md](../AGENTS.md) instead of this file.

This is a centralized collection of reusable skills designed to enhance our AI agents and streamline development across Entur.

## 📌 Purpose

This folder serves as a hub for:

- **Collecting reusable skills** that can be leveraged across multiple Entur projects
- **Sharing best practices** for building and integrating AI skills

## 🤖 What Are Skills?

Agent Skills are folders of instructions, scripts, and resources that AI agents can discover and use to perform at specific tasks. Write once, use everywhere.

We document skills in a SKILL.md file following a structure:

```text
-
name: your-skill
description: [...]
---
# Your Skill Name
# Instructions
### Step 1: [First Major Step]
Clear explanation of what happens.
```

## 📁 Repository Structure

Skills are organized under the skills folder, naming the folder as the skill.
A skill folder should at minimum contain a SKILL.md file, but can also include scripts and assets.

```text
skills/
└── your-skill-name/
    ├── SKILL.md              # Required - main skill file
    ├── scripts/              # Optional - executable code
    │   ├── process_data.py   # Example
    │   └── validate.sh       # Example
    ├── references/           # Optional - documentation
    │   ├── api-guide.md      # Example
    │   └── examples/         # Example
    └── assets/               # Optional - templates, etc.
        └── report-template.md # Example
```

### Naming Conventions

- Use **lowercase with hyphens** for skill folder names (e.g., `weather-fetcher`, `database-query`)
- Use **snake_case** for Python files and functions
- Include meaningful names that describe the skill's purpose

## 🚀 Installation Guide

There are several ways to install skills from this repository, depending on which agent or tooling you use. All skills below live in `skills/` at the repo root and are also exposed as Claude Code / Codex plugins (one plugin per skill).

### Option 1 — Claude Code plugin marketplace

```shell
claude plugin marketplace add entur/ai
claude # then run /plugin to browse
```

### Option 2 — Codex CLI plugin marketplace

```shell
codex plugin marketplace add entur/ai
codex # then run /plugins to browse
```

### Option 3 — `gh skill install` (any agent)

`gh skill` is a third-party `gh` CLI extension. Install the extension once via `gh extension install <owner>/gh-skill` (see `gh extension --help`), then the command below walks the repo for `SKILL.md` files recursively and lets you pick which to install.

```shell
gh skill install entur/ai
```

### Option 4 — Manual upload (Claude Code UI)

1. Clone or download this repo.
2. In Claude Code, open Customize → Skills → + Upload a skill.
3. Drag the skill folder (e.g. `skills/scr-situation-complication-resolution/`) into the upload area.

## Available Skills

| Skill | Purpose |
|-------|---------|
| [entur-project-bootstrap](entur-project-bootstrap/) | Bootstrap a new Entur app: self-service manifests, Helm, Terraform, Docker, CI/CD |
| [setup-agent-config](setup-agent-config/) | Set up AI agent config for an existing repo: AGENTS.md, permissions, hooks, skill recommendations |
| [setup-cicd-workflows](setup-cicd-workflows/) | Generate CI/CD GitHub Actions workflows using Entur reusable workflows |
| [scr-situation-complication-resolution](scr-situation-complication-resolution/) | Structure problems and decisions in SCR format for leadership |
| [entur-linje](https://github.com/entur/design-system/tree/main/skills) | Entur Linje design system: `@entur/*` components, design tokens, brand identity, universell utforming, package upgrades. Lives in `entur/design-system`; install via the `entur-linje` plugin |
| [entur-uniformen](https://github.com/entur/uniformen/tree/main/skills) | Uniformen: Entur's shared header and footer for B2B and partner-facing portals — `@entur/uniformen` integration and language contracts, and the CSP the rendered chrome needs. Lives in `entur/uniformen`; install via the `entur-uniformen` plugin |

## 🤝 Contributing

We encourage you to share skills that provide value across Entur! Follow the repo guidelines for contribution.
