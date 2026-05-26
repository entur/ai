# Add a Custom Domain

Expose a Kubernetes service on a custom hostname under `entur.no`, `entur.io`, or `entur.org` (the only domains the platform serves). For Firebase Hosting, Cloud Run with a custom load balancer, or any setup off the GKE golden path, see [Off the golden path](#off-the-golden-path) below.

## Goal

A request to `https://<app>.entur.io` (or `.no` / `.org`) reaches your service in prd, with a valid Google-managed TLS certificate, and the dev/tst hostnames resolve to the corresponding non-prd environments.

## Prerequisites

- The application is already on the platform and deploys via the Entur `common` Helm chart (otherwise start with [bootstrap-service.md](bootstrap-service.md)).
- You know which domain root the service belongs under (`entur.no`, `entur.io`, or `entur.org`). The platform does not serve hostnames outside these three.
- For public/internet-facing services: confirm with the platform team that the service is allowed to be public.

## Hostname pattern

| Environment | Hostname |
|-------------|----------|
| `dev`       | `<app>.dev.entur.{no,io,org}` |
| `tst`       | `<app>.staging.entur.{no,io,org}` |
| `prd`       | `<app>.entur.{no,io,org}` |

Note that the **tst** environment uses the `staging` token in the hostname, not `tst`. The dev environment uses `dev`. The prd environment uses no environment token at all.

## Steps

1. **Pick a hostname.** Use the pattern above. The `<app>` portion should match `metadata.name` from your self-service manifest (the same value used as `common.app` and the Kubernetes namespace). See [self-service.md](../platform/self-service.md) for the identity chain.

2. **Set `common.ingress.host` per environment in Helm values.** Example for a `products-api` service on `entur.io`:

   ```yaml
   # helm/<app>/env/values-kub-ent-dev.yaml
   common:
     ingress:
       host: products-api.dev.entur.io
   ```

   ```yaml
   # helm/<app>/env/values-kub-ent-tst.yaml
   common:
     ingress:
       host: products-api.staging.entur.io
   ```

   ```yaml
   # helm/<app>/env/values-kub-ent-prd.yaml
   common:
     ingress:
       host: products-api.entur.io
   ```

   See [common-helm.md](../platform/common-helm.md#ingress) for the full ingress reference.

3. **Choose the right `trafficType`.** Set this in the base `values.yaml`:

   - `api` (default) -- internal API traffic; not internet-reachable from outside the Entur network.
   - `public` -- internet-facing; required for any hostname that users hit directly from the public internet.
   - `http2` -- gRPC / HTTP/2 backend.

   See [common-helm.md](../platform/common-helm.md#ingress) for the table of traffic types.

4. **Deploy.** The platform's shared ingress automatically picks up the host, DNS resolves via the `entur.{no,io,org}` zones, and Google provisions a managed TLS certificate. No DNS record changes, no certificate management, no platform-team ticket are needed for this path.

## TLS certificates

The platform uses **Google-managed certificates** exclusively. They are issued and renewed automatically once `common.ingress.host` is set and DNS has converged. Entur does **not** provision custom TLS certificates -- if a third party asks for a custom cert, the answer is that the platform only serves Google-managed certs.

> The certificate may take up to ~60 minutes to provision the first time. Until it is `ACTIVE`, browsers will show a TLS warning. The hostname is correct; just wait.

For **developer / code-signing certificates** (a separate concern, not service TLS) contact **Team Sikkerhet** in [#talk-utviklerplattform](https://entur.slack.com/archives/talk-utviklerplattform). They are issued and tracked by chat in that channel.

## Off the golden path

For setups that are **not** deployed via the common Helm chart on GKE, the platform team configures the domain manually. Ask in [#talk-utviklerplattform](https://entur.slack.com/archives/talk-utviklerplattform) when your service is one of:

- **Firebase Hosting** -- the Firebase project hosts the static site and the platform team wires the domain into Firebase Hosting's custom-domain flow.
- **Cloud Run with a custom external HTTPS load balancer** -- you provision the LB in your Terraform, then the platform team adds the DNS record and validates managed-cert provisioning.
- **Anything else under `entur.{no,io,org}`** -- ask first; the platform may already have a pattern, or may need to register the host.

Even on these paths, TLS is still Google-managed; teams do not provision custom certificates.

## Verify

- `kubectl get ingress -n <app>` shows your hostname with an attached managed certificate after deploy.
- `curl -I https://<app>.dev.entur.io/health/liveness` returns `200 OK` with a valid certificate.
- The Google Cloud console under **Network services → Load balancing → Certificates** lists your hostname with status `ACTIVE`.
- Browser: visit the URL, confirm there is no certificate warning.

If after ~60 minutes the cert is still not `ACTIVE`, check that `common.ingress.host` is exactly the hostname pattern above (typos like `prd.entur.io` or `.test.entur.io` will not match the managed zones) and that `trafficType` is set. Then ask in [#talk-utviklerplattform](https://entur.slack.com/archives/talk-utviklerplattform).

## See also

- [common-helm.md](../platform/common-helm.md#ingress) -- full ingress configuration reference
- [bootstrap-service.md](bootstrap-service.md) -- standing up a new service from scratch
- [deploy-to-prd.md](deploy-to-prd.md) -- promoting to prd, where the bare-hostname (no env token) form applies
