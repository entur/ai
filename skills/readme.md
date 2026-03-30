# Skills

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

**Learn more:** [The Complete Guide to Building Skills for Claude](https://resources.anthropic.com/hubfs/The-Complete-Guide-to-Building-Skill-for-Claude.pdf)

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

**Reference:** [The Complete Guide to Building Skills for Claude](https://resources.anthropic.com/hubfs/The-Complete-Guide-to-Building-Skill-for-Claude.pdf) - Section on Best Practices

## 🚀 Installation Guide

Installing skills differs for different agents. This example is for Claude Code, but similar approaches will work for other agents.

1. Download the skills folder.
2. Using Claude Code as an example, select Customize -> Skills -> + Upload a skill.
3. Drag and drop the folder or select the skills you would like to add.

## Available Skills

| Skill | Purpose |
|-------|---------|
| [entur-project-bootstrap](entur-project-bootstrap/) | Bootstrap a new Entur app: self-service manifests, Helm, Terraform, Docker, CI/CD |
| [setup-cicd-workflows](setup-cicd-workflows/) | Generate CI/CD GitHub Actions workflows using Entur reusable workflows |
| [scr-situation-complication-resolution](scr-situation-complication-resolution/) | Structure problems and decisions in SCR format for leadership |

## 🤝 Contributing

We encourage you to share skills that provide value across Entur! Follow the repo guidelines for contribution.
