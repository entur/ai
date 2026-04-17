# Reference template for the Entur Developer Portal

This is the canonical template for writing reference pages. A
reference page is the authoritative, exhaustive description of an API
surface -- endpoints, fields, parameters, error codes, enums. Readers
arrive with a specific lookup in mind: "what does this field mean",
"what are the valid values", "what error codes can I get". Optimise
for scanning.

Text inside `[brackets]` is instructions -- replace with actual
content.

---

## Template

```markdown
---
title: [API or resource name] reference
description: [One sentence summarising what this reference covers]
owner: team-[github-team]
sidebar_label: [Optional shorter label]
category: Reference
---

[One to two sentences describing what this API or resource does and
when to use it. Link to the relevant how-to guide so readers who
picked the wrong page can find their way. Zudoku renders the
frontmatter `title` as the H1, so begin the body with prose.]

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/resource/{id}` | [One-line description] |
| POST | `/resource` | [One-line description] |

## [Resource or object name]

[One sentence describing the resource. Link to the concept page if one
exists, rather than redefining the domain model here.]

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | [Description. Put constraints (length, format, allowed values) in this column so the type column stays scannable.] |
| `name` | string | yes | [Description] |
| `status` | enum | no | [Description. List valid values below.] |

### Enum values

[Use when an enum has more than two values or each value needs explanation.]

- `ACTIVE` -- [Meaning and when this value is set]
- `INACTIVE` -- [Meaning and when this value is set]

### Example

[One minimal, correct example -- request and response for endpoints, or a
sample payload for resources. Use fenced code blocks with language tags.]

```json
{
  "id": "NSR:StopPlace:1234",
  "name": "Oslo S"
}
```

## Errors

| Code | HTTP status | Meaning | Action |
|------|-------------|---------|--------|
| `NOT_FOUND` | 404 | [When this is returned] | [What the caller should do] |
| `INVALID_ARGUMENT` | 400 | [When this is returned] | [What the caller should do] |

## Rate limits

[Optional. Include when rate limits are part of the contract the caller must
plan around. State the limits plainly -- requests per minute, burst allowance,
how the limit is communicated (header name, error code).]

## Related topics

- [Link to the how-to guide that uses this API]
- [Link to the concept page for the domain model]
- [Up to five bullets, written as complete sentences with links]

*Last updated on [date]*
```

---

## Writing conventions for reference pages

### What belongs here

- Exhaustive field, parameter, and error descriptions
- Authoritative types, formats, and constraints
- Minimal correct examples -- one per endpoint or resource
- Links out to guides (for tasks) and concepts (for meaning)

### What belongs elsewhere

- Step-by-step instructions -- link to the matching guide
- Background, motivation, or mental models -- link to the concept page
- Multiple variants of the same example -- show the canonical one and
  describe variants in prose

### Structure rules specific to reference pages

- Tables are the default for field and error lists. Readers scan them
- Every field has a description, even obvious ones. "Obvious" is
  relative
- Required/optional lives in its own dedicated column
- Constraints (length, format, allowed values) live in the
  description column; types stay short and scannable
- Every enum value with a non-obvious meaning gets its own bullet

### Tone

- Terse and factual. The reader is already motivated, so lead with
  the answer
- Use the imperative for "Action" columns: "Retry with exponential
  backoff" reads faster than "The caller should retry with
  exponential backoff"
- Skip marketing language and apology prefaces ("note that", "please
  be aware") -- state the fact and move on

### Language and formatting

The same rules as guides apply:

- British English (organisation, colour, authorisation, behaviour)
- Oxford style guide for capitalisation -- only first word and
  proper nouns take a capital
- Frontmatter `title` renders as the H1; body starts at H2. Heading
  levels step by one. Headings end at the last meaningful word
- Fenced code blocks with language tags

### Sources

- Diataxis framework (Reference quadrant): <https://diataxis.fr/reference/>
- Oxford style guide: <https://www.ox.ac.uk/public-affairs/style-guide>
