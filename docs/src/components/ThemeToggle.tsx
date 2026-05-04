import { useEffect, useState } from 'react';
import { IconButton } from '@entur/button';
import { Tooltip } from '@entur/tooltip';
import { SunIcon, NightIcon } from '@entur/icons';

type Mode = 'light' | 'dark';

const STORAGE_KEY = 'entur-ai:color-mode';

function readInitial(): Mode {
  if (typeof window === 'undefined') return 'light';
  const stored = window.localStorage.getItem(STORAGE_KEY) as Mode | null;
  if (stored === 'light' || stored === 'dark') return stored;
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export function ThemeToggle() {
  const [mode, setMode] = useState<Mode>('light');

  useEffect(() => {
    const initial = readInitial();
    setMode(initial);
    document.documentElement.setAttribute('data-color-mode', initial);
  }, []);

  const toggle = () => {
    const next: Mode = mode === 'dark' ? 'light' : 'dark';
    setMode(next);
    document.documentElement.setAttribute('data-color-mode', next);
    window.localStorage.setItem(STORAGE_KEY, next);
  };

  const label = mode === 'light' ? 'Switch to dark mode' : 'Switch to light mode';

  return (
    <Tooltip content={label} placement="bottom">
      <IconButton aria-label={label} onClick={toggle} type="button">
        {mode === 'light' ? <SunIcon aria-hidden /> : <NightIcon aria-hidden />}
      </IconButton>
    </Tooltip>
  );
}

export default ThemeToggle;
