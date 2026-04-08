---
name: write-api-documentation
description: >
  Write API guides for the Entur Developer Portal following Entur's guide template
  and writing standards. Use this skill when a developer wants to write, draft, or
  create a guide or how-to for the Developer Portal, document an API for external
  users, write documentation for partners, or create step-by-step instructions for
  using an Entur API or service. Also trigger when someone mentions "developer portal
  guide", "API guide", "how-to guide", "partner documentation", or wants help
  structuring documentation for external consumption on developer.entur.no.
---

# Write API Documentation for the Entur Developer Portal

Help developers write high-quality guides for the Entur Developer Portal. Guides
are practical, step-by-step documents that help external users (partners, developers)
solve concrete problems or achieve goals using Entur APIs and services.

A guide is not a concept explanation, not an internal conventions document, and not
a reference page. It is a "how-to" focused on *doing*, aimed at users who already
have some context and need guidance to get the job done.

## Before writing anything

Read `references/guide-template.md` in this skill directory to understand the
full template and writing conventions. The template is based on Entur's official
writing guide and must be followed closely.

## Step 1: Understand what the developer wants to document

Ask the developer:

| Question | Why it matters |
|----------|----------------|
| **What API or service is this guide about?** | Determines scope and technical content |
| **What task will the reader accomplish?** | Shapes the title ("How to ...") and step structure |
| **Who is the reader?** | Usually an Entur partner or external developer, but confirm |
| **What prerequisites does the reader need?** | Authentication, partner status, specific API access |
| **How many steps are involved?** | Helps plan the guide structure |
| **Are there code examples to include?** | Guides benefit greatly from runnable examples |

If the developer is unsure whether they need a guide or something else (a concept
page, a reference page, internal docs), help them figure it out. A guide answers
"How do I do X?" If their content answers "What is X?" it belongs under Concepts
in the Developer Portal instead.

## Step 2: Draft the guide structure

Before writing content, produce an outline following this skeleton:

```text
# How to [verb] [object]                          <- H1, one per guide
  Intro paragraph                                  <- What the guide helps you do
  Link to API documentation                        <- "Eager to start experimenting..."

## Before you begin                                <- Prerequisites
  - Bullet list of requirements

## 1 [Step heading]                                <- Numbered steps, H2
  Description
  ### Example: [topic]                             <- Optional code example, H3

## 2 [Step heading]                                <- (Optional) prefix if step is optional

## 3 [Step heading]

## Next steps                                      <- What to do after completing the guide
  - Up to 5 bullets with links

## Related topics                                  <- Optional
  - Up to 5 bullets with links

Last updated on [date]                             <- Italic
```

Share the outline with the developer and confirm before writing full content.

## Step 3: Write the guide

Follow these rules from Entur's writing standards:

### Language and tone

- Write in **British English** following the Oxford style guide
- Use plain, direct language
- Use active voice, present tense: "The service validates the token" not "The token
  is validated by the service"
- Use strong verbs: "generates" not "is responsible for generating"
- One sentence per thought
- Be consistent with abbreviations: always `PaaS`, never `Paas` or `PAAS`

### Headings

- Follow Oxford capitalisation: only the first word and proper nouns are capitalised.
  Write "New approaches in mobility research", not "New Approaches in Mobility Research"
- Headings never end with a period
- Numbered steps use the format: `## 1 Heading describing the step`
- Optional steps use the format: `## (Optional) 2 Heading describing the step`
- Example sub-sections use H3: `### Example: Topic of the example`

### Structure rules

- The guide has exactly one H1 heading, which is the title
- Steps are numbered in the headings (1, 2, 3...)
- Each step heading describes what the step accomplishes
- Every step has descriptive content below the heading before any sub-sections
- Examples follow their parent step as H3 sub-sections with 1-2 sentences of context

### Content in specific sections

**Title (H1):** Always starts with "How to". Keep it short and on point.

**Intro paragraph:** One sentence completing "This guide shows you how to ..."

**Link to API docs:** Include the standard line: "Eager to start experimenting with
code? Head over to our [API documentation](link)."

**Before you begin:** Bullet list of prerequisites. Common ones include:

- Be an Entur partner
- Get your authentication
- Mark optional prerequisites with "(Optional)" prefix

**Numbered steps:** The core of the guide. Each step:

- Has a descriptive heading
- Explains what to do and why
- Includes code examples where relevant (as H3 sub-sections)
- References concepts via links rather than explaining them inline

**Next steps:** Up to 5 bullets, written as complete sentences with links,
describing what the reader can do after completing the guide.

**Related topics:** Optional. Up to 5 bullets with links to related guides
or concept pages.

**Last updated:** The current date in italic at the very end.

### Concepts and terminology

Do not explain concepts in detail within the guide. Concept descriptions belong
under the Concepts section of the Developer Portal. When you reference a key
concept, link to its concept page so readers can learn more if needed.

### Code examples

- Make examples runnable when possible
- Lead with correct usage; clearly label bad examples
- Keep examples minimal -- only what illustrates the point
- Always specify a language tag on fenced code blocks
- Include both the request and response for API calls

## Step 4: Format as Markdown

The guide will be published via GitHub as a Markdown file. Follow these formatting
rules (which align with the repo's markdownlint configuration):

- ATX-style headings (`# Heading`)
- Heading levels increment by one, no skipping
- Blank lines above and below headings, lists, and code blocks
- Use `-` for unordered lists
- Fenced code blocks with language tags (never indented code blocks)
- No trailing spaces or hard tabs
- End the file with a single newline
- Use relative paths for internal links, full URLs for external

For the complete markdown rules, read `guides/markdown.md` in the entur/ai repo.

## Step 5: Save and present the guide

Save the finished guide as a `.md` file. Use a descriptive filename in kebab-case
that matches the guide topic (e.g., `how-to-search-for-stop-places.md`,
`how-to-authenticate-with-entur.md`).

Ask the developer where the file should be saved -- typically in the repository
that will publish it to the Developer Portal.

## Step 6: Review checklist

Before delivering the guide, verify:

- [ ] Title starts with "How to" and is short and specific
- [ ] Intro paragraph completes "This guide shows you how to ..."
- [ ] "Before you begin" section lists all prerequisites
- [ ] Steps are numbered and have descriptive headings
- [ ] Optional steps are marked with "(Optional)" prefix
- [ ] Code examples have language tags and are runnable
- [ ] Concepts are linked, not explained inline
- [ ] "Next steps" has up to 5 bullets with links
- [ ] British English spelling throughout (organisation, colour, authorisation)
- [ ] Oxford-style heading capitalisation (only first word + proper nouns)
- [ ] Headings do not end with periods
- [ ] Markdown passes linting (no skipped heading levels, blank lines around blocks)
- [ ] "Last updated on [date]" in italic at the end
