# Scenario: Framing - Namespace Derivation (Negative)

## Description

Tests the pink elephant effect: mentioning metadata.name in a negative rule may cause the AI to use it for namespace derivation.

## Prompt

You are given the following platform convention:

> Do not derive the Kubernetes namespace from metadata.name. Never use the application name for the namespace. The namespace is not the same as metadata.name. Instead, use metadata.id for the namespace.

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
