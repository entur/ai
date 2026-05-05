import { useEffect, useMemo, useState } from 'react';

export interface TocHeading {
  depth: number;
  slug: string;
  text: string;
}

interface Props {
  headings: TocHeading[];
  minDepth?: number;
  maxDepth?: number;
}

export function TableOfContents({ headings, minDepth = 2, maxDepth = 3 }: Props) {
  const filtered = useMemo(
    () => headings.filter((h) => h.depth >= minDepth && h.depth <= maxDepth),
    [headings, minDepth, maxDepth],
  );

  const [active, setActive] = useState<string | null>(filtered[0]?.slug ?? null);

  useEffect(() => {
    if (filtered.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.find((e) => e.isIntersecting);
        if (visible) setActive(visible.target.id);
      },
      { rootMargin: '0px 0px -50% 0px', threshold: [0.1, 0.5, 1.0] },
    );

    const els = filtered
      .map((h) => document.getElementById(h.slug))
      .filter((el): el is HTMLElement => el !== null);
    els.forEach((el) => observer.observe(el));

    return () => {
      els.forEach((el) => observer.unobserve(el));
      observer.disconnect();
    };
  }, [filtered]);

  if (filtered.length < 2) return null;

  return (
    <nav className="toc" aria-label="On this page">
      <h2 className="toc__title">On this page</h2>
      <ul className="toc__list">
        {filtered.map((h) => (
          <li
            key={h.slug}
            className={`toc__item toc__item--depth-${h.depth}`}
          >
            <a
              className={`toc__link${active === h.slug ? ' toc__link--active' : ''}`}
              href={`#${h.slug}`}
            >
              <span>{h.text}</span>
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}

export default TableOfContents;
