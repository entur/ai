# Entur Platform Controls

Use this reference for applications deployed on Entur infrastructure or claiming
Entur compliance. Current central Entur platform, security, and policy
requirements take precedence. Repository-local `AGENTS.md` instructions may add
stricter controls but must not weaken them. Record a conflict as policy or
documentation drift instead of accepting the weaker rule or inventing a new
standard.

## Secret and workload identity controls

- Secrets belong in Google Secret Manager and reach Kubernetes through
  ExternalSecrets. They must not be committed in source, values, Terraform
  variables, CI configuration, or local test configuration.
- Use Workload Identity for service authentication. Verify that the runtime
  identity, not only a CI identity, has the intended least privilege, and flag
  long-lived service-account JSON keys.
- Check that default Kubernetes service-account token automounting is disabled
  unless the workload has a documented need.

## Authorisation and IAM controls

- Use Entur Permission Store and Permission Client for application-level
  capability and responsibility-set authorisation where applicable.
- Terraform IAM roles must come from the current Entur allowed-role list. Compare
  actual bindings with documented runtime needs and flag both excessive grants
  and misleading allow-list documentation.
- Pay special attention to `allUsers`, `allAuthenticatedUsers`, broad project
  roles, wildcard permissions, public buckets, cross-project data access, and
  runtime service accounts used as confused deputies.
- A public-looking principal is not enough evidence by itself. Verify ingress,
  load-balancer, gateway, organisation-policy, and application controls before
  assigning reachability.

## Provisioning and deployment controls

- GCP projects must come from Entur self-service manifests, never `google_project`
  resources or ad hoc project creation.
- Terraform must use the Entur modules required by repository guidance, with
  tagged module references and correctly isolated state.
- Kubernetes deployments must use the Entur `common` Helm chart. Verify probes,
  non-root execution, resource limits, Prometheus metrics, configuration, ingress,
  and ExternalSecrets in the effective values for each environment.
- Docker base images must use the pinning form required by current Entur
  conventions: a specific tag, never `latest`. Do not report a missing digest as
  an Entur violation when the standard requires tags.

## CI and software controls

- CI/CD steps must use Entur reusable GitHub Actions workflows. Current Entur
  conventions pin Actions and reusable workflows to major tags such as `@v2`;
  do not impose a conflicting full-SHA rule.
- Every PR must pass lint, unit tests, CodeQL, Docker scanning, and Helm lint under
  current Entur rules. Verify workflow triggers and permissions rather than
  merely checking filenames.
- Dependencies must be pinned according to the ecosystem, and Dependabot is the
  Entur default. A new tool or package recommendation must also follow the IT
  systems policy; do not claim approval or registration without evidence.

## Platform protection caveat

Headers, TLS, authentication, network policy, and audit logging may be provided
by a gateway, load balancer, common chart, or organisation policy. Conversely,
documentation may claim a platform control that is not enabled for this service.
Verify the effective deployment chain. Report an application omission only when
the control is absent or bypassable end to end.
