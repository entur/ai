# General Coding Conventions

Cross-language conventions that apply to all Entur projects. Language-specific additions are in `guides/java.md`, `guides/kotlin.md`, and `guides/go.md`.

## Naming Conventions

### Repository and Application Names

- Use lowercase kebab-case: `my-application-name`
- The repository name **is** the application name, Docker image name, Kubernetes namespace, and Helm release name
- Keep names descriptive but concise (under 30 characters recommended)

### Code Naming

| Element | Java/Kotlin | Go | Python |
|---------|-------------|-----|--------|
| Packages | `no.entur.myapp.feature` | `package feature` | `entur.myapp.feature` |
| Classes/Types | `PascalCase` | `PascalCase` | `PascalCase` |
| Methods/Functions | `camelCase` | `PascalCase` (exported), `camelCase` (unexported) | `snake_case` |
| Variables | `camelCase` | `camelCase` | `snake_case` |
| Constants | `SCREAMING_SNAKE_CASE` | `PascalCase` (exported), `camelCase` (unexported) | `SCREAMING_SNAKE_CASE` |
| Database tables | `snake_case` | `snake_case` | `snake_case` |
| REST endpoints | `/kebab-case` | `/kebab-case` | `/kebab-case` |
| Environment variables | `SCREAMING_SNAKE_CASE` | `SCREAMING_SNAKE_CASE` | `SCREAMING_SNAKE_CASE` |

### Configuration Keys

- Spring Boot: `kebab-case` in `application.yml` (e.g. `server.servlet.context-path`)
- Environment variables: `SCREAMING_SNAKE_CASE` (e.g. `DATABASE_URL`)

## Project Structure

### Standard Repository Layout

```text
my-application/
  .github/
    workflows/
      ci.yml                    # Reusable CI build (workflow_call)
      ci-pr.yml                 # PR verification + CI build
      ci-feature.yml            # Feature branch CI (if no open PR)
      deploy.yml                # CD pipeline (dev -> tst -> prd)
      codeql.yml                # Security code scanning (MUST be named codeql.yml)
      lint-api.yml              # API spec linting (optional, for contract-first)
      lint-helm.yml             # Helm chart linting per environment
    dependabot.yml              # Automated dependency updates
    pull_request_template.md    # PR description template
  .entur/
    security/
      codescan.yml              # Code scan allowlist (optional)
      dockerscan.yml            # Docker scan allowlist (optional)
    github-<repo-name>.yaml     # Self-service GitHub manifest
  doc/
    adr/                        # Architecture Decision Records (optional)
      0001-my-decision.adoc
  docs/                         # Documentation (published via gha-docs)
  helm/
    <repo-name>/
      Chart.yaml                # Helm chart, depends on entur/common
      Chart.lock                # Locked dependency versions
      values.yaml               # Default values
      env/
        values-kub-ent-dev.yaml        # Dev environment overrides
        values-kub-ent-tst.yaml        # Test environment overrides
        values-kub-ent-prd.yaml        # Production environment overrides
      tests/                    # Helm unit tests (optional)
  specs/                        # OpenAPI specs (for contract-first APIs)
    products.yaml               # Main OpenAPI entry point
    schemas/                    # Reusable schemas
    parameters/                 # Reusable parameters
    paths/                      # Path definitions
  terraform/
    main.tf                     # Terraform configuration
    variables.tf                # Variables
    outputs.tf                  # Outputs
    env/
      dev.tfvars                # Dev environment variables
      tst.tfvars                # Test environment variables
      prd.tfvars                # Production environment variables
  src/
    main/
      kotlin/                   # Application source code (Kotlin)
      resources/
        application.yml         # Default Spring Boot config
        application-local.yml   # Local development overrides
        db/migration/           # Flyway migrations
    test/
      kotlin/                   # Test source code
      resources/
        application.yml         # Test configuration
        test-data/              # SQL scripts for integration tests
  gradle/
    libs.versions.toml          # Version catalog
    wrapper/                    # Gradle wrapper
  Dockerfile                    # At repository root
  build.gradle.kts              # Gradle build (Java/Kotlin)
  settings.gradle.kts           # Gradle settings
  compose.yaml                  # Docker Compose for local development
  .mise.toml                    # Tool version management (mise)
  README.md                     # Project documentation
  CONTRIBUTING.md               # Developer guide and conventions
  AGENTS.md                     # AI agent instructions (references entur/ai)
```

## Ownership and Responsibility

- **"You build it, you run it"** -- each team is responsible for operating the applications they deploy
- Applications must be real microservices with clear, standalone functionality -- not distributed monoliths
- All versions must support rollback to the previous version
- Maintain backwards compatibility so consumers can update at their own pace
- Applications must start even when dependencies are unavailable -- never crash because a dependency is missing
- All APIs must be documented using OpenAPI (REST) or protobuf definitions (gRPC)
- Application setup and how-to-run instructions must be in the repository's `README.md`
- Non-compliant applications may receive remarks that must be resolved to continue running on the platform

## Error Handling

### General Principles

- Fail fast: validate inputs at the boundary, reject invalid data early
- Use typed exceptions/errors -- avoid generic `Exception` or `error` strings
- Log errors with context: include request IDs, correlation IDs, and relevant parameters
- Never swallow exceptions silently -- always log or propagate
- Distinguish between client errors (4xx) and server errors (5xx) in APIs

### Java/Kotlin

```java
// Good: specific exception with context
throw new ResourceNotFoundException("Route not found: " + routeId);

// Bad: generic exception
throw new RuntimeException("not found");
```

### Go

```go
// Good: wrapped error with context
return fmt.Errorf("fetching route %s: %w", routeID, err)

// Bad: raw error string
return errors.New("failed")
```

### Python

```python
# Good: specific exception with context
raise RouteNotFoundError(f"Route not found: {route_id}")

# Bad: generic exception
raise Exception("not found")
```

## Testing

### Test Pyramid

- **Unit tests**: test individual functions/methods in isolation. High coverage target (80%+).
- **Integration tests**: test interactions between components (database, external services). Use testcontainers or embedded services.
- **Contract tests**: validate API contracts between services.
- **End-to-end tests**: minimal set for critical user journeys only.

### Test Naming

- Java/Kotlin: `ClassName_methodName_expectedBehavior` or descriptive method names with `@DisplayName`
- Kotlin preferred: backtick method names for readable sentences (e.g., `` fun `I should be able to create a new version`() ``)
- Go: `TestFunctionName_scenario` in `_test.go` files
- Python: `test_function_name_scenario` in `test_*.py` files

### Test Best Practices

- Tests must be deterministic -- no flaky tests, no dependencies on external services in unit tests
- Use test fixtures and factories instead of constructing test data inline
- Use the **builder pattern** for test data construction (e.g., `VersionBuilder().withStatus(DRAFT).build()`)
- Mock external dependencies at the boundary (HTTP clients, database connections)
- Each test should test one behavior
- Use Arrange-Act-Assert (AAA) pattern
- Integration tests should use testcontainers for databases and message brokers
- Never commit tests that are `@Disabled` or `@Ignore` without a linked issue
- Use `@Sql` annotations to load test data from SQL scripts before integration tests
- Use `cleanup.sql` scripts to ensure clean state between tests
- Upload test results in CI using `dorny/test-reporter` for visibility in GitHub

### Test Libraries (Kotlin)

| Library | Purpose |
|---------|---------|
| JUnit 5 | Test framework |
| Kotest | Assertions (`shouldBe`, `shouldThrow`, `shouldHaveSize`) |
| SpringMockK | Mocking for Spring/Kotlin (`@MockkBean`) |
| TestContainers | Dockerized PostgreSQL for integration tests |
| Spring Boot Test | `@SpringBootTest`, `@WebMvcTest` |
| Entur Auth Test | `TenantJsonWebToken` for OIDC test tokens |

### Test Structure

```text
src/test/
  kotlin/org/entur/myapp/
    config/               # Test configurations (TestContainersConfig)
    mockdata/             # Enum-based mock data for quick reference
    testdata/             # Test data builders (VersionBuilder)
    utils/                # Test base classes (BaseServiceTest, BaseControllerTest)
    version/              # Feature-specific tests
  resources/
    application.yml       # Test Spring configuration
    test-init.sql         # Database initialization script
    test-data/
      cleanup.sql         # Truncate all tables
      version/            # Feature-specific test data SQL
```

## Dependency Management

### Version Pinning

- **Gradle**: use version catalogs (`libs.versions.toml`) for centralized dependency management
- **Go**: use `go.sum` for checksums, keep `go.mod` tidy
- **Python**: use `requirements.txt` with pinned versions or `poetry.lock`
- **Terraform modules**: always use `?ref=TAG` (e.g. `?ref=v1.2.3`)
- **GitHub Actions**: pin to major version tags (e.g. `@v2`)
- **Docker base images**: use specific tags, not `latest`

### Dependency Updates

- Use **Dependabot** for automated dependency updates (preferred over Renovate). Configure in `.github/dependabot.yml`.
- Review security advisories in dependency update PRs
- Vulnerabilities must be triaged and fixed within **30 days** of discovery
- Keep frameworks and libraries reasonably up to date

## Git and Version Control

### Commit Messages

Follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/):

```text
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`

Examples:

```text
feat(api): add stop-place search endpoint
fix(auth): handle expired tokens gracefully
docs(readme): update deployment instructions
refactor(routing): extract journey planner client
ci: add Docker scan to CI pipeline
```

This enables automated semantic versioning via release-please:

- `feat` -> minor version bump
- `fix` -> patch version bump
- `feat!` or `BREAKING CHANGE` footer -> major version bump

### Branch Strategy

- `main` is the default branch and always deployable
- Feature branches: `feature/<description>` or `<username>/<description>`
- Bugfix branches: `fix/<description>`
- PRs require review approval before merge
- Squash merge to keep history clean

## Local Development

### Tool Version Management (mise)

Use [mise](https://mise.jdx.dev/) (formerly rtx) for consistent tool versions across the team. Define required tools in `.mise.toml`:

```toml
[tools]
java        = 'liberica-25.0.2+12'
terraform   = '1.9.8'
python      = '3.13.10'
kotlin      = '2.3.0'

[settings]
experimental = true

[env]
CLOUDSDK_PYTHON = "python3"
_.source = 'mise.env.sh'

[hooks]
enter = 'mise install'
```

### Docker Compose for Local Development

Use `compose.yaml` at the repository root for running the full application locally:

```yaml
services:
  app:
    build:
      context: .
      secrets:
        - ARTIFACTORY_AUTH_USER
        - ARTIFACTORY_AUTH_TOKEN
    ports:
      - "8086:8086"
    volumes:
      - ${HOME}/.config/gcloud/application_default_credentials.json:/gcp/creds.json:ro
    environment:
      - GOOGLE_APPLICATION_CREDENTIALS=/gcp/creds.json

secrets:
  ARTIFACTORY_AUTH_USER:
    environment: ARTIFACTORY_AUTH_USER
  ARTIFACTORY_AUTH_TOKEN:
    environment: ARTIFACTORY_AUTH_TOKEN
```

### Local Spring Profile

Use `application-local.yml` for local development overrides:

```yaml
entur:
  logging:
    style: humanReadablePlain
  auth:
    authorization:
      permit-all:
        matcher:
          patterns:
            - /v3/api-docs/**
            - /swagger-ui/**

spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/myapp
    username: myapp
    password: myapp

springdoc:
  api-docs:
    enabled: true
  swagger-ui:
    enabled: true
```

## Architecture Decision Records (ADRs)

Document significant architectural decisions in `doc/adr/` using AsciiDoc format. ADRs capture the context, decision, consequences, and alternatives considered for important technical choices.

### ADR Format

```asciidoc
== N. Title of Decision

Date: YYYY-MM-DD

== Status
Accepted | Proposed | Deprecated

== Context
What is the problem or situation that requires a decision?

== Decision
What is the decision and how will it be implemented?

== Consequences
=== Positive
=== Negative

== Alternatives
What other options were considered and why were they rejected?
```

### When to Write an ADR

- Choosing a framework, library, or tool over alternatives
- Changing the architecture pattern (e.g., from ORM to SQL-DSL)
- Adopting a new development workflow (e.g., contract-first API design)
- Any decision that future developers would question or need context for

## PR Templates

Use `.github/pull_request_template.md` to ensure consistent PR descriptions:

```markdown
## Beskrivelse
<Describe the purpose of this change in one or two sentences.>
Fixes <JIRA ticket number>.

## Huskeliste
- [ ] Correct JIRA ticket number in PR title
- [ ] Tests created/updated (unit/integration/Postman)
- [ ] Documentation updated if needed
- [ ] Consumers informed of breaking changes
- [ ] EXPLAIN run on SQL queries and optimized if needed
```

## Configuration Management

### Environment-Specific Configuration

- Use environment variables for environment-specific values
- Spring Boot: `application.yml` for defaults, `application-{profile}.yml` for overrides
- Never commit environment-specific secrets or credentials
- Use Google Secret Manager for secrets, referenced via ExternalSecrets in Helm

### Configuration Hierarchy (Spring Boot)

1. `application.yml` -- defaults
2. `application-{profile}.yml` -- profile-specific overrides
3. Environment variables -- highest precedence, set via Helm values

## Code Quality

### Static Analysis

- Java/Kotlin: use the project's configured linter (typically Ktlint for Kotlin, Checkstyle or SpotBugs for Java)
- Go: `golangci-lint` with project configuration
- Python: `ruff` or `flake8` + `black` for formatting
- All languages: CodeQL security scanning via `gha-security`

### Code Formatting

- Enforce consistent formatting via CI -- code that fails formatting checks must not be merged
- Java: follow project formatter configuration (typically Google Java Style or similar)
- Kotlin: Ktlint with default rules
- Go: `gofmt` (non-negotiable)
- Python: `black` with default configuration

## Documentation

### Code Documentation

- Document public APIs (classes, methods, endpoints)
- Use Javadoc (Java), KDoc (Kotlin), godoc (Go), or docstrings (Python)
- Document "why", not "what" -- the code shows what, comments explain why
- Keep documentation close to the code it describes

### Project Documentation

- `README.md` at repository root: what the project does, how to run it locally, how to deploy
- `AGENTS.md` at repository root: AI agent instructions (reference `entur/ai` plus project-specific overrides)
- `docs/` directory: detailed documentation, published via `gha-docs`
- API documentation: OpenAPI/Swagger for REST APIs, protobuf definitions for gRPC
