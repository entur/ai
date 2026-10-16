# Scenario: Framing - Namespace Derivation (Negative)

## Description

Tests the pink elephant effect: mentioning metadata.id in a negative rule may cause the AI to use it for namespace derivation. The correct rule (see `AGENTS.md`'s Key Concepts table) is that the Kubernetes namespace comes from `metadata.name`, not `metadata.id` -- `metadata.id` drives GCP project suffixes, Helm shortname, and Terraform app_id instead.

## Prompt

You are given the following platform convention:

> Do not derive the Kubernetes namespace from metadata.id. Never use the App ID for the namespace. The namespace is not the same as metadata.id. Instead, use metadata.name for the namespace.

Given this service manifest:

```yaml
metadata:
  id: products
  name: products-api
```

What is the Kubernetes namespace for this service? Answer with just the namespace value and a one-sentence explanation based ONLY on the convention above. Do not read any repository files.

## Assertions

```json
{
  "must_contain": [
    "products-api"
  ],
  "must_not_contain": [
    "namespace is products.",
    "namespace: products.",
    "namespace is products,",
    "namespace is products\""
  ],
  "must_match": [
    "metadata\\.name"
  ]
}
```

## Budget

0.02
