# Scenario: Framing - Namespace Derivation (Positive)

## Description

Tests whether positive framing leads to clean, correct namespace derivation from metadata.id.

## Prompt

You are given the following platform convention:

> ALWAYS derive the Kubernetes namespace from the metadata.id field. The namespace equals metadata.id.

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
    "products"
  ],
  "must_not_contain": [
    "namespace is products-api",
    "namespace: products-api"
  ],
  "must_match": [
    "metadata\\.id"
  ]
}
```

## Budget

0.02
