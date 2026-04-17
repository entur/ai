# Concept template for the Entur Developer Portal

This is the canonical template for writing concept pages. Concept
pages explain *what something is* and *why it matters* -- they help
readers understand the domain, terminology, and models behind the
APIs. Step-by-step instructions belong in a guide; exhaustive field
references belong in a reference page.

Text inside `[brackets]` is instructions -- replace with actual
content.

---

## Template

```markdown
---
title: [Concept name]
description: [One sentence summarising the concept, used for SEO and listings]
owner: team-[github-team]
sidebar_label: [Optional shorter label if the title is long]
category: Concepts
---

[One or two sentences that define the concept in plain language. Start
with the concept name and complete the sentence "A [concept] is ..."
or "[Concept] is ...". Zudoku renders the frontmatter `title` as the
H1, so begin the body with prose.]

[Optional second paragraph explaining why this concept exists or what problem
it solves -- the motivation the reader needs to care.]

## Overview

[Two to four paragraphs giving the reader a working mental model. Focus on
relationships between this concept and others the reader already knows.
Link to related concept pages rather than redefining terms inline.]

## Key properties

[Optional section. Use when the concept has a small set of defining
attributes the reader must internalise. Prefer a bullet list over a table
unless you genuinely need two columns.]

- **[Property name]** -- [one-sentence description]
- **[Property name]** -- [one-sentence description]

## How it relates to [adjacent concept]

[Optional. Use when readers routinely confuse this concept with another, or
when the relationship is load-bearing for later guides. Keep it short and
link to the adjacent concept page.]

## Example

[Optional. A short illustrative example -- often a realistic payload
or a concrete scenario -- that grounds the abstract description.
Keep it short; the step-by-step walkthrough belongs in a guide.]

## Next steps

- [Link to the most relevant how-to guide that uses this concept]
- [Link to the reference page that documents the API surface for this concept]
- [Up to five bullets, written as complete sentences with links]

*Last updated on [date]*
```

---

## Writing conventions for concept pages

### What belongs here

- Definitions, mental models, and terminology
- Relationships between concepts
- Motivation ("why this exists")
- Short, illustrative examples that ground the abstraction

### What belongs elsewhere

- Step-by-step instructions -- link to the matching guide
- Exhaustive field or parameter lists -- link to the reference page
- Troubleshooting recipes -- link to a guide

### Tone

- Present tense throughout. Concepts describe how the system *is*
  today; leave history to commit messages
- Prefer concrete nouns over abstractions. "A Stop Place groups
  quays that travellers use for the same transport purpose" beats
  "Stop Place is an aggregation entity"
- When two terms are genuinely synonymous, pick one and use it
  everywhere. Call out the synonym once so readers can connect
  external material

### Language and formatting

The same rules as guides apply:

- British English (organisation, colour, authorisation, behaviour)
- Oxford style guide for capitalisation -- only first word and
  proper nouns take a capital
- Frontmatter `title` renders as the H1; body starts at H2. Heading
  levels step by one. Headings end at the last meaningful word
- Active voice, present tense, one idea per sentence

### Sources

- Diataxis framework (Explanation quadrant): <https://diataxis.fr/explanation/>
- Oxford style guide: <https://www.ox.ac.uk/public-affairs/style-guide>
