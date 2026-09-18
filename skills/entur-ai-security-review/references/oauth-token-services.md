# OAuth and Token Services

Use this reference when a service implements login callbacks, acts as an OAuth or
OpenID Connect client, issues tokens, verifies tokens itself, or exposes
authorization-server endpoints. Prefer established Entur libraries and inspect
the security-critical implementation in full.

Use RFC 9700 as the general OAuth 2.0 security baseline. For JavaScript or
WebAssembly applications acting as OAuth clients in a browser, also apply RFC
10017, which expands and tightens RFC 9700 for browser-specific threats and
architecture choices. RFC 10017 does not cover a same-domain web application
that uses OpenID Connect only for federated login and then maintains a normal
server-side application session.

## Token verification and identity binding

- Verify signatures with an expected algorithm and trusted key set. Reject
  algorithm confusion, unsigned tokens, attacker-selected keys, unsafe `kid` or
  `jku` resolution, and decode-only validation.
- Validate issuer, audience, expiry, not-before time where used, token type, and
  required tenant or domain claims. Bind the resulting subject to the correct
  client, session, organisation, and authorisation context.
- Prevent token substitution between ID tokens, access tokens, refresh tokens,
  authorization codes, state values, cursors, and other signed envelopes.
  Separate keys or include and verify an explicit type and audience.
- Use constant-time comparison for secrets, MACs, authenticators, and other
  secret-derived values. Redirect URI comparison instead requires protocol-correct
  exact matching; it is not a timing-sensitive secret comparison.

## Redirect flows and request binding

- Require exact redirect URI matching except where the protocol explicitly
  permits a narrow exception. Test user-info, suffix confusion, encoded path
  separators, fragments, duplicate parameters, and callback open redirects.
- Bind each authorization response to the initiating browser session. Verify
  transaction-specific `state`, OpenID Connect `nonce`, or another protocol-valid
  CSRF defense and reject replay.
- Require PKCE where the client and authorization-server profile call for it,
  use `S256`, bind the challenge to the code, validate the verifier at exchange,
  and reject downgrade behavior.
- If more than one issuer is supported, verify mix-up defenses and bind endpoint
  metadata to the expected issuer.
- Derive public callback and metadata URLs from trusted configuration or a
  correctly constrained proxy setup, not arbitrary inbound host headers.

## Browser-based OAuth clients

For a browser-based OAuth client, review these RFC 10017 controls in addition to
the RFC 9700 baseline:

- Identify whether the application uses a backend for frontend (BFF), a
  token-mediating backend, or a browser-only public client. Record why the chosen
  pattern fits the threat model. A BFF offers the strongest token-confidentiality
  properties, but the standard presents architectural trade-offs rather than a
  universal requirement to use one.
- Require the Authorization Code grant with PKCE for public browser clients and
  verify that the authorization server enforces PKCE. Reject the Implicit and
  Resource Owner Password Credentials grants.
- Do not treat a shared secret shipped in browser code as confidential or as
  proof of client identity.
- For a BFF, keep access and refresh tokens out of browser-accessible code and
  storage. Protect the session cookie with `Secure` and `HttpOnly`, evaluate
  `SameSite=Strict`, avoid a broad `Domain`, constrain its path, and implement a
  protocol-appropriate CSRF defense.
- Ensure a BFF or token-mediating backend cannot become an open proxy. Allow only
  intended resource-server hosts, paths, and HTTP methods, and never forward a
  user's access token to an attacker-controlled destination.
- If the browser handles tokens directly, examine malicious-JavaScript threats,
  third-party scripts, CSP, origin isolation, and every storage choice. Persistent
  stores such as `localStorage` are readable by same-origin malicious code, and
  isolating stored tokens does not stop compromised application code from
  initiating new flows or proxying authorised requests.
- When browser clients receive refresh tokens, require rotation or sender
  constraint, set a maximum or inactivity lifetime, and ensure rotation does not
  extend an already bounded overall lifetime.

## Token lifecycle and storage

- Keep authorization codes short-lived and single-use. Scope access tokens to
  the intended resource and privileges.
- Check refresh-token rotation or sender constraints where applicable, replay
  detection, revocation, logout semantics, and behavior during key rotation.
- Source signing and client secrets from Secret Manager, enforce suitable key
  strength, avoid logging or placing tokens in URLs, and define safe rollover.
- Reject a token whenever its validity cannot be established. A temporary
  discovery or JWKS refresh outage may continue using still-valid, previously
  trusted cached metadata; introspection and revocation behavior depends on the
  protocol and token type. Bound caches and refreshes so attacker-selected
  identifiers cannot cause unbounded key fetches.

## Avoid architecture assumptions

Do not flag a secondary identity store merely because it exists. Trace whether it
allows identity substitution, weakens upstream assurance, creates an ungoverned
account lifecycle, or bypasses Entur authorisation. Rate the demonstrated
security consequence.
