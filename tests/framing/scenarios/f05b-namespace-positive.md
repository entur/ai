# Scenario: Framing - Namespace Derivation (Positive)

## Description

Tests whether positive framing leads to clean, correct namespace derivation from metadata.name. The correct rule (see `AGENTS.md`'s Key Concepts table) is that the Kubernetes namespace comes from `metadata.name`, not `metadata.id`.

## Prompt

You are given the following platform convention:

> ALWAYS derive the Kubernetes namespace from the metadata.name field. The namespace equals metadata.name.

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
