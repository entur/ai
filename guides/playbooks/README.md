# Playbooks

End-to-end playbooks for tasks that span multiple platform capabilities. Each playbook is short by design: it names the goal, points at the relevant [`platform/`](../platform/) capability docs and [`reference/`](../reference/) standards, and ends with a verification step.

Use these as the agent's entry point for any non-trivial task. They exist precisely so the agent does not have to assemble a workflow from 4--5 unrelated guides.

## Available Playbooks

| Playbook | Use when… |
|----------|-----------|
| [bootstrap-service.md](bootstrap-service.md) | Standing up a brand-new application on the platform |
| [add-postgres.md](add-postgres.md) | Adding a managed PostgreSQL database to an existing service |
| [add-redis.md](add-redis.md) | Adding Memorystore Redis for caching, locks, or dedup |
| [add-kafka.md](add-kafka.md) | Producing or consuming events via Aiven Kafka |
| [set-up-auth.md](set-up-auth.md) | Adding OIDC + Permission Store authorization |
| [add-custom-domain.md](add-custom-domain.md) | Exposing a service on a `*.entur.{no,io,org}` hostname with managed TLS |
| [deploy-to-prd.md](deploy-to-prd.md) | Promoting a service from dev/tst to production |
| [deprecate-service.md](deprecate-service.md) | Retiring an application gracefully |
| [local-dev.md](local-dev.md) | Running the application locally with the right tooling |

## How a playbook is structured

1. **Goal** -- one sentence.
2. **Prerequisites** -- what must already be true.
3. **Steps** -- numbered, each step linking to the canonical reference.
4. **Verify** -- how to know it worked.
5. **See also** -- adjacent playbooks or deeper references.

If something belongs in multiple playbooks, document it once in [`platform/`](../platform/) or [`reference/`](../reference/) and link from each playbook.
