# Set Up Freyja

Turn on [Freyja](https://github.com/entur/freyja), a GitHub App that reviews Dependabot PRs, through the `freyja-mode` custom property, which a repository admin sets in the repository's settings. Freyja ignores every PR that is not from Dependabot.

## Goal

Freyja reviews every Dependabot PR in the repository and, in `auto-merge` mode, merges the low-risk ones.

## How Freyja Works

Freyja assesses a Dependabot PR only when `freyja-mode` is set to `advise` or `auto-merge` and CI on the PR is green. For each such PR, Freyja:

1. Checks the PR against hardcoded rules.
2. Asks TypeSafe's Jev model a few questions, if the rules pass.
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
- Before you choose `auto-merge`, the repository passes the [auto-merge checklist](https://github.com/entur/freyja#before-you-turn-on-auto-merge) and has the branch protection below.

## Branch Protection for `auto-merge`

In `auto-merge`, Freyja approves and merges as the `entur-freyja` app, once every check that has reported on the PR is green. Freyja does not bring its own safeguards: your repository's rulesets decide what has to be true before a merge. Set up two rulesets on the default branch. Prefer rulesets to classic branch protection.

**Ruleset 1: checks and branch rules.** Freyja must never bypass this one.

| Setting | Value |
|---------|-------|
| Target | Default branch |
| Bypass list | Your team if you need it, **never `entur-freyja`** |
| Restrict deletions | On |
| Block force pushes | On |
| Require status checks to pass | On, with every check that matters, including checks that report late, such as SonarCloud or external CI statuses |

**Ruleset 2: pull request rule.** Freyja may bypass this one, for pull requests only.

| Setting | Value |
|---------|-------|
| Target | Default branch |
| Bypass list | `entur-freyja`, with **For pull requests only** |
| Require a pull request before merging | On, with the approvals and code owner review you want for people |
| Dismiss stale pull request approvals when new commits are pushed | On |
| Require approval of the most recent reviewable push | On |

Why this matters:

- **Required checks are the only thing that waits for a slow check.** Without them, a fast build can be merged before a slower check has started.
- **Required checks must run on Dependabot PRs.** A required check that is skipped by a path filter or a `dependabot[bot]` condition never reports, so GitHub blocks the merge forever.
- **The pull request rule needs the bypass.** A GitHub App cannot be a code owner. Without the bypass, Freyja approves but cannot merge, and the PR waits for a person. If your pull request rule requires no approval and no code owner review, one ruleset is enough.

To check a repository against these recommendations, run `make audit REPO=entur/<repo>` from the Freyja repository. See [Setting up a repository for Freyja](https://github.com/entur/freyja/blob/main/docs/repo-setup.md) for the full guide.

## Steps

Set this manually in the GitHub UI. An agent cannot do it.

1. Open the repository you want to set up Freyja in.
2. As a repository admin, open **Settings** → **Custom properties** and set `freyja-mode` to `advise`.
3. Review Freyja's verdicts on the next Dependabot PRs.
4. When the verdicts look right, set `freyja-mode` to `auto-merge` the same way.

Freyja now reviews Dependabot PRs in this repository only. To turn it off, set `freyja-mode` back to `off`.

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
