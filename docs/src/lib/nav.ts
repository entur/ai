import { getCollection } from 'astro:content';
import type { CollectionEntry } from 'astro:content';
import type { SideNavGroup } from '../components/DocsSideNav.tsx';

export async function buildSkillsNav(): Promise<SideNavGroup[]> {
  const entries = await getCollection('skills');
  return [
    {
      title: 'Skills',
      items: [
        { label: 'All skills', href: '/skills' },
        ...sorted(entries).map((e) => ({
          label: e.data.title,
          href: `/skills/${e.id}`,
        })),
      ],
    },
  ];
}

export async function buildPluginsNav(): Promise<SideNavGroup[]> {
  const entries = await getCollection('plugins');
  return [
    {
      title: 'Plugins',
      items: [
        { label: 'All plugins', href: '/plugins' },
        ...sorted(entries).map((e) => ({
          label: e.data.title,
          href: `/plugins/${e.id}`,
        })),
      ],
    },
  ];
}

export async function buildGuidesNav(): Promise<SideNavGroup[]> {
  const entries = await getCollection('guides');
  const groups = new Map<string, { label: string; href: string }[]>();
  for (const e of sorted(entries)) {
    const segments = e.id.split('/');
    const top = segments.length > 1 ? segments[0] : 'general';
    const groupTitle = top
      .split('-')
      .map((w: string) => w[0]?.toUpperCase() + w.slice(1))
      .join(' ');
    const arr = groups.get(groupTitle) ?? [];
    arr.push({ label: e.data.title, href: `/guides/${e.id}` });
    groups.set(groupTitle, arr);
  }
  return [...groups.entries()].map(([title, items]) => ({ title, items }));
}

function sorted<T extends CollectionEntry<'skills' | 'plugins' | 'guides'>>(entries: T[]): T[] {
  return [...entries].sort((a, b) => a.data.title.localeCompare(b.data.title));
}
