# Contributing

This is a shared resource for all of Entur, and we'd love your help making it better. Every contribution matters -- whether it's fixing a typo, clarifying a confusing section, adding a new skill, or sharing a pattern that works well for your team.

A few ways to contribute:

- **Found something wrong or unclear?** Open an issue or just submit a PR directly
- **Have a pattern or skill that works great for your team?** Share it -- others will benefit
- **Not sure if something belongs here?** Open an issue and let's figure it out together
- **Want to improve the AI output for your stack?** Tweak the relevant `guides/` file and see how your agent responds -- that's the fastest feedback loop

When submitting changes:

1. Use Conventional Commits (`<type>(<scope>): <description>`) for commit messages
2. Do **not** introduce links to non-Entur external URLs -- see [External Links](guides/reference/documentation.md#external-links) for the allow-list and how to handle the cases where you would have linked out
3. Keep in mind the audience is AI agents, not humans -- follow [Writing AI Documentation](#writing-ai-documentation) below
4. Run the comprehension tests (see [Comprehension Tests](#comprehension-tests))
5. Get a review from another colleague - you know your own solutions best!

For questions, ideas, or just to say hi, find us in `#talk-utviklerplattform` on Slack.

## Writing AI Documentation

The docs in `guides/` are read by AI coding agents that generate platform-compliant code. Humans skim; agents execute. Optimise every guide for the agent.

### Be direct -- "here's how you do it"

- Lead with the action. The first lines of a section should tell the agent what to do, not why it exists.
- Cut theory, history, and motivation unless the agent will make a worse decision without it. If background is needed, push it to the end of the section.
- Imperative voice. Prefer: "Set `metadata.id` to a 3--10 char alphanumeric identifier." over: "The `metadata.id` field is used to..."
- One way to do a thing. If there are multiple options, name the default and only mention alternatives when the agent needs to choose.
- Avoid hedging ("might", "can", "could", "in some cases"). Either it's the rule or it isn't.

### Front-load the answer; rationale comes after

- Open each section with the rule the agent has to follow. The first sentence should be actionable on its own.
- Push reasons, history, and edge cases below the rule so the agent can stop reading once it has what it needs.
- Prefer: "Always create GCP projects via self-service YAML. The Platform Orchestrator owns project lifecycle..." over: "Because the Platform Orchestrator owns project lifecycle, you should always..."

### Use one name per concept

- Pick one term for each concept and use it everywhere. Never silently switch between synonyms ("app ID", "appid", `metadata.id`) -- agents treat the variants as different things.
- Repetition beats pronouns. Prefer: "the common Helm chart" over: "the chart" once you're more than a sentence away from the introduction.
- The single biggest failure mode in this repo has been `metadata.id` vs `metadata.name`. Comprehension test `05-derive-from-manifest` exists because of it. Apply the same discipline to every concept pair that could be confused.

### Name the wrong path, not just the right one

- Agents arrive with priors from other codebases. Explicitly forbid the wrong approach so the prior doesn't win.
- Prefer: "Do **not** use `gcloud projects create` or Terraform `google_project` -- create GCP projects via self-service YAML in `.entur/`." over: "Always use self-service for GCP projects."
- The `AGENTS.md` Critical Rules already follow this pattern. Mirror it in new guides whenever there's a plausible wrong path.

### Use concrete, current code examples

- Every non-trivial rule needs a runnable example. Agents pattern-match -- a worked example beats a paragraph of prose.
- Show the correct form first. Label any anti-example clearly (`# wrong`, `# do not`) and keep it short.
- Pin versions, tags, and references the same way production code does (`?ref=TAG`, `@vN`, specific image tag). Agents will copy whatever they see.
- Keep examples minimal -- only the lines that illustrate the point. Trim imports, boilerplate, and unrelated config.
- Always tag fenced code blocks with a language (`yaml`, `hcl`, `kotlin`, `bash`, ...). Agents use the tag to choose syntax.
- Update examples when the underlying tool, module, or chart version moves. A stale example will be copied verbatim into production.

### Test continuously

- The `tests/` directory contains comprehension tests that send real prompts to an agent, let it read the docs, and check the answer. A doc change that humans understand but agents misread is a regression.
- Run the suite before opening a PR (see [Comprehension Tests](#comprehension-tests)).
- When you add or change a rule that agents are likely to get wrong, add a scenario for it. See `tests/README.md` for how.
- If a test fails after your change, fix the guide or update the scenario -- do not ignore it.

### No external links

- Do **not** link to third-party documentation, vendor docs, framework guides, spec pages, package registries, or any GitHub repo outside `github.com/entur/*`. External URLs are a supply-chain and link-rot risk and require security review on every addition.
- The full allow-list and the pattern for handling cases where you would have linked out are in [guides/reference/documentation.md](guides/reference/documentation.md#external-links).
- When you would have linked: name the tool inline, inline the normative part of the spec the agent needs, and give shell commands directly.

### One source of truth -- link, don't duplicate

- A rule lives in exactly one place. Every other guide that needs it links to that place.
- Duplicated rules drift; the agent will read whichever copy it hit first, and the copies will silently disagree.
- When you find yourself restating a rule, replace the restatement with a link to the canonical guide.

### Worked example: bad vs good

Same topic -- how to create a GCP project -- written two ways.

**Bad:**

> ## GCP Projects
>
> In our platform, GCP projects are an important resource. There are several ways you might create one -- some teams have used Terraform with `google_project`, and others have run `gcloud projects create` directly. We generally recommend going through the self-service orchestrator, since it usually handles lifecycle better. The orchestrator reads YAML manifests; see the cloud vendor docs and the various platform guides for the full picture. The app's ID determines the project name.

What's wrong: no action in the opening line, hedged ("might", "generally", "usually"), doesn't forbid the wrong paths it just named, switches between "app", "application", and "ID", links out, no example.

**Good:**

> ## Create a GCP project
>
> Create GCP projects via self-service YAML in `.entur/`. Do **not** use `gcloud projects create` or Terraform `google_project` -- the Platform Orchestrator owns project lifecycle.
>
> ```yaml
> # .entur/application.yaml
> apiVersion: orchestrator.entur.io/apps/v1
> kind: GoogleCloudApplication
> metadata:
>   id: products       # 3--10 lowercase alphanumeric, unique across Entur
>   name: products-api # becomes the Kubernetes namespace
> spec:
>   environments: [dev, tst, prd]
> ```
>
> `metadata.id` becomes the project suffix: `ent-products-dev`, `ent-products-tst`, `ent-products-prd`. See [self-service.md](guides/platform/self-service.md) for Firebase and data project variants.

What's right: imperative opening, wrong paths named and forbidden, runnable example with pinned `apiVersion`, `metadata.id` and `metadata.name` used consistently, internal link only, rationale ("owns project lifecycle") after the rule.

### Other rules that matter for agent-facing docs

- Use outcome-oriented headings ("Provision a Cloud SQL instance"), not topic dumps ("Cloud SQL").
- State target audience, intent, and scope at the top of each guide.
- Follow [guides/reference/markdown.md](guides/reference/markdown.md) for formatting and run `markdownlint-cli2 "**/*.md"` before committing.

## Comprehension Tests

The `tests/` directory contains automated tests that verify AI agents correctly understand the documentation. The tests send real prompts to Claude, let it read the docs, and validate that the answers are correct.

Run these tests before submitting changes to any existing guides.
A documentation change that humans can read but AI agents misinterpret is a regression.

*NB* For new documentation this is *optional*! 

```bash
# Prerequisite: Go 1.25+ and claude CLI installed

# The tests/ directory is its own Go module, so run the commands from inside it.
cd tests

# Dry run -- validate scenario syntax, no API calls
go run . --dry-run

# Full suite -- ~$0.70, ~3-5 minutes
go run . --verbose

# Run a single scenario for faster iteration
go run . --scenario "05-*" --verbose
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
