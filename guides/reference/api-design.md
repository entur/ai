# API Design Standards

- **Target audience**: developers and AI agents designing or changing REST APIs.
- **Intent**: APIs follow Entur API guidelines, publishing valid OpenAPI contracts.
- **Scope**: REST API design and OpenAPI conventions. CI linting is covered in [gha-workflows.md](../platform/gha-workflows.md) and [gha-actions.md](../platform/gha-actions.md).

## API guidelines

<https://github.com/entur/api-guidelines>

Entur's authoritative rules for designing RESTful APIs that are described with OpenAPI. Some rules are linted automatically, while others require manual review. Guidelines in markdown format available at <https://raw.githubusercontent.com/entur/api-guidelines/refs/heads/main/guidelines.md>.

## `entur-springdoc-starter`

Entur Spring Boot OpenAPI generation. Add `org.entur.openapi:entur-springdoc-starter` (see the [Common Dependencies table](java.md#dependencies-1) in java.md) to auto-generate an OpenAPI document from Spring controller annotations, instead of hand-maintaining a spec file for a code-first API. For contract-first APIs (spec authored under `specs/`, code generated from it), skip this starter and use the OpenAPI Generator Gradle plugin instead.

## `x-entur-permissions-extension-automatic`

Auto-documenting permissions. With `entur-springdoc-starter` on the classpath, `@PreAuthorize("hasPermission(...)")` business-capability checks are reflected automatically as an `x-entur-permissions` OpenAPI extension on each operation -- no manual annotation is needed. See [permission-store.md](../platform/permission-store.md#minimal-setup-checklist) step 8 for the setup step.

## Linting, validating, and publishing OpenAPI specs in CI/CD

Use `entur/gha-api` for spec lint, validate, and publish -- see the [Available Workflows](../platform/gha-workflows.md#available-workflows) table for the workflow reference. Its `path` input does not accept globs; use a GitHub Actions matrix strategy when the repository has more than one OpenAPI spec file (one matrix entry per `specs/*.yaml`).

## Rate limiting and resilience

Set explicit per-endpoint timeouts and use a circuit breaker (e.g. Resilience4j) around outbound calls to other Entur APIs. Return `429` with a `Retry-After` header when applying your own rate limits; treat a `429` from a dependency as a signal to back off, not an error to retry immediately. See [architecture.md](architecture.md#design-principles) for the broader resilience principles this supports.
