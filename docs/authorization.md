# Authorization with Permission Store

Entur uses **Permission Store** (backend) and **Permission Client** (Java/Node SDK) for centralized authorization.

## Overview

| Component | Repository | Purpose |
|-----------|-----------|---------|
| **Permission Store** | `permission-store` | Central backend storing permissions, agreements, responsibility sets |
| **Permission Client** (Java) | `permission-client` | Spring Boot SDK: caching, Spring Security integration |
| **Permission Client** (Node) | `permission-client-node` | Node.js SDK |

### How It Works

1. **Applications register** with Permission Store and declare required permissions (business capabilities)
2. **Users authenticate** via OIDC/JWT with tenant info (authority + subject)
3. **Permission Client** caches permissions locally and evaluates access
4. **Controllers** use `@PreAuthorize` annotations for endpoint-level access control

```text
User (JWT token)
       ↓
  Application
  (Permission Client SDK)
       ↓ polls/websocket
  Permission Store
  (central permission database)
```

## Permission Types

### Business Capabilities

Primary authorization model: **operation** + **access level**.

| Access Level | Norwegian | Description |
|-------------|-----------|-------------|
| `LES` | Les | Read access |
| `OPPRETT` | Opprett | Create access |
| `ENDRE` | Endre | Update/modify access |
| `SLETT` | Slett | Delete access |

Example: `product-api-access` + `LES,ENDRE` = can read and modify products, but not create or delete.

### Responsibility Sets

Data-level access control -- restricts who can access specific data objects. Combines:

- **Operation** + **Responsibility Type** (e.g., `product.organisation`) + **Responsibility Key** (e.g., org ID) + **Access Level**

Linked to users via **Agreements**. Use when restricting access to specific data partitions (e.g., "user X can edit products belonging to organisation Y").

## Adding Authorization to Your Application

### Prerequisites

Your app needs Auth0 internal M2M credentials. In your self-service manifest:

```yaml
spec:
  auth0:
    internal:
      enabled: true
      type: m2m
```

See [self-service.md](self-service.md) for details.

### 1. Add Dependencies

```kotlin
// build.gradle.kts
dependencies {
    implementation(libs.bundles.entur.auth)
    // This bundle typically includes:
    // - org.entur.auth.client:oidc-client-spring-boot
    // - org.entur.auth.resource-server:oidc-rs-spring-boot-web-config
    // - org.entur.auth:permission-client

    testImplementation(libs.entur.auth.oidc)
    // - org.entur.auth.resource-server:oidc-rs-spring-boot-web-test
}
```

Or explicitly:

```groovy
// build.gradle
implementation 'org.entur.auth:permission-client:3.x.x'
implementation 'org.entur.auth.resource-server:oidc-rs-spring-boot-web-config:2.x.x'
implementation 'org.entur.auth.client:oidc-client-spring-boot:2.x.x'
```

### 2. Configure application.yml

```yaml
entur:
  auth:
    tenants:
      environment: dev              # dev | tst | prd
      include: internal, partner    # Which tenants to accept tokens from

  # OIDC client credentials for M2M calls to Permission Store
  clients:
    auth0:
      permission-store:
        clientId: ${sm@MNG_AUTH0_INT_CLIENT_ID}
        secret: ${sm@MNG_AUTH0_INT_CLIENT_SECRET}
        domain: internal.dev.entur.org
        audience: https://api.dev.entur.io

  permission:
    scheduler: ws                    # ws (websocket) or poll
    bean: client                     # client | oidc | jwt
    permission-cache:
      type: IN_MEMORY                # IN_MEMORY | LOCAL_TEST_CACHE | FULL_ACCESS
      url: https://api.dev.entur.io/permission-store/v1
      application-name: my-application
      refresh-rate: 300              # Seconds between full cache refreshes
      auth-qualifier: "permission-store"
    businessCapabilities:
      - my-operation, LES, ENDRE     # Compact format: operation, access1, access2
```

### 3. Protect Endpoints

Use `@PreAuthorize` with `hasPermission()`:

```kotlin
@RestController
class VersionController(
    private val versionService: VersionService,
) : VersionApi {

    // Business capability check: operation + access
    @PreAuthorize("hasPermission('product-api-access', 'endre')")
    override fun createVersion(request: CreateVersionRequest): ResponseEntity<Version> {
        // Only users with 'product-api-access' + 'endre' access can reach this
    }

    @PreAuthorize("hasPermission('product-api-access', 'les')")
    override fun getVersionById(id: String): ResponseEntity<Version> {
        // Read-only access
    }
}
```

For responsibility set checks:

```java
@PreAuthorize("hasPermission(#organisationId, 'product.organisation', 'endre')")
public ResponseEntity<?> updateProduct(@PathVariable String organisationId, ...) {
    // Only users with responsibility set access for this organisation
}
```

### 4. Programmatic Access Checks

Inject `AuthorizeTenant`:

```java
@Service
public class MyService {
    private final AuthorizeTenant authorizeTenant;

    public MyService(AuthorizeTenant authorizeTenant) {
        this.authorizeTenant = authorizeTenant;
    }

    public void doSomething(Authentication auth) {
        // Check business capability
        boolean canEdit = authorizeTenant.checkBusinessCapabilityPermission(
            auth.getPrincipal(), "my-operation", "endre"
        );

        // Get all permissions for a user
        Set<Permission> permissions = authorizeTenant.getPermissions(auth.getPrincipal());

        // Check responsibility set
        boolean canAccessOrg = authorizeTenant.checkResponsibilitySetPermission(
            auth.getPrincipal(), "my-operation", new ObjectId("org-123"), "les"
        );
    }
}
```

## Granting Access Between Apps, Clients, and Partners

### Defining Business Capabilities

Declare in `application.yml`:

```yaml
entur:
  permission:
    businessCapabilities:
      # Compact format (preferred):
      - my-operation, LES, OPPRETT, ENDRE, SLETT

      # Verbose format:
      - operation: my-admin-operation
        access: LES, ENDRE
        description: "Administrative access to my-app"
```

These are registered with Permission Store on application start.

### Defining Responsibility Types

```yaml
entur:
  permission:
    responsibilityTypes:
      # Compact format (preferred):
      - product.organisation, LES, ENDRE

      # Verbose format:
      - name: product.id
        access: LES, OPPRETT, ENDRE, SLETT
```

### Permission Store API

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `POST` | `/applications` | Register application |
| `POST` | `/applications/{id}/permissions` | Set permissions for all tenants |
| `GET` | `/applications/{id}/permissions` | Get current permissions |
| `GET` | `/applications/{id}/changes` | Get permission changes (delta) |

### Managing Agreements (Responsibility Sets)

Agreements link responsibility sets to organisations for data-level access:

```java
// Programmatic agreement management via AuthorizeTenant
authorizeTenant.storeAgreement(
    Agreement.builder()
        .operation("product")
        .access(Access.ENDRE)
        .responsibilityType("organisation")
        .responsibilityKey("ENT:Organisation:123")
        .organisationId("org-456")
        .build()
);

// Query agreements
Set<Agreement> agreements = authorizeTenant.getAgreements(permission);

// Delete agreement
authorizeTenant.deleteAgreement(agreementId);
```

Agreement REST API:

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/agreements` | List agreements (with filters) |
| `POST` | `/agreements` | Create agreement(s) |
| `DELETE` | `/agreements` | Delete agreement |

## Cache Types

### IN_MEMORY (Production)

Default production cache. Polls or receives WebSocket push from Permission Store:

```yaml
entur:
  permission:
    scheduler: ws                  # Use WebSocket push (alternative: poll)
    permission-cache:
      type: IN_MEMORY
      url: https://api.dev.entur.io/permission-store/v1
      refresh-rate: 300            # Full refresh every 5 minutes
      delta-refresh-rate: 60       # Delta refresh every 1 minute (optional)
```

### LOCAL_TEST_CACHE (Testing)

Define test users with specific permissions directly in config:

```yaml
entur:
  permission:
    permission-cache:
      type: LOCAL_TEST_CACHE
    test-users:
      - authority: internal
        subject: fullAccess
        business-capability-permissions:
          - operation: product-api-access
            access: les, endre
      - authority: internal
        subject: readAccess
        business-capability-permissions:
          - operation: product-api-access
            access: les
      - authority: partner
        subject: orgAdmin
        organisationId: "1"
        responsibility-permissions:
          - operation: product
            responsibilityType: organisation
            access: les, endre
        user-information:
          email: test@example.com
          givenName: Test
          familyName: User
```

### FULL_ACCESS (Development)

Grants all permissions to all users. Local development only:

```yaml
entur:
  permission:
    permission-cache:
      type: FULL_ACCESS
```

## Test Configuration

### Controller Tests

Use Entur's test auth library with `TenantJsonWebToken`:

```kotlin
@WebMvcTest(VersionController::class)
@Import(VersionMapper::class)
@ExtendWith(TenantJsonWebToken::class)
class VersionControllerTests : BaseControllerTest() {

    @MockkBean
    lateinit var versionService: VersionService

    @Test
    fun `internal user with full access can read`(
        @InternalTenant(clientId = "fullAccess") token: String,
    ) {
        every { versionService.find("ENT:Version:1") } returns someVersion()

        mockMvc.perform(
            get("/v3/versions/ENT:Version:1")
                .header("Authorization", token)
        ).andExpect(status().isOk)
    }

    @Test
    fun `traveller tenant is unauthorized`(
        @TravellerTenant(clientId = "user") token: String,
    ) {
        mockMvc.perform(
            get("/v3/versions/ENT:Version:1")
                .header("Authorization", token)
        ).andExpect(status().isUnauthorized)
    }

    @Test
    fun `read-only user cannot create`(
        @InternalTenant(clientId = "readAccess") token: String,
    ) {
        mockMvc.perform(
            post("/v3/versions")
                .header("Authorization", token)
                .contentType(MediaType.APPLICATION_JSON)
                .content("""{"id": "ENT:Version:001", "status": "DRAFT"}""")
        ).andExpect(status().isForbidden)
    }
}
```

### Integration Tests

Use `LOCAL_TEST_CACHE` in `src/test/resources/application.yml`:

```yaml
entur:
  auth:
    tenants:
      include: partner, internal
      environment: mock              # Use mock OIDC providers
    lazy-load: true
  permission:
    permission-cache:
      type: LOCAL_TEST_CACHE
    test-users:
      - authority: internal
        subject: fullAccess
        business-capability-permissions:
          - operation: product-api-access
            access: les, endre
      - authority: internal
        subject: readAccess
        business-capability-permissions:
          - operation: product-api-access
            access: les
```

## Access Aliases

Map Norwegian access names to custom names:

```yaml
entur:
  permission:
    permission-cache:
      access-aliases:
        opprett: create
        les: read
        endre: change, admin       # Multiple aliases supported
        slett: delete
```

## Internal Has Full Access

Grant all permissions to internal (Entur employee) tenants automatically:

```yaml
entur:
  permission:
    permission-cache:
      internalHasFullAccess: true   # Default: false
```

Useful during development or for internal-only services.

## Environment-Specific Configuration

### Permission Store URLs

| Environment | Permission Store URL |
|-------------|---------------------|
| dev | `http://permission-store.dev.entur.internal` or `https://api.dev.entur.io/permission-store/v1` |
| tst | `http://permission-store.tst.entur.internal` or `https://api.staging.entur.io/permission-store/v1` |
| prd | `http://permission-store.prd.entur.internal` or `https://api.entur.io/permission-store/v1` |

Use internal URL (`.entur.internal`) for cluster-internal communication. Set via Helm:

```yaml
# helm/my-app/env/values-kub-ent-dev.yaml
common:
  configmap:
    data:
      ENTUR_PERMISSION_PERMISSIONCACHE_URL: "http://permission-store.dev.entur.internal"
```

### Auth0 Domains

| Environment | Domain | Audience |
|-------------|--------|----------|
| dev | `internal.dev.entur.org` | `https://api.dev.entur.io` |
| tst | `internal.staging.entur.org` | `https://api.staging.entur.io` |
| prd | `internal.entur.org` | `https://api.entur.io` |

Set via Helm:

```yaml
# helm/my-app/env/values-kub-ent-tst.yaml
common:
  configmap:
    data:
      ENTUR_AUTH_TENANTS_ENVIRONMENT: "tst"
      ENTUR_CLIENTS_AUTH0_PERMISSIONSTORE_AUDIENCE: "https://api.staging.entur.io"
      ENTUR_CLIENTS_AUTH0_PERMISSIONSTORE_DOMAIN: "internal.staging.entur.org"
```

## Authentication Modes

| Mode | Config | When to Use |
|------|--------|-------------|
| `oidc` | `bean: oidc` | Recommended for production with `oidc-auth-client` |
| `client` | `bean: client` | Direct client credentials (Auth0 M2M) |
| `jwt` | `bean: jwt` | JWT resource server without OIDC client |

```yaml
entur:
  permission:
    bean: client    # or oidc, jwt
```

## Permission Store Architecture

Spring Boot application backed by PostgreSQL. Key internals:

- **Hibernate Envers** for full audit trail
- **WebSocket (STOMP/SockJS)** for push notifications
- **ShedLock** for distributed scheduling
- **LZ4-compressed Kryo serialization** for efficient cache
- **Apigee API Gateway** for external access (1200 rpm rate limit)

### Domain Model

```text
Application
  └── ApplicationInstance
        ├── BusinessCapability (operation + access)
        │     └── BusinessCapabilityPermission (tenant binding)
        └── Responsibility (operation + responsibilityType)
              └── ResponsibilitySet (+ objectKey)
                    ├── ResponsibilityPermission (tenant binding)
                    └── Agreement (+ organisationId)
```

### Automatic Cleanup

Permission Store cleans up resources not refreshed in 30 days (applications, permissions, responsibility sets with no agreements). Permission Client handles automatic refresh.

## Quick Reference

### Minimal Setup Checklist

1. Ensure `auth0.internal.enabled: true` and `auth0.internal.type: m2m` in self-service manifest
2. Add `permission-client` + `oidc-auth` dependencies
3. Configure `entur.auth.tenants` in `application.yml`
4. Configure `entur.clients.auth0.permission-store` credentials
5. Configure `entur.permission.permission-cache` with URL and application name
6. Declare `businessCapabilities` used by your application
7. Add `@PreAuthorize("hasPermission('operation', 'access')")` to protected endpoints
8. Add `entur-springdoc-starter` to auto-document permissions in OpenAPI (see [api-design.md](api-design.md#x-entur-permissions-extension-automatic))
9. Configure `LOCAL_TEST_CACHE` with test users in test `application.yml`
10. Set environment-specific Permission Store URL and auth config via Helm configmap
