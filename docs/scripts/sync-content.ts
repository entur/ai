/**
 * Sync source skills/plugins/guides into Astro content collections.
 * Reads from sibling repo dirs, normalizes frontmatter (slug, title, category, tags).
 */
import { mkdir, readdir, readFile, rm, stat, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import matter from 'gray-matter';

const HERE = dirname(fileURLToPath(import.meta.url));
const DOCS_ROOT = resolve(HERE, '..');
const REPO_ROOT = resolve(DOCS_ROOT, '..');

const SOURCES = {
  skills: resolve(REPO_ROOT, 'skills'),
  plugins: resolve(REPO_ROOT, 'plugins'),
  guides: resolve(REPO_ROOT, 'guides'),
} as const;

const DEST_ROOT = resolve(DOCS_ROOT, 'src/content');

type Collection = keyof typeof SOURCES;

async function exists(path: string) {
  try {
    await stat(path);
    return true;
  } catch {
    return false;
  }
}

async function readDirEntries(path: string) {
  if (!(await exists(path))) return [];
  return readdir(path, { withFileTypes: true });
}

function slugify(input: string) {
  return input
    .toLowerCase()
    .replace(/\.(md|mdx)$/i, '')
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

const TERM_MAP: Record<string, string> = {
  api: 'API',
  apis: 'APIs',
  cicd: 'CI/CD',
  ci: 'CI',
  cd: 'CD',
  gcp: 'GCP',
  gke: 'GKE',
  iam: 'IAM',
  sql: 'SQL',
  nosql: 'NoSQL',
  http: 'HTTP',
  https: 'HTTPS',
  url: 'URL',
  urls: 'URLs',
  ui: 'UI',
  sdk: 'SDK',
  jwt: 'JWT',
  oauth: 'OAuth',
  rest: 'REST',
  grpc: 'gRPC',
  yaml: 'YAML',
  json: 'JSON',
  mdx: 'MDX',
  css: 'CSS',
  html: 'HTML',
  k8s: 'Kubernetes',
  scr: 'SCR',
};

function titleCase(input: string) {
  return input
    .replace(/[-_]+/g, ' ')
    .split(/\s+/)
    .filter(Boolean)
    .map((w) => TERM_MAP[w.toLowerCase()] ?? w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function normalize(
  collection: Collection,
  data: Record<string, unknown>,
  slug: string,
  fallbackTitle: string,
) {
  const out: Record<string, unknown> = { ...data };
  out.slug = slug;
  if (!out.title) out.title = typeof data.name === 'string' ? titleCase(data.name) : fallbackTitle;
  if (!out.category) out.category = collection;
  if (!Array.isArray(out.tags)) out.tags = [];
  return out;
}

async function syncFile(
  collection: Collection,
  srcAbs: string,
  destRel: string,
  slug: string,
  fallbackTitle: string,
) {
  const raw = await readFile(srcAbs, 'utf8');
  const parsed = matter(raw);
  const fm = normalize(collection, parsed.data, slug, fallbackTitle);
  const out = matter.stringify(parsed.content, fm);
  const destAbs = join(DEST_ROOT, collection, destRel);
  await mkdir(dirname(destAbs), { recursive: true });
  await writeFile(destAbs, out, 'utf8');
}

async function syncSkills() {
  const src = SOURCES.skills;
  const entries = await readDirEntries(src);
  for (const e of entries) {
    if (!e.isDirectory()) continue;
    const skillDir = join(src, e.name);
    const skillFile = join(skillDir, 'SKILL.md');
    if (!existsSync(skillFile)) continue;
    const slug = slugify(e.name);
    await syncFile('skills', skillFile, `${slug}.md`, slug, titleCase(e.name));
  }
}

async function syncPlugins() {
  const src = SOURCES.plugins;
  const entries = await readDirEntries(src);
  for (const e of entries) {
    if (!e.isDirectory()) continue;
    const slug = slugify(e.name);
    const skillsDir = join(src, e.name, 'skills');
    const skillEntries = await readDirEntries(skillsDir);
    const skillSlugs: string[] = [];
    for (const s of skillEntries) {
      const real = await stat(join(skillsDir, s.name)).catch(() => null);
      if (!real) continue;
      if (real.isDirectory()) skillSlugs.push(slugify(s.name));
    }
    const data = {
      slug,
      title: titleCase(e.name),
      category: 'plugins',
      tags: [] as string[],
      skills: skillSlugs,
    };
    const body = `# ${data.title}\n\nPlugin bundles the following skills:\n\n${skillSlugs
      .map((s) => `- [${s}](/skills/${s})`)
      .join('\n')}\n`;
    const out = matter.stringify(body, data);
    const destAbs = join(DEST_ROOT, 'plugins', `${slug}.md`);
    await mkdir(dirname(destAbs), { recursive: true });
    await writeFile(destAbs, out, 'utf8');
  }
}

async function syncGuides() {
  const src = SOURCES.guides;
  if (!(await exists(src))) return;
  const queue: string[] = [src];
  while (queue.length) {
    const dir = queue.shift()!;
    const entries = await readdir(dir, { withFileTypes: true });
    for (const e of entries) {
      const abs = join(dir, e.name);
      if (e.isDirectory()) {
        queue.push(abs);
        continue;
      }
      if (!/\.mdx?$/.test(e.name)) continue;
      const rel = relative(src, abs);
      const slugSegments = rel
        .replace(/\.(md|mdx)$/i, '')
        .split('/')
        .map(slugify);
      const slug = slugSegments.join('/');
      const fallbackTitle = titleCase(slugSegments[slugSegments.length - 1] ?? 'guide');
      const destRel = `${slug}.${e.name.endsWith('.mdx') ? 'mdx' : 'md'}`;
      await syncFile('guides', abs, destRel, slug, fallbackTitle);
    }
  }
}

async function syncAbout() {
  const src = resolve(REPO_ROOT, 'AGENTS.md');
  if (!(await exists(src))) return;
  const raw = await readFile(src, 'utf8');
  const parsed = matter(raw);
  const fm = {
    slug: 'about',
    title: 'About Entur AI',
    description: 'Entur AI agent instructions and platform overview',
    category: 'about',
    tags: [] as string[],
  };
  const out = matter.stringify(parsed.content, fm);
  const destAbs = join(DEST_ROOT, 'guides', 'about.md');
  await mkdir(dirname(destAbs), { recursive: true });
  await writeFile(destAbs, out, 'utf8');
}

async function main() {
  for (const c of Object.keys(SOURCES) as Collection[]) {
    const dest = join(DEST_ROOT, c);
    await rm(dest, { recursive: true, force: true });
    await mkdir(dest, { recursive: true });
  }
  await syncSkills();
  await syncPlugins();
  await syncGuides();
  await syncAbout();
  console.log('content sync complete');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
