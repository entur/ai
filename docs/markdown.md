# Markdown Standards

All Entur docs must pass [markdownlint](https://github.com/DavidAnson/markdownlint).

## Linting

### Configuration

`.markdownlint-cli2.jsonc` at repository root:

```json
{
  "config": {
    "default": true,
    "MD013": false
  }
}
```

MD013 (line length) disabled -- tables, code blocks, and URLs frequently exceed 80 chars.

### Running Locally

```bash
npm install -g markdownlint-cli2
markdownlint-cli2 "**/*.md"
```

### Running in CI

Add markdownlint to CI or use a pre-commit hook. All markdown files must pass before merge.

## Rules Summary

Full reference: [markdownlint rules v0.40.0](https://github.com/DavidAnson/markdownlint/tree/v0.40.0/doc).

### Headings

- **MD001**: Heading levels increment by one (no skipping `#` to `###`)
- **MD003**: ATX-style headings (`# Heading`, not underlines)
- **MD018/MD019**: Exactly one space after `#`
- **MD022**: Blank lines above and below headings
- **MD023**: Headings start at column 1
- **MD024**: No duplicate heading text at same level
- **MD025**: One top-level `#` per file
- **MD026**: No trailing punctuation in headings
- **MD041**: First line must be a top-level heading

### Lists

- **MD004**: Consistent list markers (`-` preferred)
- **MD005**: Consistent indentation for same-level items
- **MD007**: Indent nested lists by 2 spaces
- **MD029**: Ordered lists use `1.` prefix (or sequential)
- **MD030**: One space after list markers
- **MD032**: Blank lines around lists

### Code Blocks

- **MD031**: Blank lines around fenced code blocks
- **MD040**: Fenced code blocks must specify a language (`text` for plain text)
- **MD046**: Use fenced (not indented) code blocks

### Whitespace

- **MD009**: No trailing spaces
- **MD010**: No hard tabs
- **MD012**: No consecutive blank lines

### Links and Images

- **MD011**: No reversed link syntax
- **MD034**: No bare URLs (use `<url>` or `[text](url)`)
- **MD039**: No spaces inside link text
- **MD042**: No empty links
- **MD045**: Images must have alt text

### Inline Formatting

- **MD037**: No spaces inside emphasis markers
- **MD038**: No spaces inside code span backticks

## Writing Guidelines

- Start every file with a single `#` heading
- Use heading hierarchy without skipping levels
- Separate all block elements (headings, lists, code blocks, tables) with blank lines
- End every file with a single newline
- Always specify the language tag on fenced code blocks:

````markdown
```java
public class Example {}
```
````

Use `text` for blocks with no specific language. Common tags: `yaml`, `json`, `hcl`, `kotlin`, `go`, `python`, `bash`, `dockerfile`, `xml`, `sql`.

- Use `-` for unordered lists, `1.` for ordered lists:

```markdown
- First item
- Second item
  - Nested item

1. Step one
2. Step two
```

- Align table columns:

```markdown
| Name   | Description     |
|--------|-----------------|
| `foo`  | Does foo things |
| `bar`  | Does bar things |
```

- Use relative paths for internal links, full URLs for external:

```markdown
[CONVENTIONS.md](../CONVENTIONS.md)
[java.md](java.md)
[entur/helm-charts](https://github.com/entur/helm-charts)
```
