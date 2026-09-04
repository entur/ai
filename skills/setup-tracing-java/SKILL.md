---
name: setup-tracing-java
description: >
    Wire distributed tracing into an Entur Kotlin/Java Spring Boot service
    using the OpenTelemetry Java Agent, following the golden path to Cloud
    Trace. Use when the user says "add tracing", "set up OpenTelemetry",
    "instrument for Cloud Trace", or "add distributed tracing" for a Kotlin
    or Java service that runs on Spring Boot -- typically a repo with
    `build.gradle.kts` and a `spring-boot-starter-*` dependency. Does not
    apply to non-Spring-Boot Kotlin/Java services (e.g. plain Ktor, Micronaut,
    or CLI apps) -- those need manual OpenTelemetry SDK instrumentation,
    which is out of scope for this skill.
---

# Set Up Tracing -- Kotlin/Java (Golden Path)

Wire distributed tracing into an Entur Kotlin/Java service so every inbound request produces a span in Cloud Trace, correlated with structured logs. This golden path covers **Spring Boot via the OpenTelemetry Java Agent** only -- it does not cover Go, Python, or any other language. Trace spans are exported to a shared host project (`ent-kub-<env>`), not the application's own project -- this is what makes it possible to correlate a trace with the logs from the same request.

**Before starting, confirm the project is actually Spring Boot** (a `spring-boot-starter-*` dependency in `build.gradle.kts`). If it isn't -- e.g. a plain Ktor, Micronaut, or non-web service -- stop and tell the user this skill only covers Spring Boot; manual OpenTelemetry SDK instrumentation is out of scope here.

## Step 1: Instrument the application

### OpenTelemetry Java Agent (default; do not hand-instrument without a specific reason)

Attach the agent via `-javaagent` in the existing Dockerfile: add one temporary stage that downloads the JARs, and merge the flags into the final stage's `ENTRYPOINT` -- do not touch any other existing stages (bundler, builder, layers, etc.) and do not introduce a `CMD`. Entur's Docker golden path ([docker.md](../../guides/reference/docker.md)) always launches the JVM via `ENTRYPOINT`, both in the preferred layered-JAR pattern and the single-jar Alpine alternative, so the agent flags belong in that array, prepended before the existing launch arguments.

```dockerfile
# Stage: download OTel JARs (temporary -- never shipped)
FROM alpine:3.24 AS otel
RUN mkdir /otel && \
    wget -q -O /otel/opentelemetry-javaagent.jar \
      https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v2.29.0/opentelemetry-javaagent.jar && \
    wget -q -O /otel/gcp-auth-extension.jar \
      https://repo1.maven.org/maven2/io/opentelemetry/contrib/opentelemetry-gcp-auth-extension/1.58.0-alpha/opentelemetry-gcp-auth-extension-1.58.0-alpha-shadow.jar
```

Then, in the project's existing final (runtime) stage, add `COPY --from=otel /otel /otel` and prepend the agent flags to the existing `ENTRYPOINT`. For the **preferred layered-JAR pattern** (docker.md's default), that looks like:

```dockerfile
# Final runtime stage -- only the two lines below are new; everything else
# (the COPY --from=layers lines, base image, etc.) is whatever the project already has
COPY --from=otel /otel /otel
ENTRYPOINT ["java", \
    "-javaagent:/otel/opentelemetry-javaagent.jar", \
    "-Dotel.javaagent.extensions=/otel/gcp-auth-extension.jar", \
    "-Dotel.javaagent.logging=application", \
    "-XX:MaxRAMPercentage=75.0", \
    "org.springframework.boot.loader.launch.JarLauncher"]
```

For the **single-jar Alpine alternative** (`ENTRYPOINT ["java", "-jar", "app.jar"]`), insert the same three `-javaagent`/`-D` flags immediately before `-jar` in that array instead.

- Both JARs are required: `opentelemetry-javaagent.jar` does the bytecode instrumentation; `gcp-auth-extension` attaches a valid GCP access token (via Application Default Credentials) to outbound OTLP calls -- without it, `telemetry.googleapis.com` rejects the export as unauthenticated.
- Use the pinned versions above as-is -- `v2.29.0` (agent) / `1.58.0-alpha` (gcp-auth-extension). Do not spend time checking upstream for a newer release. In Step 4's summary, tell the user which versions were used and point to the `otel` build stage in the Dockerfile as where to bump them later.
- The base/runtime image (e.g. `java25-debian13`) is whatever the project's Dockerfile already uses -- do not change it. Only add the `otel` stage and edit the final stage's `ENTRYPOINT`.

## Step 2: Set the sampler explicitly

Sampling must be set explicitly per environment -- do not rely on the default everywhere.

**If `helm/<app>/env/values-kub-ent-<env>.yaml` doesn't exist yet for an environment**, ask the user: *"No Helm values file exists yet for `<env>` -- do you want me to create `helm/<app>/env/values-kub-ent-<env>.yaml` with the sampler env vars?"* Only create it if they say yes.

```yaml
# values-kub-ent-dev.yaml
common:
  container:
    env:
      - name: OTEL_TRACES_SAMPLER
        value: parentbased_always_on # sample everything, easiest for debugging
```

```yaml
# values-kub-ent-prd.yaml / values-kub-ent-tst.yaml
common:
  container:
    env:
      - name: OTEL_TRACES_SAMPLER
        value: parentbased_traceidratio # sample a fixed ratio of requests
      - name: OTEL_TRACES_SAMPLER_ARG
        value: "0.1" # 10% of requests
```

If the file already exists with its own `common.container.env` entries, append these to the existing list instead of overwriting the file -- check for entries with the same `name` first and update their `value` in place rather than duplicating.

Only create the single env file this way -- do not scaffold `helm/<app>/Chart.yaml` or `helm/<app>/values.yaml` ad hoc. If those don't exist either, Helm hasn't been bootstrapped for this service at all, which is out of scope for this skill.

## Step 3: Correlate logs with traces

For Java/Kotlin Spring Boot services, no manual work is needed -- **provided the project is on `entur/cloud-logging` v7.1.0 or later**. Only 7.1.0+ injects the trace/span fields via Micrometer Tracing; earlier versions do not support this correlation. Together with `spring-boot-starter-gcp-web`, that's all that's required for standard SLF4J logging to pick up the trace/span ID on every log line within a traced request.

Check the `entur/cloud-logging` version pinned in `build.gradle.kts` (or the version catalog):

- **v7.1.0 or later already present**: nothing to do.
- **Older version, or not present at all**: log/trace correlation will not work until this is fixed, and it's outside what this skill edits. Tell the user their logs won't correlate with traces yet, and point them to [logging.md](../../guides/reference/logging.md) to add or upgrade `entur/cloud-logging` and `spring-boot-starter-gcp-web`. Note that `cloud-logging` 7.x requires Spring Boot 4.1.x -- if the project is on an older Spring Boot line, upgrading `cloud-logging` alone won't be enough; flag that dependency too.

## Step 4: Tell the user what's left

Steps 1-3 are everything this skill can do by editing the repo. Verifying traces actually arrive requires a live deploy and real traffic, which happens outside this skill -- tell the user to:

1. Commit and merge the changes, and let the normal CD pipeline deploy.
2. Send a few requests to the deployed service to generate spans.
3. Check **Monitoring → Trace → Trace Explorer** in the GCP Console under the shared host project `ent-kub-<env>` (not the application project, traces always land in the host project). Since traces from multiple applications land in the same project, filter Trace Explorer by `service.name` to scope the view to just this service.
4. If no spans show up: trace storage provisions automatically the first time a span is successfully written to the project, and it isn't instant -- give it a few minutes before assuming something is broken. Also double check the sampler env var actually reached the deployed container for that environment (a common miss is setting it in the wrong Helm values file).

Also state in the summary: the OpenTelemetry Java agent (`v2.29.0`) and gcp-auth-extension (`1.58.0-alpha`) versions were pinned as-is, not checked against upstream for something newer, and are set in the Dockerfile's `otel` build stage if the user wants to bump them later.

## Critical Rules

- **Traces land in the shared host project `ent-kub-<env>`**, never the application's own project -- this applies to Kubernetes and Cloud Run alike. No `GCP_PROJECT_ID`/`GOOGLE_CLOUD_PROJECT` env var configuration is needed for tracing to work.
- **IAM roles and the required Google Cloud APIs are provisioned automatically** through the common Helm chart. Never add Terraform to enable `cloudtrace.googleapis.com`/`telemetry.googleapis.com` or to grant a trace-related IAM role for this.
- **Trace storage auto-provisions** on first successful span write -- never tell the user to manually enable it in the console, and never script it.
- **Defaults to the Java Agent.** Only hand-roll manual OpenTelemetry instrumentation if the user gives a specific reason.
- **One JVM launch entrypoint per Dockerfile, and it's `ENTRYPOINT`, not `CMD`** -- Entur's Docker golden path launches the JVM via `ENTRYPOINT` (see [docker.md](../../guides/reference/docker.md)). If the project also uses [Cloud Profiler](../../guides/reference/profiler.md) (its own `-javaagent`/`-D` flags for the profiler agent), merge those into the same `ENTRYPOINT` array as the tracing flags rather than adding a second launch mechanism.
