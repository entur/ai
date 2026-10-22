# Set Up Freyja

Turn on [Freyja](https://github.com/entur/freyja), a GitHub App that reviews Dependabot PRs, through a custom property in your repository. Freyja ignores every PR that is not from Dependabot.

## Goal

Freyja reviews every Dependabot PR in the repository and, in `auto-merge` mode, merges the low-risk ones.

## How Freyja Works

Freyja assesses a Dependabot PR only when `freyja-mode` is set to `advise` or `auto-merge` and CI on the PR is green. For each such PR, Freyja:

1. Checks the PR against hardcoded rules.
2. Asks [TypeSafe](https://typesafe.ai)'s Jev model a few questions, if the rules pass.
3. Posts a PR comment explaining the result, and labels the PR.
4. Merges the PR, in `auto-merge` mode.

Freyja has three modes:

| Value | Behaviour |
|-------|-----------|
| `off` (default) | Freyja does nothing |
| `advise` | Posts the `freyja/risk` check and a PR comment explaining the verdict, and labels the PR |
| `auto-merge` | Same as `advise`, and approves and auto-merges low-risk PRs |

Freyja puts at most one label on each PR, mirroring its latest outcome:

| Label | Meaning |
|-------|---------|
| `freyja:approved` | Low risk, needs a human to merge |
| `freyja:merged` | Low risk, merged by Freyja |
| `freyja:uncertain` | Needs a human to review |

## Prerequisites

The following must be true before you start:

- You are a repository admin. Only repository admins can change custom properties.
- The repository has a CI/CD workflow that follows Entur standards. See [setup-cicd-workflows](../../skills/setup-cicd-workflows/SKILL.md).
- Before you choose `auto-merge`, the repository passes the [auto-merge checklist](https://github.com/entur/freyja#before-you-turn-on-auto-merge).

## Steps

Set this manually in the GitHub UI. An agent cannot do it.

1. Open the repository you want to set up Freyja in.
2. Open **Settings** → **Custom properties** and set `freyja-mode` to `advise`.
3. Review Freyja's verdicts on the next Dependabot PRs.
4. When the verdicts look right, set `freyja-mode` to `auto-merge`.

Freyja now reviews Dependabot PRs in this repository only.

## Verify

When CI on the next Dependabot PR is green, the PR has:

- A `freyja/risk` check.
- A Freyja comment explaining the verdict.
- One `freyja:*` label.

## Give Feedback

How you tell Freyja it was wrong depends on the mode:

- **`advise`**: React to Freyja's comment with 👍 or 👎.
- **`auto-merge`**: If a merge was wrong, undo it with the **Revert** button on the PR. Freyja will not auto-merge that bump again. If Freyja left the PR to a person, react with 👍 if it needed a closer look, or 👎 if it was safe to merge.

## See also

- [Freyja README](https://github.com/entur/freyja)
