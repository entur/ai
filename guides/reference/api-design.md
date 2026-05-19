# API Design Standards

## API guidelines

<https://raw.githubusercontent.com/entur/api-guidelines/refs/heads/main/guidelines.md>

Entur's authoritative rules for designing RESTful APIs. Covers HTTP method and status code usage, JSON conventions, HTTPS, error formats, pagination, and versioning. Mandates OpenAPI 3.x documentation with an `info.x-entur-metadata` block (required fields: `id`, `owner`, `audience` — one of `open`, `partner`, or `internal`). Specifies JWT authentication for partner/internal endpoints and the `x-entur-permissions` extension for documenting required permissions. Rules are tagged to indicate which are automatically enforced by the linter versus requiring manual review.

## Linting with Spectral (local, IDE, git hooks)

<https://raw.githubusercontent.com/entur/api-guidelines/refs/heads/main/README.md>

Describes how to run the Entur Spectral ruleset outside of CI. Two rulesets are published: `.spectral.yml` (full guideline checks) and `.spectral-required.yml` (minimum rules that must pass to publish to the developer portal). Reference them by tag, e.g. `--ruleset https://raw.githubusercontent.com/entur/api-guidelines/refs/tags/v2/.spectral.yml`. Covers Spectral CLI install, the VS Code extension, and a Husky pre-commit/pre-push hook example. Use this when validating a spec locally before a PR; use the `gha-api` workflows (below) for CI enforcement.

## Linting, validating, and publishing OpenAPI specs in CI/CD

<https://raw.githubusercontent.com/entur/gha-api/refs/heads/main/.github/README.md>

Entur reusable GitHub Actions workflows for OpenAPI specs: `lint.yml` (style/guideline checks), `validate.yml` (schema validity), and `publish.yml` (push to the developer portal via the `api-spec-registry`). Pin to the current major version (e.g. `@v6`) and call with `secrets: inherit`. Default spec path is `specs/openapi.yaml`; the `path` input does not accept globs, so use a matrix strategy for multiple specs. Workflows auto-bundle with Redocly, so no pre-bundling step is needed. Visibility is derived from `info.x-entur-metadata.audience` in the spec — not a workflow input. Lint and validate in CI on PRs; publish from CD after a successful production deploy.
