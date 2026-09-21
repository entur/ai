# Scenario: Critical Rule 6 - Pin All Dependencies

## Description

Tests that the agent applies the pinning requirement across all three dependency types at once (Terraform module refs, GitHub Actions versions, Docker base image tags) and flags each unpinned example as wrong, not just one of the three.

## Prompt

You are reviewing a colleague's draft PR. It contains these three snippets:

```hcl
module "postgresql" {
  source = "github.com/entur/terraform-google-sql-db//modules/postgresql"
}
```

```yaml
- uses: actions/checkout@main
```

```dockerfile
FROM gcr.io/distroless/java25-debian13:latest
```

Read the Entur AI documentation in this repository (start with AGENTS.md) and list every pinning problem in the three snippets above, with a corrected version of each.

## Assertions

```json
{
  "must_contain": [
    "ref=",
    "checkout@v",
    "nonroot"
  ],
  "must_not_contain": [
    "these are all fine",
    "no problems",
    "no issues found"
  ],
  "must_match": [
    "\\?ref=v?[0-9]",
    "checkout@v[0-9]",
    "latest.*(pin|specific|avoid|not|don't|do not)|(pin|specific|avoid|not|don't|do not).*latest"
  ]
}
```

## Budget

0.10
