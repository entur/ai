---
name: setup-tracing-java
description: >
    Wire distributed tracing into an Entur Kotlin/Java (Spring Boot) service
    using the OpenTelemetry Java Agent, following the golden path to Cloud
    Trace. Use when the user says "add tracing", "set up OpenTelemetry",
    "instrument for Cloud Trace", or "add distributed tracing" for a Kotlin,
    Java, or Spring Boot service -- typically a repo with `build.gradle.kts`.
    For a Go service, use the `setup-tracing-go` skill instead.
---

# Set Up Tracing -- Kotlin/Java (Golden Path)

Wire distributed tracing into an Entur Kotlin/Java service so every inbound request produces a span in Cloud Trace, correlated with structured logs. This golden path covers **Spring Boot via the OpenTelemetry Java Agent** only -- it does not cover Go, Python, or any other language.

## Step 0: Confirm prerequisites -- do not guess any of these

Ask the user directly; do not infer silently and do not proceed until all four are answered.

1. **Language.** Detect from the repo: `build.gradle.kts` → Kotlin/Java. State the detected language and ask the user to confirm it.
   - If the repo is Go (`go.mod`), **stop** and point the user to the `setup-tracing-go` skill instead.
   - If the repo is neither Kotlin/Java nor Go (Python, Node, etc.), **stop**. Tell the user this golden path only documents Java Agent instrumentation for Kotlin/Java -- do not improvise OpenTelemetry setup for another language.
2. **Cloud Trace storage (manual, console-only).** Ask: *"Have you already enabled Cloud Trace storage in the GCP Console (Monitoring → Trace Explorer → Enable trace storage) for every environment you're setting up, in the application project `ent-<app>-<env>` -- not the cluster host project `ent-kub-<env>`?"*
   - If no, or unsure, **stop** and tell them to do this first, once per environment (dev, tst, prd). This is a console-only action -- there is no Terraform resource for it yet. Do not attempt to script it.
   - Only proceed with the environments the user confirms are done. If they've only enabled it for `dev`, scope the rest of this skill to `dev`.
3. **App ID.** Needed to build `ent-<app>-<env>`. Read `metadata.id` from the self-service manifest under `.entur/*.yaml` if present. If the manifest exists but `metadata.id` is missing or empty, ask the user for it instead of guessing -- do not derive it from the repo name or any other field. If no manifest exists at all, ask the user directly.
4. **Runtime.** Kubernetes (`helm/<app>/env/values-kub-ent-<env>.yaml`) or Cloud Run (`cloudrun.yaml`)? Determines where env vars in Step 4 are set.

## Step 1: Enable the Cloud Trace API in Terraform

Add `cloudtrace.googleapis.com` to the per-environment service-activation block:

```hcl
# terraform/main.tf
resource "google_project_service" "services" {
  for_each = toset([
    "cloudtrace.googleapis.com",
    # ... other APIs your service needs
  ])
  project            = module.init.app.project_id # "init" assumes that's this repo's terraform-google-init alias
  service            = each.value
  disable_on_destroy = false
}
```

If the repo has no `terraform/` directory yet, do not create one ad hoc -- tell the user to bootstrap Terraform with the Entur `terraform-google-init` module first.

## Step 2: Grant the workload service account `roles/cloudtrace.agent`

```hcl
# terraform/main.tf
resource "google_project_iam_member" "runtime_cloudtrace_agent" {
  project    = module.init.app.project_id # adapt "init" to this repo's module alias
  role       = "roles/cloudtrace.agent"
  member     = "serviceAccount:${module.init.service_accounts.runtime.email}"
  depends_on = [google_project_service.services]
}
```

Bind exactly this role -- it is write-only. Never bind `roles/cloudtrace.user`, `.admin`, or `.editor` to the workload; read access is granted separately at the folder level.

## Step 3: Instrument the application

### OpenTelemetry Java Agent (default; do not hand-instrument without a specific reason)

Attach the agent via `-javaagent` in a multi-stage Dockerfile -- a temporary Alpine stage downloads the JARs, the distroless final stage only carries the JARs themselves:

```dockerfile
# Dockerfile
FROM alpine:3.24 AS otel
RUN mkdir /otel && \
    wget -q -O /otel/opentelemetry-javaagent.jar \
      https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v2.29.0/opentelemetry-javaagent.jar && \
    wget -q -O /otel/gcp-auth-extension.jar \
      https://repo1.maven.org/maven2/io/opentelemetry/contrib/opentelemetry-gcp-auth-extension/1.58.0-alpha/opentelemetry-gcp-auth-extension-1.58.0-alpha-shadow.jar

FROM gcr.io/distroless/java25-debian13:nonroot
WORKDIR /app
COPY --from=otel /otel /otel
COPY build/libs/app.jar app.jar
CMD ["-javaagent:/otel/opentelemetry-javaagent.jar", \
     "-Dotel.javaagent.extensions=/otel/gcp-auth-extension.jar", \
     "-Dotel.javaagent.logging=application", \
     "-jar", "/app/app.jar"]
```

- Both JARs are required: `opentelemetry-javaagent.jar` does the bytecode instrumentation; `gcp-auth-extension` attaches a Workload Identity token to outbound OTLP calls -- without it, `telemetry.googleapis.com` rejects the export as unauthenticated.
- Use the pinned versions above as-is -- `v2.29.0` (agent) / `1.58.0-alpha` (gcp-auth-extension). Do not spend time checking upstream for a newer release. In Step 6's summary, tell the user which versions were used and point to the `otel` build stage in the Dockerfile as where to bump them later.
- `java25-debian13` is an example tag, not a guarantee -- match the distroless image version to the project's own Java toolchain (`build.gradle.kts`/`.tool-versions`), and confirm the resulting tag actually exists before using it. Flag the tag used in the Step 6 summary as something the user may need to change.
- If the project also uses Cloud Profiler, merge its `-javaagent`/flags into this **same** `CMD` -- a Dockerfile only honors its last `CMD`, a second one silently disables the first.

## Step 4: Set runtime environment variables

Set on Helm values (`values-kub-ent-<env>.yaml`, under `common.container.env`) for Kubernetes, or on the container spec (`cloudrun.yaml`) for Cloud Run. `GCP_PROJECT_ID` is always the application's per-env project (`ent-<app>-<env>`) -- never the cluster host project.

**If `helm/<app>/env/values-kub-ent-<env>.yaml` doesn't exist yet for a confirmed environment**, ask the user: *"No Helm values file exists yet for `<env>` -- do you want me to create `helm/<app>/env/values-kub-ent-<env>.yaml` with the tracing env vars?"* Only create it if they say yes, and only for the environment(s) confirmed in Step 0. Use this shape:

```yaml
# helm/<app>/env/values-kub-ent-<env>.yaml
common:
  env: <env>
  container:
    env:
      - name: OTEL_TRACES_EXPORTER
        value: otlp
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: https://telemetry.googleapis.com
      - name: OTEL_SERVICE_NAME
        value: <service-name>
      - name: TRACING_ENABLED
        value: "true"
      - name: GCP_PROJECT_ID
        value: ent-<app>-<env>
      - name: OTEL_TRACES_SAMPLER
        value: parentbased_always_on
      - name: OTEL_METRICS_EXPORTER
        value: none
      - name: OTEL_LOGS_EXPORTER
        value: none
```

If the file already exists with its own `common.container.env` entries, append these to the existing list instead of overwriting the file -- check for entries with the same `name` first and update their `value` in place rather than duplicating. Omit `OTEL_TRACES_SAMPLER` on Cloud Run, per the note below.

Only create the single env file this way -- do not scaffold `helm/<app>/Chart.yaml` or `helm/<app>/values.yaml` ad hoc. If those don't exist either, Helm hasn't been bootstrapped for this service at all, which is out of scope for this skill (same rule as Step 1's Terraform bootstrap).

**Java Agent auto-configures from env vars.** On Cloud Run, set the same variables shown above on the `cloudrun.yaml` container spec, with one difference: omit `OTEL_TRACES_SAMPLER` entirely.

Always set `OTEL_METRICS_EXPORTER=none` and `OTEL_LOGS_EXPORTER=none` on both runtimes -- the agent defaults both to `otlp` like traces, but Step 2's IAM grant doesn't authorize either.

Do not set `OTEL_TRACES_SAMPLER` on Cloud Run: the Cloud Run load balancer injects its own (thin) sampling decision upstream, and `parentbased_always_on` would defer to it, sampling far less than intended. Omitting the variable lets the agent default to `always_on`.

## Step 5: Correlate logs with traces

Every log line inside a traced request needs `logging.googleapis.com/trace` (`projects/<project>/traces/<id>`), `logging.googleapis.com/spanId`, and `logging.googleapis.com/trace_sampled` -- otherwise Cloud Logging can't join it to the trace in Trace Explorer.

Use `entur/cloud-logging` with `spring-boot-starter-gcp-web` and plain SLF4J logging. The fields are injected automatically via Micrometer Tracing -- no manual extraction.

## Step 6: Tell the user what's left

Steps 0-5 are everything this skill can do by editing the repo. Verifying traces actually arrive requires a live deploy and real traffic, which happens outside this skill -- tell the user to:

1. Commit and merge the changes, and let the normal CD pipeline deploy to the confirmed environment(s) from Step 0.
2. Send a few requests to the deployed service to generate spans.
3. Check **Trace → Trace Explorer** in the GCP Console under `ent-<app>-<env>` (not `ent-kub-<env>`, even for Kubernetes workloads -- traces always land in the application project).
4. If no spans show up: re-check that Step 0's trace storage was enabled for the *deployed* environment, and that `TRACING_ENABLED`/`GCP_PROJECT_ID` reached the running container (a common miss is setting them in the wrong Helm values file for the target environment).

Also state in the summary: the OpenTelemetry Java agent (`v2.29.0`) and gcp-auth-extension (`1.58.0-alpha`) versions were pinned as-is, not checked against upstream for something newer, and are set in the Dockerfile's `otel` build stage if the user wants to bump them later. Also state the distroless base image tag used and that it may need to change to match the project's Java toolchain.

## Critical Rules

- **The Step 0 trace-storage check is a manual console action, once per environment.** Never attempt to Terraform it; never proceed with Steps 1-5 for an environment the user hasn't confirmed is enabled.
- **Grant exactly `roles/cloudtrace.agent`** to the workload -- never a broader trace role.
- **`GCP_PROJECT_ID` is always the application project (`ent-<app>-<env>`)**, never the cluster host project, even on Kubernetes.
- **Defaults to the Java Agent.** Only hand-roll manual OpenTelemetry instrumentation if the user gives a specific reason.
- **One `CMD` per Dockerfile.** Merge Cloud Profiler flags into the same `CMD` as the tracing agent flags if present.
- **Kubernetes vs Cloud Run sampler differs** -- see Step 4. Getting this backwards silently under-samples on Cloud Run or double-samples on Kubernetes.
- **Also disables metrics/logs export** -- see Step 4. They default to `otlp` too if left unset, but Step 2 doesn't authorize either.
