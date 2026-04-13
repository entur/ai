# Containerization with Docker

All Entur applications deployed to Kubernetes are packaged as Docker images.

## Conventions

- Dockerfile lives at the **repository root**
- Image name follows golden path: **repository name = application name = Docker image name**
- Images pushed to **Google Artifact Registry** via [gha-docker](https://github.com/entur/gha-docker) reusable workflows
- Dockerfiles linted with **Hadolint** via `gha-docker` lint workflow

## Base Images by Language

Prefer **distroless** or **slim-musl** images. Use Alpine only when you need a shell or package manager.

| Language | Recommended | Alternative |
|----------|------------|-------------|
| Java/Kotlin | `bellsoft/liberica-runtime-container:jre-25-cds-slim-musl` | `eclipse-temurin:25-jre-alpine` |
| Go | `gcr.io/distroless/static-debian12:nonroot` | `golang:1.25-alpine` (build only) |
| Node.js | `gcr.io/distroless/nodejs24-debian12` | `node:24-alpine` |
| Python | `gcr.io/distroless/python3-debian12` | `python:3.12-slim` |

Liberica Runtime Container with CDS is preferred for Java/Kotlin: supports Class Data Sharing for faster startup, optimized for containerized JVM workloads. ALWAYS pin base image versions to specific tags.

## Dockerfile Examples

### Java / Kotlin (Preferred: Multi-Stage with Layered JAR and CDS)

Four stages: bundler (OpenAPI spec) → builder (compile) → layers (extract layered JAR) → run (minimal runtime).

```dockerfile
# Stage 1: Bundle OpenAPI specification (contract-first only)
FROM node:25-slim AS bundler
WORKDIR /app
COPY specs specs
RUN npx @redocly/cli bundle specs/products.yaml --output specs/openapi.json

# Stage 2: Build the application
FROM gradle:9.3.1-jdk25-alpine AS builder
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts ./
COPY gradle/libs.versions.toml gradle/libs.versions.toml

# Download dependencies first (cache layer)
RUN --mount=type=secret,id=ARTIFACTORY_AUTH_USER,env=ARTIFACTORY_AUTH_USER  \
    --mount=type=secret,id=ARTIFACTORY_AUTH_TOKEN,env=ARTIFACTORY_AUTH_TOKEN \
    gradle dependencies --no-daemon

COPY src src
COPY --from=bundler /app/specs/openapi.json specs/openapi.json

RUN --mount=type=secret,id=ARTIFACTORY_AUTH_USER,env=ARTIFACTORY_AUTH_USER  \
    --mount=type=secret,id=ARTIFACTORY_AUTH_TOKEN,env=ARTIFACTORY_AUTH_TOKEN \
    gradle bootJar -x clean -x bundleOpenApiSpecification --no-daemon

# Stage 3: Extract layered JAR
FROM bellsoft/liberica-runtime-container:jre-25-cds-slim-musl AS layers
WORKDIR /app
COPY --from=builder /app/build/libs/my-app.jar my-app.jar
RUN java -Djarmode=tools -jar my-app.jar extract --layers --launcher

# Stage 4: Final runtime image
FROM bellsoft/liberica-runtime-container:jre-25-cds-slim-musl AS run
LABEL maintainer="Team Name <team@entur.org>"
EXPOSE 8086
WORKDIR /app

# Copy layers in order of change frequency (least to most)
COPY --from=layers /app/my-app/dependencies/ ./
COPY --from=layers /app/my-app/internal-dependencies/ ./
COPY --from=layers /app/my-app/spring-boot-loader/ ./
COPY --from=layers /app/my-app/application/ .

ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75.0", "org.springframework.boot.loader.launch.JarLauncher"]
```

Key practices:

- **Dependency caching**: Copy build files first, download deps, then copy source -- source changes don't invalidate dependency cache
- **Build secrets**: Use `--mount=type=secret` instead of `ARG`/`ENV` (secrets don't persist in layers)
- **Layered JAR**: Only changed layers are rebuilt when pushing new images
- **CDS**: Liberica CDS image pre-computes class metadata for faster startup
- **`-XX:MaxRAMPercentage=75.0`**: 75% of container memory for JVM, leaving room for OS and native memory

### Java / Kotlin (Simple)

```dockerfile
FROM eclipse-temurin:25-jre-alpine
WORKDIR /app
COPY build/libs/*.jar app.jar

# Non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### Go

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/my-service

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

### Python

```dockerfile
FROM python:3.12-slim AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /install /usr/local
COPY . .

# Non-root user
RUN groupadd -r appgroup && useradd -r -g appgroup appuser
USER appuser

EXPOSE 8080
ENTRYPOINT ["python", "-m", "my_service"]
```

## Best Practices

- **Multi-stage builds**: Separate build from runtime to exclude build tools and source from final image
  > **Note:** Multi-stage builds in GitHub Actions do not support GitHub caching for the build step or GitHub secrets injection. Split into separate workflow steps if needed.
- **Run as non-root**: All containers must run as non-root. Java/Kotlin: `addgroup`/`adduser` + `USER`. Go: use `nonroot` distroless variant (UID 65532). Python: `groupadd`/`useradd` + `USER`. The common Helm chart enforces `runAsNonRoot: true`.
- **Minimize image size**: Use Alpine/slim base images, remove caches (`--no-cache`, `--no-cache-dir`), copy only runtime artifacts
- **No secrets in images**: Use Google Secret Manager + ExternalSecrets for runtime secrets, Helm-injected env vars for non-sensitive config
- **Pin dependencies**: Pin base image tags, build tool versions, and use lock files (`go.sum`, `gradle.lockfile`, `requirements.txt`)
- **Port**: All Entur apps default to port `8080`. Ensure `EXPOSE` and application binding match.
- **Health endpoints**: Expose `GET /actuator/health/liveness` and `GET /actuator/health/readiness` (or equivalent). See [Helm guide](helm.md) for probe config.

## CI/CD Integration

Docker images are built, scanned, and pushed using [gha-docker](https://github.com/entur/gha-docker) reusable workflows.

```yaml
docker-lint:
  uses: entur/gha-docker/.github/workflows/lint.yml@v1

docker-build:
  uses: entur/gha-docker/.github/workflows/build.yml@v1

docker-scan:
  needs: [docker-build]
  uses: entur/gha-security/.github/workflows/docker-scan.yml@v2
  secrets: inherit
  with:
    image_artifact: ${{ needs.docker-build.outputs.image_artifact }}

docker-push:
  needs: [docker-build, docker-scan]
  uses: entur/gha-docker/.github/workflows/push.yml@v1
```

The `docker-push` output `image_and_tag` is used by downstream Helm deploy jobs. For complete CI/CD pipeline examples, see [CI/CD workflows](cicd/workflows.md).

## Hadolint Configuration

Suppress specific rules in `.hadolint.yaml` at repository root:

```yaml
ignored:
  - DL3018  # Pin versions in apk add
```

Only suppress rules with justification. The `gha-docker` lint workflow picks up this file automatically.
