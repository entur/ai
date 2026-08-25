---
name: setup-agent-config
description: >
  Set up AI agent configuration for an existing Entur repository. Analyzes the
  project, generates an AGENTS.md that references Entur-wide standards, configures
  agent tool permissions and an optional formatting hook, recommends Entur skills,
  and updates .gitignore. Use this skill when the user says "set up Claude Code",
  "set up agent config", "create AGENTS.md", "auto-setup", "onboard this repo for
  AI agents", or wants agent instructions added to an existing repository.
---

# Entur Agent Config Setup

Configure an existing Entur repository so AI agents (Claude Code, Codex) work with it safely and follow Entur platform standards. Generate the minimum configuration that points agents at the Entur-wide standards -- do **not** copy standards into the repository.

## Step 1: Analyze the Repository

Detect the following from files. Do not ask the user for information you can read from the repository.

| What | How to detect |
|------|---------------|
| Language and build | `build.gradle.kts` or `build.gradle`: Scala plugin or `src/main/scala` = Scala (Gradle), otherwise Kotlin/Java (Gradle); `pom.xml` with `scala-maven-plugin` or `src/main/scala` = Scala (Maven); `go.mod` = Go; `pyproject.toml` or `requirements.txt` = Python; `package.json` = TypeScript/Node (yarn when `yarn.lock` exists, npm otherwise) |
| App identity | `.entur/*.yaml` manifests: read `metadata.id` (App ID), `metadata.name` (Kubernetes namespace), `metadata.owner` (team) |
| Platform surface | `helm/` (common chart), `terraform/`, `.github/workflows/`, `Dockerfile` |
| Formatter | Spotless or ktlint in Gradle config, including Spotless with scalafmt for Scala; Spotless Maven plugin in `pom.xml`; `gofmt` (always present for Go); `ruff` in `pyproject.toml`; `prettier` or `biome` in `package.json` devDependencies |
| Existing agent config | `AGENTS.md`, `CLAUDE.md`, `.claude/settings.json` |

Build files are usually at the repository root, but check one directory level down too (for example `website/package.json`) -- some repositories keep the application in a subdirectory.

Never overwrite existing agent configuration. When `AGENTS.md`, `CLAUDE.md`, or `.claude/settings.json` already exist, add missing sections or entries and preserve everything else.

## Step 2: Generate AGENTS.md

`AGENTS.md` is the single agent instruction file for the repository. It references the Entur-wide standards and adds only project-specific facts an agent cannot derive from the code.

Do **not** copy rules from the entur/ai repository into `AGENTS.md`, and do **not** generate per-language rule files (for example `.claude/rules/kotlin.md`) -- standards live in one place at github.com/entur/ai, and copies drift.

```markdown
# {Display Name}

{Language} application that {one-line description}.

## Entur Standards

Read and follow the Entur platform standards at:
https://github.com/entur/ai/blob/main/AGENTS.md

## Project-Specific

- App ID: {appId}
- Kubernetes namespace: {metadata.name}
- GCP projects: ent-{appId}-dev, ent-{appId}-tst, ent-{appId}-prd

## Commands

| Task | Command |
|------|---------|
| Build | {build command} |
| Test | {test command} |
| Lint | {lint command} |
```

When the existing `AGENTS.md` already references the entur/ai standards in any wording (for example "See https://github.com/entur/ai for Entur-wide standards"), do **not** add a second reference section -- add only the sections that are genuinely missing, such as the identity facts.

Fill the identity facts from the `.entur/` manifest found in Step 1. List one GCP project per environment in the manifest's `spec.environments` -- do not assume all three; a prd-only service gets only `ent-{appId}-prd`. Keep the Kubernetes namespace line only when the service deploys to GKE (`helm/` directory present); for Cloud Run services (`cloudrun.yaml`) replace it with `App name: {metadata.name}`. When the repository has no `.entur/` manifest, omit the identity lines and tell the user to run the **entur-project-bootstrap** skill if the service still needs GCP projects.

Fill Commands from the detected build system:

| Task | Kotlin/Java/Scala (Gradle) | Scala (Maven) | Go | Python | TypeScript/Node |
|------|----------------------------|---------------|-----|--------|-----------------|
| Build | `./gradlew build` | `mvn clean install` | `go build ./...` | -- | `yarn build` |
| Test | `./gradlew test` | `mvn test` | `go test ./...` | `pytest` | `yarn test` |
| Lint | `./gradlew check` | `mvn spotless:check` | `go vet ./...` | `ruff check .` | `yarn lint` |

Treat these as defaults. Prefer a repository-specific aggregate task when its README or CI workflows document one. For Maven, use `./mvnw` instead of `mvn` when the Maven wrapper exists, and include the Spotless command only when the repository configures the plugin.

For TypeScript/Node, use only the scripts that exist in the `scripts` block of `package.json` (`npm run <script>` when the repository uses npm), and note the directory to run them from when the application lives in a subdirectory.

Add a `## Gotchas` section only when the analysis found non-obvious constraints (for example a required local emulator or a generated-code step). Do not pad it.

When tooling in the team only reads `CLAUDE.md`, create it as a symlink to `AGENTS.md` (`ln -s AGENTS.md CLAUDE.md`). Never create `CLAUDE.md` as a separate file with duplicated content. When a `CLAUDE.md` with its own content already exists, merge that content into `AGENTS.md` first and ask the user before replacing the file with a symlink.

## Step 3: Configure Agent Permissions

Write tool permissions to `.claude/settings.json`. Merge with existing content -- never replace the file. Allow the build, test, and lint commands detected in Step 1; deny mutations of cloud infrastructure and reads of secret material.

Example for a Kotlin/Gradle service with Helm and Terraform:

```json
{
  "permissions": {
    "allow": [
      "Bash(./gradlew build:*)",
      "Bash(./gradlew test:*)",
      "Bash(./gradlew check:*)",
      "Bash(helm lint:*)",
      "Bash(helm template:*)",
      "Bash(terraform fmt:*)",
      "Bash(terraform validate:*)"
    ],
    "deny": [
      "Bash(terraform apply:*)",
      "Bash(kubectl apply:*)",
      "Bash(kubectl delete:*)",
      "Bash(gcloud projects create:*)",
      "Read(./.env)",
      "Read(./.env.*)"
    ]
  }
}
```

Use the Gradle entries unchanged for Kotlin, Java, or Scala Gradle repositories. Replace them with the matching commands from the table in Step 2 for Scala with Maven (`mvn clean install`, `mvn test`, plus `mvn spotless:check` when configured), Go (`go build`, `go test`, `go vet`), Python (`pytest`, `ruff check`, `ruff format`), or TypeScript/Node (`yarn build`, `yarn test`, plus `yarn lint` when the script exists). Use the Maven wrapper form when present. Omit the Helm entries when the repository has no `helm/` directory. For Cloud Run services (`cloudrun.yaml`), also deny `Bash(gcloud run deploy:*)` -- rollout runs through CD, not a local session. Keep the deny list in every variant: agents must never apply Terraform, mutate Kubernetes resources, deploy Cloud Run revisions, or create GCP projects from a local session -- those run through CI/CD and self-service manifests.

## Step 4: Add a Formatting Hook (Conditional)

Add a `PostToolUse` hook only when Step 1 detected a formatter. Skip this step otherwise -- a hook that calls a missing tool fails on every edit.

Example for a Go repository, merged into `.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path // empty' | grep '\\.go$' | xargs -r gofmt -w || true"
          }
        ]
      }
    ]
  }
}
```

For Python with ruff, replace the command with `jq -r '.tool_input.file_path // empty' | grep '\.py$' | xargs -r ruff format || true`. For TypeScript/Node with prettier, replace the `grep`/format part with `grep -E '\.(ts|tsx|js|jsx)$' | xargs -r npx prettier --write || true`. For Gradle and Maven projects, do not add a hook -- running the build tool on every edit is too slow; formatting runs in the lint command instead.

Keep hooks non-blocking (`|| true`): a formatting failure must not abort the agent's edit.

## Step 5: Recommend Entur Skills

Tell the user to install the Entur plugin marketplace instead of generating custom commit, review, or test skills -- agents ship with those capabilities, and Entur conventions come from the shared skills.

Do **not** generate `.claude/agents/` files (for example a security-reviewer or refactor-helper agent) -- built-in review and security-review commands and the entur/ai skills cover those, and generated agent files restate Entur standards that live in one place. Hand-author a custom subagent only when a task needs a restricted toolset or a separate context window.

```shell
claude plugin marketplace add entur/ai
```

Codex CLI: `codex plugin marketplace add entur/ai`.

Recommend by project state:

| Situation | Skill |
|-----------|-------|
| Any Entur repository | `guides` (routes Entur conventions on demand) |
| Missing or outdated CI/CD workflows | `cicd-workflows` |
| New service without infrastructure | `bootstrap` |

## Step 6: MCP Servers

Recommend **Entur Kompass** (`https://ki.entur.io/mcp`) -- Entur's approved MCP server on the org-wide MCP allowlist. It gives agents Entur's docs, source code, APIs, GCP runtime state, and GitHub context, authenticated with the user's own Entur Google account.

```shell
claude mcp add --scope user --transport http entur-kompass https://ki.entur.io/mcp
```

Codex CLI: `codex mcp add entur-kompass --url https://ki.entur.io/mcp`. Tools and the access model are documented at github.com/entur/kompass.

Do **not** recommend any other MCP server (database, issue-tracker, error-tracking, or chat integrations). Entur admins enforce an org-wide allowlist of approved MCP servers; additions go through the MCP registry process in the entur/kompass repository (`docs/mcp-registry.md`) and must follow `it-systems-policy.md` in the entur/ai repository.

## Step 7: Update .gitignore

Append the following entries when missing. Do not remove or reorder existing entries.

```gitignore
CLAUDE.local.md
.claude/settings.local.json
```

## Step 8: Print Summary

After all steps, print:

1. Files created and files modified (with paths)
2. Permissions allowed and denied
3. Skills recommended and the marketplace install command
4. Next steps: review the diff, commit with a Conventional Commit (`chore: add AI agent configuration`), open a PR

## Critical Rules

- **Never** overwrite existing `AGENTS.md`, `CLAUDE.md`, or `.claude/settings.json` -- merge and extend
- **Never** copy Entur standards into the repository -- `AGENTS.md` references github.com/entur/ai
- **Never** allow agents to apply Terraform, mutate Kubernetes, or create GCP projects locally
- **Never** recommend software or MCP servers outside `it-systems-policy.md`
- `CLAUDE.md` is only ever a symlink to `AGENTS.md`, never a second instruction file
