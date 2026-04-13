# Scenario: CD Workflow with Image Promotion

## Description

Verifies cd.yaml implements the image promotion model: resolve-image job looks up the PR's git tag instead of rebuilding, deploy jobs chain through dev -> tst -> prd with correct concurrency (cancel-in-progress: false), and workflow_dispatch supports manual deployment with optional pre-built image.

## Prompt

You are setting up CI/CD workflows for a Kotlin project at Entur.

Details:

- Repository name: platform-examples
- Environments: dev, tst, prd

Read the Entur AI documentation in this repository (start with AGENTS.md, then read the setup-cicd-workflows skill) to answer.

Generate the complete `.github/workflows/cd.yaml` file. Output ONLY the YAML content.

## Assertions

```json
{
  "must_contain": [
    "name: Deploy",
    "workflow_dispatch:",
    "resolve-image:",
    "deploy-dev:",
    "deploy-tst:",
    "deploy-prd:",
    "cancel-in-progress: false",
    "entur/gha-helm/.github/workflows/deploy.yml@v1",
    "fetch-depth: 0",
    "fetch-tags: true",
    "gh pr list",
    "git tag -l",
    "paths-ignore:",
    "'terraform/**'"
  ],
  "must_match": [
    "resolve-image\\.outputs\\.image.*inputs\\.image.*build\\.outputs\\.image_and_tag",
    "always\\(\\).*!cancelled\\(\\).*!contains\\(needs",
    "inputs\\.image.*required.*false",
    "environment.*choice",
    "cd-deploy-dev",
    "cd-deploy-tst",
    "cd-deploy-prd",
    "needs.*build.*resolve-image",
    "BRANCH.*tr.*upper.*lower",
    "cd-\\$\\{\\{.*github\\.ref"
  ]
}
```

## Budget

0.15
