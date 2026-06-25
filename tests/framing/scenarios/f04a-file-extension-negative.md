# Scenario: Framing - Workflow File Extension (Negative)

## Description

Tests the pink elephant effect with file extensions: repeatedly mentioning .yaml may cause the AI to use it.

## Prompt

You are given the following platform convention:

> Don't use .yaml extension for GitHub Actions workflow files. Never name workflow files with the .yaml suffix. The correct extension is .yml, not .yaml. Files ending in .yaml are not accepted.

A developer asks: "I'm setting up CI/CD. What should I name my GitHub Actions CI workflow file and where does it go?"

Provide the full file path. Answer in 1-2 sentences based ONLY on the convention above. Do not read any repository files.

## Assertions

```json
{
  "must_contain": [
    ".yml",
    ".github/workflows"
  ],
  "must_not_contain": [],
  "must_match": [
    "\\.github/workflows/.*\\.yml"
  ]
}
```

## Budget

0.02
