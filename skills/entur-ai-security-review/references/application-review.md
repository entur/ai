# Application Security Review

Use this baseline for every security review. Select checks from the deployed
attack surface and trace high-risk code paths end to end; do not report keyword
matches without evidence.

## Security claims and deployment reality

- Compare the README, architecture diagrams, security documents, comments, and
  tests with registered entrypoints and deployed configuration.
- Verify that claimed middleware, gateways, feature flags, and environment
  restrictions are active on every relevant path.
- Distinguish production paths from local, test, migration, ingest-only, latent,
  and dead code. A false security claim can be independently harmful when a team
  relies on it, but documentation drift usually belongs with its underlying root
  cause.
- Look for infrastructure or permissions provisioned before the feature that
  needs them and for code deployed before its documented control is wired.

## Secrets and credentials

- Inspect source, configuration, CI, generated frontend assets, examples, test
  fixtures, logs, and relevant Git history for credential material.
- Trace how secrets enter the workload, which identity can read them, whether
  they are logged or exposed to child processes, and how they rotate.
- Treat any committed real credential as burned. Severity depends on whether it
  is live, its privileges, exposure, lifetime, and reachable systems.
- Distinguish obvious placeholders and isolated non-production fixtures from
  credentials that can reach shared or production resources.
- Never print or copy the full value. Report the secret type, file and line,
  exposure path, and a redacted fingerprint only when needed for identification.

## Authentication and session integrity

- Establish where identity originates and where signature, issuer, audience,
  expiry, and relevant tenant or domain claims are enforced.
- Check every entrypoint, including administrative, callback, debug, health,
  WebSocket, background-job, message-consumer, and tool endpoints.
- Reject identity, role, tenant, organisation, or admin status taken from
  attacker-controlled parameters without binding to authenticated context.
- Check session fixation, logout and revocation behavior, token leakage, cookie
  flags, CSRF protection for cookie-authenticated state changes, and fail-open
  error handling.
- Verify production cannot enable test users, bypass headers, mock identity, or
  debug authentication through missing or malformed configuration.

For services that implement OAuth or token handling, also read
[OAuth and token services](oauth-token-services.md).

## Authorisation and personal data

- Trace object-, tenant-, organisation-, capability-, and field-level access on
  every read and write. Do not reduce authorisation to record ownership when
  Entur responsibility sets or other attributes apply.
- Prefer enforcement in the query or trusted service boundary so callers cannot
  forget a post-fetch check. Verify bulk, export, search, count, attachment, and
  error paths as well as single-record endpoints.
- Test identifier substitution, list filtering, nested-resource access, confused
  deputy behavior, mass assignment, over-broad response fields, and existence
  leaks through status codes or timing.
- Identify data sensitivity, volume, affected subjects, purpose, retention, and
  deletion path. Travel history, employee data, credentials, payment data, and
  linkable identifiers can have high impact, but personal data does not
  automatically set severity without scope and exploitability evidence.
- Do not declare a GDPR breach or lawful basis from code inspection alone.
  Describe the exposure and recommend immediate security and privacy assessment
  when production personal data may be affected.

## Injection and unsafe interpretation

Follow untrusted input into interpreters and privileged APIs:

- SQL, JPQL, BigQuery SQL, search filters, PromQL, Cloud Logging filters, and
  template or expression languages;
- HTML and browser DOM sinks, redirects, response headers, and spreadsheet
  exports;
- shell commands, subprocess arguments, dynamic code, unsafe deserialisation,
  reflection, and template compilation;
- file paths, archive extraction, object keys, repository identifiers, and
  configuration loaders.

Confirm parameterisation or context-specific encoding at the actual sink.
Allow-listing and parsing must cover alternate encodings, duplicated parameters,
normalisation, and error paths. When a caller's own IAM bounds an injected cloud
query, reflect that reduced blast radius rather than ignoring the bug or
overrating it.

## Server-side requests and redirects

- Distinguish a URL the server fetches from a URL it only stores or returns.
- For fetched URLs, check scheme and host allow-lists, credential forwarding,
  redirect revalidation, DNS resolution and rebinding, internal and link-local
  ranges, metadata endpoints, proxy behavior, and response-size or timeout
  limits.
- For redirects, require semantic URL parsing and an allow-list appropriate to
  the protocol. Prefix, suffix, user-info, encoded-separator, fragment, and open
  redirect tricks commonly defeat string heuristics.
- Check webhook callbacks, importers, image or document fetchers, health checks,
  preview generators, and cloud API proxy endpoints.

## Files and resource exhaustion

- Enforce size and count limits before expensive parsing or buffering.
- Validate file type server-side using the format actually consumed. Treat the
  original name as untrusted, generate a storage name, and keep uploads outside
  executable or web-served paths unless intentionally published.
- Check archive traversal, compression bombs, parser vulnerabilities, temporary
  file permissions, malware handling, and cleanup after failures.
- Bound pagination, concurrency, retries, regular expressions, decompression,
  serialization, cache entries, queue depth, and response size.
- Verify rate-limit keys come from trusted identity or proxy context. On a load
  balancer, do not assume the leftmost `X-Forwarded-For` value is trustworthy;
  validate the configured proxy chain and cap limiter state independently.

## Dependencies and CI supply chain

- Inspect lockfiles, version constraints, repositories, checksums, build plugins,
  code generators, container bases, and GitHub Actions.
- Treat age, popularity, or maintainer count as investigation signals, not
  vulnerabilities. For an advisory, establish the resolved version and whether
  the vulnerable feature is reachable.
- Check dependency-confusion and typosquat risk, untrusted registries, mutable
  artifacts, install hooks, excessive SDK privileges, and generated artifacts
  that are accepted without verification.
- Inspect workflow triggers, permissions, secret scope, artifact trust, and use
  of fork-controlled code. In particular, reject `pull_request_target` flows
  that execute untrusted pull-request content with secrets or write tokens.
- Prefer existing CI scan results. Before running a package manager, build, or
  scanner, inspect what it executes and follow the main skill's safety boundary.

## Logging, errors, and operational security

- Check logs, metrics labels, traces, analytics, crash reports, and audit sinks
  for tokens, request bodies, user-controlled free text, personal data, and
  high-cardinality attacker-controlled values.
- Verify retention and access controls for log buckets and long-term sinks.
- Keep application audit events sufficient for accountability without logging
  sensitive payloads; verify platform Data Access audit logs where required.
- Confirm client errors do not expose stack traces, queries, filesystem paths,
  credentials, internal topology, or cross-tenant existence.
- Trace retry storms, fail-open fallbacks, health-check accuracy, graceful
  dependency failure, backup and restore assumptions, and denial-of-service
  paths that matter to the deployed service.

## HTTP and browser controls

- Verify effective HTTPS enforcement, HSTS, content-type and nosniff headers.
- For browser applications, check content security policy, clickjacking defense,
  output encoding, secure cookies, CSRF, and trusted-origin handling.
- CORS must match the credential model and intended clients. A wildcard is not
  automatically a finding for a public credential-free API, and browser rejection
  is not a substitute for correct server-side authorisation.
- Rate-limit authentication, recovery, expensive search, export, and other
  abuse-prone operations according to identity and business impact.
