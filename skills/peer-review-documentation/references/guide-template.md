# Guide template for the Entur Developer Portal

This is the canonical template for writing how-to guides. Guide pages
walk a reader through accomplishing a concrete task -- an API
integration, a specific workflow, a piece of setup. Concept pages
explain *what something is*; reference pages document the full
surface. Guides focus on *doing*.

Text inside `[brackets]` is instructions -- replace with actual
content.

---

## Template

```markdown
---
title: How to [complete the sentence, keep it short and on point]
description: [One sentence summarising what this guide helps you accomplish]
owner: team-[github-team]
sidebar_label: [Optional shorter label if the title is long]
category: Guides
---

This guide shows you how to [complete the sentence, describing what
you can accomplish by using this guide]. Zudoku renders the
frontmatter `title` as the H1, so begin the body with prose.

Eager to start experimenting with code? Head over to our
[API documentation]([link to the relevant API reference]).

## Before you begin

To use this guide, you need to:

- Be an Entur partner
- Get your [authentication]([link to authentication guide])
- (Optional) [Write optional prerequisites like this]

## 1 [Header describing the step]

[Description of the step.]

### Example: [Header describing the topic of the example]

[One to two sentences describing what you can see or explore in the
example.]

[Include a code block with the example]

## (Optional) 2 [Header describing the step]

[Optional steps carry a number in the heading too, prefixed with
`(Optional)`.]

## 3 [Header describing the step]

[Description of the step.]

## Next steps

You made it! Here is what you can do next:

- [Bullets, written as complete sentences, containing links.]
- [Keep this list to five bullets or fewer.]

## Related topics

[Optional section. Include it when the reader has clear next places
to go.]

- [Bullets, written as complete sentences, containing links.]
- [Keep this list to five bullets or fewer.]

*Last updated on [date]*
```

---

## Writing conventions for guide pages

### What belongs here

- Numbered steps the reader follows to complete a concrete task
- Runnable code examples that illustrate each step
- Short prerequisites in "Before you begin"
- Pointers out to other pages in "Next steps" and "Related topics"

### What belongs elsewhere

- Definitions of the underlying concepts -- link to the concept page
- Exhaustive parameter / field / error listings -- link to the
  reference page
- Background motivation beyond a sentence or two

### Tone

- Direct, second-person ("You do X"). The reader is in the middle of
  a task, so lead with the action
- Keep examples concrete -- realistic payloads and scenarios do more
  work than abstract ones
- Link key concepts the first time they appear; let the concept page
  carry the definition

### Language and formatting

The same rules as the other page types apply:

- British English (organisation, colour, authorisation, behaviour)
- Oxford style guide for capitalisation -- only first word and
  proper nouns take a capital
- Frontmatter `title` renders as the H1; body starts at H2. Heading
  levels step by one. Headings end at the last meaningful word
- Active voice, present tense, one idea per sentence

### Sources

- Diataxis framework (How-to quadrant): <https://diataxis.fr/how-to-guides/>
- Oxford style guide: <https://www.ox.ac.uk/public-affairs/style-guide>
