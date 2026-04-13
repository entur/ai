# Scenario: Framing - Workflow File Extension (Positive)

## Description

Tests whether positive framing leads to clean, correct file extension usage without mentioning .yml.

## Prompt

You are given the following platform convention:

> ALWAYS use the .yaml extension for all GitHub Actions workflow files. Place workflow files in the .github/workflows/ directory with the .yaml extension.

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
