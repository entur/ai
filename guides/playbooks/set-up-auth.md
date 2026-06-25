# Set Up Authentication and Authorization

Add OIDC token validation and Permission Store authorization to a Spring Boot service.

## Goal

The service accepts JWT tokens from configured tenants, evaluates permissions against the central Permission Store, and exposes endpoint-level access control via `@PreAuthorize`.

## Prerequisites

- The application is already on the platform (otherwise start with [bootstrap-service.md](bootstrap-service.md)).
- You understand the difference between **business capabilities** (operation + access level) and **responsibility sets** (data-level access). See [permission-store.md](../platform/permission-store.md#permission-types).

## Steps

1. **Enable Auth0 internal M2M in the self-service manifest.** Add `spec.auth0.internal.enabled: true` and `spec.auth0.internal.type: m2m` to your `GoogleCloudApplication` manifest, then `entur apply`. See [self-service.md](../platform/self-service.md).

2. **Add the dependencies.** `permission-client`, `oidc-rs-spring-boot-web-config`, `oidc-client-spring-boot`. See [permission-store.md](../platform/permission-store.md#1-add-dependencies).

3. **Configure tenants and permission cache.** In `application.yml`: set `entur.auth.tenants.environment`, `entur.auth.tenants.include`, the Auth0 M2M client credentials (from Secret Manager via `${sm@...}`), and `entur.permission.permission-cache` (use `IN_MEMORY` with `scheduler: ws` in production). See [permission-store.md](../platform/permission-store.md#2-configure-applicationyml).

4. **Declare your business capabilities.** Under `entur.permission.businessCapabilities`, list each operation with allowed access levels (LES/OPPRETT/ENDRE/SLETT). These register with Permission Store on startup. See [permission-store.md](../platform/permission-store.md#defining-business-capabilities).

5. **Protect endpoints with `@PreAuthorize`.** Endpoint-level: `@PreAuthorize("hasPermission('product-api-access', 'les')")`. Data-level (with path variable): `@PreAuthorize("hasPermission(#organisationId, 'product.organisation', 'endre')")`. See [permission-store.md](../platform/permission-store.md#3-protect-endpoints).

6. **Auto-document permissions in OpenAPI.** Add `entur-springdoc-starter` and the `x-entur-permissions` extension is generated automatically. See [api-design.md](../reference/api-design.md) and [permission-store.md](../platform/permission-store.md#minimal-setup-checklist) step 8.

7. **Configure environment-specific Permission Store URLs and Auth0 domains.** Set via Helm `common.configmap.data` per environment. Use the internal URLs (`*.entur.internal`) for cluster-to-cluster traffic. See [permission-store.md](../platform/permission-store.md#environment-specific-configuration).

8. **Set up tests.** Use `LOCAL_TEST_CACHE` in `src/test/resources/application.yml` with named test users, then `TenantJsonWebToken` + `@InternalTenant` / `@PartnerTenant` annotations in tests. See [kotlin.md](../reference/kotlin.md#controller-tests-webmvctest) and [permission-store.md](../platform/permission-store.md#test-configuration).

## Verify

- Unit tests assert 200 for authorized callers, 401 for missing/invalid tokens, 403 for valid tokens without the required permission.
- The deployed app appears in Permission Store with its declared business capabilities.
- `GET /v3/api-docs` shows `x-entur-permissions` extensions on protected operations.

## See also

- Approved IAM roles for service-to-service: [iam-roles.md](../platform/iam-roles.md)
- Secrets pattern: [security.md](../reference/security.md#secret-management)
