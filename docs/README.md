# Entur AI Docs

> **Audience:** Entur employees contributing to the generated documentation site.
> **AI agents:** Stop. Read [AGENTS.md](../AGENTS.md) instead of this file.

This is the source for the documentation site published at [ki.entur.no](https://ki.entur.no). It is an [Astro](https://astro.build) static site that displays the guides, skills, and plugin documentation maintained in this repository.

## How content gets here

The site does not store its own copy of the documentation. Instead, a sync script reads source files from the parent repository and writes them into `src/content/` as Astro content collections.

| Source | Destination | What gets synced |
|--------|-------------|-----------------|
| `../guides/` | `src/content/guides/` | All `.md` and `.mdx` files, preserving subdirectory structure |
| `../skills/` | `src/content/skills/` | `SKILL.md` from each skill directory |
| `../plugins/` | `src/content/plugins/` | One entry per plugin, listing its bundled skills |
| `../AGENTS.md` | `src/content/guides/about.md` | The top-level agent instructions |

The sync script (`scripts/sync-content.ts`) normalizes frontmatter (adds `slug`, `title`, `category`, `tags`) and rewrites it in place. It runs automatically before `dev` and `build` via `pre*` hooks.

`src/content/` is generated output. Do not edit files there directly -- changes will be overwritten on the next sync.

## Running locally

```bash
cd docs
npm install
npm run dev
```

The dev server runs at `http://localhost:4321`. Content is synced from the parent repo on startup.

To sync content without starting the server:

```bash
npm run sync-content
```

## Building

```bash
npm run build
npm run preview
```

Output goes to `dist/`.

## Deployment

The site deploys to GitHub Pages on every push to `main` that touches `docs/**` or `guides/**`, via `.github/workflows/deploy-docs.yml`. It is served at `ki.entur.no` using the `CNAME` file in `public/`.

## Structure

```text
docs/
  public/          Static assets and CNAME
  scripts/         sync-content.ts
  src/
    components/    Astro and React components
    content/       Generated -- do not edit
    layouts/       Page layouts
    lib/           Utilities
    pages/         Astro pages and routing
    styles/        Global CSS
  astro.config.mjs
  package.json
  tsconfig.json
```
