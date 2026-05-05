import React, { useEffect, useRef, useState } from 'react';
import { SideNavigation, SideNavigationItem, SideNavigationGroup } from '@entur/menu';
import { TextField } from '@entur/form';
import { SearchIcon } from '@entur/icons';

export type NavItem = { label: string; href: string };
export type NavGroup = { title: string; items: NavItem[] };

type Props = { groups: NavGroup[] };

export function SidebarFilter({ groups }: Props) {
  const [query, setQuery] = useState('');
  const [pathname, setPathname] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setPathname(window.location.pathname.replace(/\/$/, ''));
  }, []);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  const isActive = (href: string) => pathname === (href.replace(/\/$/, '') || '/');

  const filtered = query.trim()
    ? groups
        .map((g) => ({ ...g, items: g.items.filter((i) => i.label.toLowerCase().includes(query.toLowerCase())) }))
        .filter((g) => g.items.length > 0)
    : groups;

  const isMac = typeof navigator !== 'undefined' && /Mac/.test(navigator.platform);

  return (
    <>
      <div className="sidebar-search">
        <TextField
          label="Filter"
          size="medium"
          prepend={<SearchIcon />}
          append={
            !query ? (
              <span className="sidebar-search__kbd" aria-hidden="true">
                <kbd>{isMac ? '⌘' : 'Ctrl'}</kbd><kbd>K</kbd>
              </span>
            ) : undefined
          }
          value={query}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setQuery(e.target.value)}
          clearable
          onClear={() => setQuery('')}
          ref={inputRef}
        />
      </div>
      <SideNavigation>
        {filtered.map((group) => (
          <SideNavigationGroup key={group.title} title={group.title} defaultOpen>
            {group.items.map((item) => (
              <SideNavigationItem key={item.href} as="a" href={item.href} active={isActive(item.href)}>
                {item.label}
              </SideNavigationItem>
            ))}
          </SideNavigationGroup>
        ))}
      </SideNavigation>
    </>
  );
}
