import { useEffect, useState } from 'react';
import { SideNavigation, SideNavigationItem } from '@entur/menu';

export type SideNavGroup = {
  title: string;
  items: { label: string; href: string }[];
};

export type DocsSideNavProps = {
  groups: SideNavGroup[];
  basePath: string;
};

function withBase(basePath: string, href: string) {
  if (!basePath || basePath === '/') return href;
  const trimmed = basePath.endsWith('/') ? basePath.slice(0, -1) : basePath;
  return `${trimmed}${href.startsWith('/') ? href : `/${href}`}`;
}

export function DocsSideNav({ groups, basePath }: DocsSideNavProps) {
  const [pathname, setPathname] = useState<string>('');
  useEffect(() => setPathname(window.location.pathname.replace(/\/$/, '')), []);

  return (
    <SideNavigation>
      {groups.flatMap((group) => [
        <div key={`h-${group.title}`} style={{ padding: '12px 16px 4px', fontWeight: 600, fontSize: 12, textTransform: 'uppercase', letterSpacing: 0.5, opacity: 0.7 }}>
          {group.title}
        </div>,
        ...group.items.map((item) => {
          const fullHref = withBase(basePath, item.href).replace(/\/$/, '');
          const active = pathname === fullHref;
          return (
            <SideNavigationItem
              key={item.href}
              as="a"
              href={withBase(basePath, item.href)}
              active={active}
            >
              {item.label}
            </SideNavigationItem>
          );
        }),
      ])}
    </SideNavigation>
  );
}

export default DocsSideNav;
