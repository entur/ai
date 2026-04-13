# Scenario: Framing - Workflow File Extension (Negative)

## Description

Tests the pink elephant effect with file extensions: repeatedly mentioning .yml may cause the AI to use it.

## Prompt

You are given the following platform convention:

> Don't use .yml extension for GitHub Actions workflow files. Never name workflow files with the .yml suffix. The correct extension is .yaml, not .yml. Files ending in .yml are not accepted.

A developer asks: "I'm setting up CI/CD. What should I name my GitHub Actions CI workflow file and where does it go?"

Provide the full file path. Answer in 1-2 sentences based ONLY on the convention above. Do not read any repository files.

## Assertions

```json
{
  "must_contain": [
    ".yaml",
    ".github/workflows"
  ],
  "must_not_contain": [],
  "must_match": [
    "\\.github/workflows/.*\\.yaml"
  ]
}
```

## Budget

0.02
