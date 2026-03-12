# API Design Standards

Guidelines for REST and gRPC APIs at Entur.

## REST API Conventions

### URL Structure

```text
https://<service>.entur.org/api/v{version}/{resource}
```

- Use kebab-case for URL path segments: `/stop-places`, `/journey-plans`
- Use plural nouns for collections: `/routes`, `/stops`
- Use path parameters for identity: `/routes/{id}`
- Use query parameters for filtering/sorting/pagination: `/routes?origin=Oslo&limit=20`
- Version in URL path: `/api/v1/`, `/api/v2/`

### HTTP Methods

| Method | Purpose | Idempotent | Response |
|--------|---------|------------|----------|
| `GET` | Retrieve resource/collection | Yes | `200 OK` |
| `POST` | Create new resource | No | `201 Created` with `Location` header |
| `PUT` | Replace resource entirely | Yes | `200 OK` or `204 No Content` |
| `PATCH` | Partially update resource | No | `200 OK` |
| `DELETE` | Remove resource | Yes | `204 No Content` |

### Status Codes

#### Success

| Code | Use |
|------|-----|
| `200` | Successful GET, PUT, PATCH |
| `201` | Successful POST (include `Location` header) |
| `204` | Successful DELETE or PUT with no response body |

#### Client Errors

| Code | Use |
|------|-----|
| `400` | Malformed request, validation error |
| `401` | Missing or invalid authentication |
| `403` | Authenticated but lacks permission |
| `404` | Resource does not exist |
| `409` | Conflict with current state (duplicate, version mismatch) |
| `422` | Semantically invalid request |
| `429` | Rate limit exceeded |

#### Server Errors

| Code | Use |
|------|-----|
| `500` | Unexpected server failure |
| `502` | Upstream service failure |
| `503` | Service temporarily unavailable |
| `504` | Upstream service timeout |

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Origin stop place is required",
    "details": [
      {
        "field": "origin",
        "message": "must not be blank"
      }
    ]
  }
}
```

- Always return structured error body for 4xx/5xx responses
- Include machine-readable `code` and human-readable `message`
- Never expose stack traces, SQL, or internal details

### Pagination

Cursor-based (preferred for large collections):

```json
GET /api/v1/routes?limit=20&cursor=eyJpZCI6MTAwfQ

{
  "data": [...],
  "pagination": {
    "limit": 20,
    "cursor": "eyJpZCI6MTIwfQ",
    "hasMore": true
  }
}
```

Offset-based (acceptable for simpler use cases): use `page`, `size`, `totalElements`, `totalPages` fields.

### Request and Response Bodies

- Use `camelCase` for JSON field names
- Use ISO 8601 for dates/times: `"2025-01-15T10:30:00Z"`
- Use ISO 4217 for currency codes: `"NOK"`
- Wrap collections in a named field: `{"data": [...]}`
- Include only necessary fields

### Content Negotiation

- Default to `application/json`
- Set `Content-Type: application/json` on all JSON responses
- Accept `Accept` header when supporting multiple formats

## API Documentation

### OpenAPI / Swagger

All REST APIs must have an OpenAPI spec. Two approaches:

#### Contract-First (Preferred for Kotlin)

Define API spec before code. Use **OpenAPI Generator** to generate interfaces and DTOs:

1. Write modular OpenAPI specs in `specs/`
2. Bundle with **Redocly CLI** into a single spec
3. Generate Kotlin Spring interfaces and DTOs at build time
4. Implement generated interfaces in controllers

Ensures contract stays in sync with implementation. See [Contract-First OpenAPI](#contract-first-openapi) below.

#### Code-First (Simpler Projects)

Annotate controllers with `@Operation`, `@ApiResponse` and let springdoc-openapi generate the spec. Serve at `/api-docs` or `/v3/api-docs` (Spring Boot default).

### Entur Springdoc Starter

For code-first APIs, use **`entur-springdoc-starter`** (`org.entur.openapi:entur-springdoc-starter`) alongside springdoc-openapi. Adds Entur-specific OpenAPI extensions required by the developer portal and API gateway.

#### Setup

```kotlin
// build.gradle.kts
dependencies {
    implementation("org.springdoc:springdoc-openapi-starter-webmvc-api:<version>")
    implementation("org.entur.openapi:entur-springdoc-starter:<version>")  // check Artifactory for latest
}
```

Published to Entur's [JFrog Artifactory](https://entur2.jfrog.io). See [java.md](java.md#artifactory-jfrog) for repository configuration.

```yaml
# application.yml
springdoc:
  default-produces-media-type: application/json
  default-consumes-media-type: application/json
```

#### `x-entur-metadata` Extension

Every API must declare `x-entur-metadata` on its `info` object to identify the API in the developer portal:

Create a `@Configuration` class with a `@Bean` method returning `OpenAPI`. Set `Info` with title, version, description, and attach `EnturMetadata` via the `.enturMetadata()` extension function (Kotlin) or `enturMetadata(info, metadata)` static import (Java). Set a unique `.id()` for the API.

Use `.parentId()` for APIs that should appear under a parent in the portal (content shown under the parent API, not standalone).

#### `x-entur-permissions` Extension (Automatic)

Auto-generated from `@PreAuthorize` annotations. No code changes needed. The parser understands:

- `hasPermission('resource', 'access')` -- single permission
- `AND` combinations → `all` list
- `OR` combinations → `any` list
- Nested `AND`/`OR`; non-permission nodes like `hasAnyAuthority` are skipped

Example: `@PreAuthorize("hasPermission('items', 'les') || hasPermission('items-global', 'les')")` produces `"x-entur-permissions": { "value": { "any": ["items:les", "items-global:les"] } }`.

#### `@EnturPermissions` Override

Override auto-generated permissions when `@PreAuthorize` is insufficient. Add `@EnturPermissions` with a `description` and optional `value` containing `EnturPermissionsValue` with `any` or `all` lists.

If only `description` is set (without `value`), `@PreAuthorize` is still used for the permission value.

#### `@SchemaExample` Annotation

Generates OpenAPI examples from actual class instances, ensuring examples compile and stay in sync. Add a `static` method (or `@JvmStatic` in a Kotlin companion object) annotated with `@SchemaExample` that returns an instance of the enclosing class. Serialized using the app's `ObjectMapper`.

#### Custom TypeNameResolver

Control how generic types appear in OpenAPI schema by providing a `@Bean` of type `TypeNameResolver`. Override `nameForGenericType` to customize schema names (e.g., `PageDtoItem` -> `PageItem`).

## gRPC APIs

- Define services/messages in `.proto` files
- Use `PascalCase` for message/service names, `snake_case` for fields
- Version proto packages: `entur.myservice.v1`
- Use gRPC health checking protocol (`grpc.health.v1.Health`)
- Enable in Helm:

```yaml
common:
  grpc: true
  ingress:
    trafficType: http2
```

The common Helm chart auto-configures gRPC probes when `grpc: true`.

## Exposing APIs

### Domain Patterns

| Application Type | Production | Test | Development |
|------------------|-----------|------|-------------|
| Frontend / Web | `.entur.no` | `.staging.entur.no` | `.dev.entur.no` |
| API | `.entur.io` | `.staging.entur.io` | `.dev.entur.io` |

### Apigee API Gateway

External APIs are exposed through **Apigee** at `api.entur.io`. Set `ingress.trafficType: api` in Helm to route through Apigee.

| Environment | URL pattern |
|-------------|-------------|
| Production | `https://api.entur.io/<api-path>/<version>` |
| Test | `https://api.staging.entur.io/<api-path>/<version>` |
| Development | `https://api.dev.entur.io/<api-path>/<version>` |

**Limitation**: 10 MB max per single request. Use streaming for larger payloads.

For proxy configuration, see "API Gateway Apigee X" docs or `#talk-utviklerplattform`.

### Internal Service URLs

For cluster-internal communication:

- Same namespace: `http://<service-name>`
- Cross-namespace: `http://<service-name>.<namespace>.entur.internal`
- Only TCP ports 80 and 443 between services
- Set `ingress.enabled: false` for internal-only services

### gRPC External Limitations

gRPC is supported cluster-internal and frontend-external. **Apigee does not support gRPC proxies** -- contact `#talk-utviklerplattform` for external gRPC exposure.

## API Versioning

- URL path versioning: `/api/v1/`, `/api/v2/`
- Major version only for breaking changes
- Support previous version for minimum 6 months
- Breaking changes: removing fields, changing types, changing URL structure, changing error codes

## Contract-First OpenAPI

Preferred approach for Kotlin Spring Boot APIs using OpenAPI Generator.

### Directory Structure

```text
specs/
  products.yaml                    # Main entry point
  openapi.json                     # Bundled output (gitignored)
  schemas/
    _index.yaml                   # Schema index
    Version.yaml                  # Individual schemas
    enums/
      VersionStatus.yaml          # Enum definitions
  parameters/
    _index.yaml                   # Parameter index
    header/
      et-client-name.yaml
      x-correlation-id.yaml
    path/
      id.yaml
  paths/
    VersionV3.yaml                # Path definitions
```

### Main Spec File

```yaml
openapi: "3.0.3"
info:
  version: "1.0.0"
  title: My API
  contact:
    name: Team Name
    email: team@entur.org
servers:
  - url: "https://api.entur.io/my-api"
    description: "Production environment"
  - url: "https://api.staging.entur.io/my-api"
    description: "Staging environment"
  - url: "https://api.dev.entur.io/my-api"
    description: "Development environment"
tags:
  - name: version
    description: Version management endpoints
paths:
  /v3/versions/{id}:
    $ref: "./paths/VersionV3.yaml#/~1v3~1versions~1{id}"
```

### Schema with Read-Only Fields

Use `readOnly: true` for server-generated fields:

```yaml
type: object
required:
  - id
  - status
  - startDate
properties:
  id:
    pattern: "^([A-Z]{3}):Version:([0-9A-Za-z_\\-]*)$"
    type: string
  status:
    $ref: "./enums/VersionStatus.yaml"
  startDate:
    type: string
    format: date
  endDate:
    type: string
    format: date
allOf:
  - type: object
    properties:
      created:
        type: string
        format: date-time
        readOnly: true
      changed:
        type: string
        format: date-time
        readOnly: true
```

### Bundling and Code Generation (Gradle)

```kotlin
// build.gradle.kts
openApiGenerate {
    validateSpec = true
    inputSpec = "specs/openapi.json"
    outputDir = "$generatedSourcesDir/openapi"
    generatorName = "kotlin-spring"
    apiPackage = "org.entur.myapp.api"
    modelPackage = "org.entur.myapp.dto"
    configOptions = mapOf(
        "interfaceOnly" to "true",         // Generate interfaces, not implementations
        "useBeanValidation" to "true",     // Add Jakarta validation annotations
        "useSpringBoot3" to "true",        // Spring Boot 3 / Jakarta EE
        "exceptionHandler" to "false",     // Use custom exception handler
        "useTags" to "true",               // Group by tags
    )
    typeMappings = mapOf(
        "java.time.OffsetDateTime" to "java.time.ZonedDateTime",
        "kotlin.Float" to "java.math.BigDecimal",
    )
}
```

Bundling with Redocly CLI:

```kotlin
// Gradle task using Docker
register("bundleOpenApiSpecification", Exec::class) {
    commandLine("docker", "run", "--rm",
        "-v", "$projectDirPath/specs:/spec",
        "redocly/cli", "bundle", "products.yaml", "-o", "openapi.json")
}

// Build chain
compileKotlin { dependsOn("openApiGenerate") }
openApiGenerate { dependsOn("bundleOpenApiSpecification") }
```

### API Spec Linting in CI

```yaml
name: lint-api
on:
  pull_request:
    paths:
      - 'specs/**'
jobs:
  api-lint:
    uses: entur/gha-api/.github/workflows/lint.yml@v5
    with:
      spec: specs/*.yaml
```

### Publishing to Developer Portal

```yaml
openapi-publish:
  uses: entur/gha-api/.github/workflows/publish.yml@v5
  with:
    artifact: openapi-spec
    visibility: partner       # partner | public | internal
```

## Rate Limiting and Resilience

- Retry with exponential backoff for outgoing HTTP calls
- Set timeouts on all HTTP clients (connect: 5s, read: 30s)
- Use circuit breakers for external services
- Return `429 Too Many Requests` with `Retry-After` header when rate limiting
