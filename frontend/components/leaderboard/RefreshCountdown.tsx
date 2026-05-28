'use client';

import { useEffect, useState } from 'react';

interface RefreshCountdownProps {
  refreshAt: string | undefined;
}

function format(seconds: number): string {
  if (seconds <= 0) return 'Refreshing…';
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `Refreshes in ${m}:${s.toString().padStart(2, '0')}`;
}

export function RefreshCountdown({ refreshAt }: RefreshCountdownProps) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const secondsLeft = (() => {
    if (!refreshAt) return null;
    const target = Date.parse(refreshAt);
    if (Number.isNaN(target)) return null;
    return Math.max(0, Math.floor((target - now) / 1000));
  })();

  if (secondsLeft === null) return null;

  return (
    <span
      style={{
        padding: '6px 12px',
        borderRadius: 9999,
        background: 'var(--bg-subtle)',
        color: 'var(--text)',
        fontSize: 12,
        fontWeight: 600,
        letterSpacing: '0.02em',
        fontVariantNumeric: 'tabular-nums',
        border: '1px solid var(--border)',
      }}
      title={refreshAt}
    >
      {format(secondsLeft)}
    </span>
  );
}

