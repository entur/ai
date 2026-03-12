# Code Review

## For Authors

### Before Opening a PR

- [ ] Code compiles and all tests pass locally
- [ ] New code has appropriate test coverage
- [ ] PR title follows [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) (validated by CI)
- [ ] PR description explains **what** and **why**
- [ ] Large changes broken into smaller PRs
- [ ] No unrelated changes bundled
- [ ] No secrets or credentials committed

### PR Description Template

```markdown
## What
Brief description of the change.

## Why
Motivation and context. Link to issue if applicable.

## How
Approach taken and key decisions.

## Testing
How was this tested? Any manual verification steps?
```

## For Reviewers

### Correctness

- Does the code do what it claims?
- Edge cases handled (null, empty collections, boundaries)?
- Errors handled properly (not swallowed, logged with context)?
- Race conditions possible in concurrent code?

### Design

- Follows existing architecture patterns?
- Right level of abstraction?
- Clear separation of responsibilities (controller / service / repository)?
- New dependencies justified?

### Security

- No hardcoded secrets or credentials
- Input validated at boundaries
- Error responses don't leak internals
- IAM roles from [approved list](terraform/iam-roles.md) only
- SQL uses parameterized statements (no string concatenation)

### Entur Platform Compliance

- Uses Entur shared Terraform modules (not raw GCP resources)
- Uses Entur common Helm chart
- Uses Entur reusable GitHub Actions workflows
- Follows golden path conventions (naming, structure, configuration)
- Dependencies pinned
- Dependabot configured (`.github/dependabot.yml`)
- SonarCloud checked (if enabled)

### Testing

- New functionality has tests
- Tests verify behavior, not implementation
- Mocks at boundaries only, not for internal classes
- No flaky or non-deterministic tests
- Integration tests use testcontainers (not external services)

### Observability

- Structured logging with context
- Health checks configured
- Prometheus metrics for key operations
- No sensitive data in logs

### Review Etiquette

- Suggest improvements, don't just point out problems
- Distinguish blocking issues from nitpicks
- Use prefixes: `nit:`, `question:`, `blocker:`
- Approve with minor comments when overall approach is sound
- Review within one business day

## CI Checks

Every PR must pass before merge:

| Check | Workflow | Description |
|-------|----------|-------------|
| Lint | Various | Code formatting and style |
| Unit tests | Project-specific | All unit tests pass |
| Code scan | `codeql.yml` | CodeQL security analysis |
| Docker scan | CI pipeline | Grype vulnerability scan |
| Docker lint | CI pipeline | Hadolint Dockerfile lint |
| Helm lint | CI pipeline | Helm chart validation |
| PR title | `verify-pr.yml` | Conventional commit format |
