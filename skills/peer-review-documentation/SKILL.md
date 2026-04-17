---
name: peer-review-documentation
description: >
  Peer-review and revise documentation for the Entur Developer Portal. Use when
  a developer shares a .md or .mdx draft and wants a senior peer to tighten it
  to Entur's writing standards -- how-to guides, concept pages, or reference
  pages. Trigger on phrases like "review this doc", "review this guide",
  "peer review", "edit this", "tighten this up", "make this production ready",
  "can you polish this documentation", or any message attaching a .md/.mdx
  file from the developer portal. Also trigger when the user mentions
  documentation review, developer portal review, technical writing review,
  or wants feedback before publishing a guide, concept, or reference page.
---

# Peer-review documentation for the Entur Developer Portal

Act as a senior technical writer reviewing a draft for the Entur Developer
Portal. Take the draft in, apply Entur's writing standards and the portal's
platform conventions, and return a polished document that is ready to
publish -- plus a short summary in chat explaining what you changed and why.

The workflow is one-way: a developer supplies an existing `.md` or
`.mdx` draft, and this skill peer-reviews and rewrites it. Producing
documentation from scratch is out of scope.

## Platform context: Zudoku

The Entur Developer Portal runs on [Zudoku](https://zudoku.dev). Pages are
`.md` (plain Markdown) or `.mdx` (Markdown plus React components). Two
rules shape every review:

1. **Prefer `.md`.** Plain Markdown covers everything most portal pages
   need -- headings, tables, fenced code blocks, admonitions. When a `.mdx`
   file uses only Markdown features, recommend renaming it to `.md`
2. **Use Zudoku's native syntax.** Write admonitions as plain Markdown
   (`:::note`, `:::warning`). Keep frontmatter inside the field list Zudoku
   publishes. The full catalogue lives in
   `references/zudoku-conventions.md` -- read it before Pass 4

Zudoku's own docs are available as plain Markdown by appending `.md` to any
URL (for example <https://zudoku.dev/docs/writing.md>). Prefer those when
linking from a review, so future reviewers can read the source directly.

## How Zudoku renders the title

Zudoku renders the frontmatter `title` as the page's H1. The body starts
at H2. A body `# My page title` produces a second H1 stacked beneath the
first -- the most common defect on the portal.

Always:

- Set `title` in frontmatter
- Start the body text directly, or at H2 if the first section has a
  heading

When you see a body H1 that duplicates the `title`, remove the body H1
and keep the frontmatter one.

## What good peer review looks like

A junior writer grows faster when a senior writer explains *why*
something changed, not just what changed. This skill produces both:

- A **revised document** (.md or .mdx) that follows Entur's standards and
  is ready to commit
- A **change summary in chat** -- grouped by category, with enough
  context that the author learns from the review rather than just
  accepting a rewrite

Be honest and specific. Aim for the form "Changed X to Y because Z":
"Changed 'is validated by the service' to 'the service validates' --
active voice reads faster and matches the Entur style guide" teaches the
author something. "Tightened several sentences" teaches them nothing.

## Step 1: Identify the document type

Every Entur Developer Portal page is one of three Diataxis types. The
type determines which template and conventions apply.

| Type | Purpose | Signals |
|------|---------|---------|
| **Guide** (how-to) | Help a reader accomplish a task | Title starts with "How to", numbered steps, "Before you begin" |
| **Concept** (explanation) | Help a reader understand the domain | Defines a term or model, explains relationships, explanatory prose throughout |
| **Reference** | Authoritative lookup for an API | Field tables, endpoint lists, enum values, error codes |

When the draft is ambiguous, tell the author which type you *think* it
is and why, then proceed. When it looks like a mix (a concept page with
a how-to buried inside), call that out in the summary -- this is a
common defect and the fix is usually to split the document.

Read the template for the detected type before reviewing:

- Guides: `references/guide-template.md`
- Concepts: `references/concept-template.md`
- Reference: `references/reference-template.md`

## Step 2: Review in passes

Apply the passes one at a time. Read the document straight through
first without touching it. Then run each pass in order. Each pass has a
narrow focus, and the order matters -- fixing structure first keeps you
from polishing sentences you are about to delete.

### Pass 1: Type and structure

- Does the document match its Diataxis type? A guide is task-oriented,
  a concept explanatory, a reference exhaustive. Mismatches are the
  most expensive defect to fix later -- flag them
- Does the structure match the template for that type? Check for
  "Before you begin", correct heading hierarchy, "Next steps",
  `*Last updated on ...*`
- Is anything in the wrong document? Concept material inside a guide,
  a how-to buried inside a reference page, motivation written into a
  reference table. Move such content to the right page or flag it

### Pass 2: Content and accuracy

- Is each step or section doing the job its heading promises?
- Are claims verifiable? When a draft says "the API returns X", check
  whether that is testable from what is shown. When the claim is hard
  to verify from the page alone, leave a `<!-- REVIEWER: ... -->`
  comment and let the author confirm
- Are concepts linked to their concept pages? Guides and references
  should link out; concept pages should link to one another
- Are code examples runnable, minimal, and labelled with a language
  tag?
- For references: does every field have a description, are
  required/optional columns consistent, and do enum values carry
  meaning alongside their names?

### Pass 3: Language and tone

Apply the Entur writing standards. The most common issues to look for:

- **Passive voice** -- "The token is validated by the service" becomes
  "The service validates the token". Keep passive only when the actor
  is genuinely irrelevant
- **Weak verbs** -- "is responsible for generating" becomes "generates"
- **Hedging** -- "might potentially be able to" becomes "can"
- **Present tense for system behaviour** -- "The endpoint will return
  200" becomes "The endpoint returns 200"
- **One idea per sentence** -- split run-on sentences
- **British English** -- "organization" -> "organisation", "color" ->
  "colour", "authorization" -> "authorisation", "behavior" ->
  "behaviour"
- **Consistent abbreviations** -- use `PaaS` everywhere; normalise
  `Paas` and `PAAS` to match
- **Link concepts instead of redefining them** -- the concept page
  carries the definition

### Pass 4: Formatting, mechanics, and platform

Read `references/zudoku-conventions.md` before this pass -- the
Zudoku-specific rules live there and drift as the platform evolves.

General Markdown mechanics:

- **Frontmatter `title` is the page H1** -- start the body at H2 (or
  at prose, with no heading at all). Remove any body H1 that duplicates
  the title
- **Oxford heading capitalisation** -- only the first word and proper
  nouns carry a capital. "New approaches in mobility research" reads
  Oxford-correct
- **Heading levels step by one** -- every H3 sits under an H2, every
  H4 sits under an H3
- **Headings read as titles** -- end at the last meaningful word,
  the way a newspaper headline ends
- **ATX-style headings** (`# Heading`)
- **Fenced code blocks with language tags** -- use ` ```bash `,
  ` ```json `, ` ```mermaid ` and the like
- **Blank lines** above and below headings, lists, and code blocks
- **`-` for unordered lists**
- **Relative paths for internal links, full URLs for external**
- **`*Last updated on <date>*`** at the end, in italic

Zudoku-specific checks (see the reference file for full details):

- **File extension** -- when the `.mdx` page uses a Zudoku component,
  keep `.mdx`. When it uses only Markdown features, recommend renaming
  to `.md`
- **Required frontmatter fields** -- every page carries `title`,
  `description`, and `owner`. Treat a missing or malformed required
  field as a blocking issue (the same severity as a broken link).
  `owner` is a real GitHub team slug starting with `team-` (for
  example `team-api`). When the team is unknown, leave
  `<!-- REVIEWER: confirm GitHub team -->` and let the author fill
  it in
- **Frontmatter conventions** -- `description` reads as a complete
  sentence. `draft: true` stays only when the author intends it
- **Admonitions** -- rewrite prose like "Note: ..." or "Important:
  ..." as `:::note` / `:::warning` blocks with blank lines around the
  directives. Flip `<Callout>` to admonitions; keep the component only
  when it renders something that goes beyond admonitions
- **Diagrams** -- static images, ASCII art, and multi-step prose flows
  are candidates to become Mermaid fenced code blocks. Mermaid is
  encouraged because diagrams stay versioned and diffable alongside
  the text. Flip `<Mermaid chart="...">` to a fenced code block; keep
  the component when the page genuinely needs client-side rendering
- **Components in `.mdx`** -- each imported component earns its keep.
  Replace `<Stepper>` with numbered H2 headings in guides. Replace
  `<Button>` with a plain Markdown link, unless a page truly needs a
  CTA. `<Playground>` on a reference page earns its place -- keep it

### Pass 5: Flow and concision

Once the above is clean, smooth the reading experience:

- Cut sentences that repeat the previous sentence in different words
- Cut hedges and throat-clearing: "It should be noted that", "Please
  be aware that", "Simply", "Just"
- Break paragraphs longer than five sentences into shorter ones
- Check transitions between steps -- the reader should understand
  *why* they are moving to the next step

## Step 3: Apply revisions

Edit the document in place. Preserve the author's voice where it
already works -- peer review keeps the author's voice, making it
clearer rather than replacing it.

Two rules for revision:

1. **When intent is unclear, leave a comment.** When the author's
   intention is ambiguous, a paragraph makes a factual claim that is
   hard to verify, or a sentence could mean two different things,
   insert `<!-- REVIEWER: ... -->` and let the author resolve it.
   HTML comments live in the body, after any closing `---` of the
   frontmatter
2. **Keep structural changes visible in the summary.** Moving a
   section, splitting a step, merging two sections -- each of these
   deserves an explicit mention. Language polish can be described in
   aggregate

## Step 4: Summarise the changes in chat

After saving the revised file, post the summary in chat. The author
reads it once, in context, so chat is the right home for it. Use this
structure:

```
## Peer review: <document-name>

**Type detected:** <guide | concept | reference>  -- <one-sentence reason>

**Structural changes**
- <change 1 -- what moved, split, or merged, and why>
- <change 2>

**Content and accuracy**
- <change or flagged issue 1>
- <change or flagged issue 2>

**Language and tone**
- <aggregate note, e.g. "Converted ~8 passive constructions to active voice">
- <specific call-out for changes the author should understand>

**Formatting and mechanics**
- <aggregate note, e.g. "Normalised heading capitalisation to Oxford style
  (6 headings)">

**Questions for the author**
- <anything left as a REVIEWER comment in the document>

**Verdict:** <ready to publish | ready after author resolves questions |
needs another pass -- with the blocker>
```

Evaluation harnesses that require a file can save this same text as
`summary.md` or `review.md` alongside the revised document; the
content is identical either way.

Keep the summary tight. One to three bullets per section is plenty
for most documents. When a section has nothing to report, say "No
changes" so the author sees what was checked.

## Step 5: Review checklist before returning

Before handing back:

- [ ] Document type identified and stated
- [ ] Structure matches the template for that type
- [ ] Every revision is either applied or left as a `REVIEWER:` comment
- [ ] Summary in chat aligns with the revisions in the document
- [ ] Author's voice preserved where it was already working
- [ ] `*Last updated on <date>*` present in italic at the end
- [ ] File saved with the chosen extension (.md or .mdx, per Pass 4)
      and the same base filename

## On tone

You are a senior peer. Be direct about problems, generous about
strengths. Peer review helps the author grow, so the summary sometimes
notes good choices alongside the changes -- "Kept the concrete example
in step 2, which does more work than three paragraphs could."
