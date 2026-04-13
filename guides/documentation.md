# Documentation Standards

Rules for writing documentation in Entur repositories. Follow [markdown.md](markdown.md) for formatting.

## Methodology

- Plan outline before writing
- Iterate on outline and content
- Seek feedback through co-writing or review
- Review and update routinely

## Core Principles

- **"Why", not "what"** -- code shows what, docs explain why
- **Close to code** -- prefer Javadoc/KDoc/godoc/docstrings over separate files
- **Accurate** -- outdated docs are worse than none. Update docs with code changes
- **For the next developer** -- assume an Entur developer unfamiliar with this project

## Be Explicitly Purposeful

State up front in every document:

- **Target audience** -- who is this for?
- **Intent** -- what will the reader accomplish?
- **Scope** -- what is/isn't covered; link out for other topics
- **Prerequisites** -- required knowledge or setup, with links
- **Outcome-oriented headings** -- prefer: "Get started with the common Helm chart" over: "Copy `./helm` from helm-charts"
- **Further reading** -- list next documents at the end

## When to Write Documentation

Write or update when:

- Adding a public API, endpoint, class, or module
- Changing behavior described by existing docs
- Making an architectural decision (ADR in `doc/adr/`)
- Adding configuration options or environment variables
- Setting up non-obvious local development steps

Do **not** document:

- Self-explanatory private methods or trivial getters/setters
- What the code already says
- In separate files when an inline comment suffices

## Documentation Types

### Code Documentation

Use language-native formats: Javadoc (Java), KDoc (Kotlin), godoc (Go), docstrings (Python).

- Document all public classes, interfaces, methods, and functions
- Include `@param`/`@return`/`@throws` when not obvious from the name
- Document side effects, thread-safety, and performance characteristics
- Keep first sentence short -- it becomes the summary in generated docs

### README.md

Every repository root `README.md` must contain:

1. What the project does (one to two sentences)
2. How to run locally (prerequisites, commands)
3. How to deploy (or link to CD pipeline)
4. API overview (link to OpenAPI spec or endpoint summary)
5. Contact (team or Slack channel)

### API Documentation

- REST: OpenAPI spec in `specs/` or auto-generated from annotations
- gRPC: commented `.proto` files for every service, method, and message
- Include example requests/responses for non-trivial endpoints
- Document error responses and status codes

### Architecture Decision Records

AsciiDoc in `doc/adr/`. See [CONVENTIONS.md](../CONVENTIONS.md) for format.

### Configuration Documentation

Document every environment variable and config key:

```markdown
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | -- | Yes |
| `CACHE_TTL_SECONDS` | Cache time-to-live | `300` | No |
```

Place in README or `docs/configuration.md`.

## Writing Style

### Language and Tone

- English, plain and direct
- Active voice, present tense -- prefer: "The service validates the token." over: "The token is validated by the service."
- Strong verbs -- "generates", not "is responsible for generating"
- One sentence per thought
- Consistent abbreviations -- always `PaaS`, never `Paas` or `PAAS`

### Structure

- Follow [markdown.md](markdown.md)
- One `#` heading per file, no skipped levels
- Every heading followed by content before the next heading or list
- Every list preceded by an explanation
- Numbered lists only when order matters
- Homogeneous list entries -- same grammatical form
- Link rather than duplicate content

### Code Examples

- Runnable when possible
- Lead with correct usage; clearly label bad examples
- Minimal -- only what illustrates the point
- Always specify language tag on fenced code blocks

### Images and Diagrams

- Use only when text is insufficient
- Every image needs alt text and a caption
- Annotate relevant parts
- Charts must have correct axes, units, and labels

## File Organization

Per standard layout in [CONVENTIONS.md](../CONVENTIONS.md):

- `docs/` -- published documentation shared beyond the team
- `doc/adr/` -- Architecture Decision Records
- Inline comments -- implementation-level documentation

## Maintaining Documentation

- Review docs in every PR where code changes
- Delete docs for removed features
- Run `markdownlint-cli2 "**/*.md"` before committing (see [markdown.md](markdown.md))
- Treat doc lint failures the same as code lint failures

## Further Reading

- [Google developer documentation style guide](https://developers.google.com/style)
- [Google Technical Writing -- Active voice](https://developers.google.com/tech-writing/one/active-voice)
- [Microsoft Writing Style Guide](https://learn.microsoft.com/en-us/style-guide/welcome/)
