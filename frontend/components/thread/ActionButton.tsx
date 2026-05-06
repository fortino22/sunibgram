'use client';

import type { CSSProperties, MouseEvent, ReactNode } from 'react';

interface ActionButtonProps {
  'aria-label': string;
  onClick?: (e: MouseEvent<HTMLButtonElement>) => void;
  color?: string;
  size?: 'sm' | 'md';
  children: ReactNode;
}

function baseStyle(size: 'sm' | 'md'): CSSProperties {
  return {
    border: 0,
    borderRadius: 999,
    background: 'transparent',
    color: 'inherit',
    cursor: 'pointer',
    display: 'inline-flex',
    alignItems: 'center',
    gap: 6,
    minWidth: size === 'sm' ? 36 : 44,
    height: size === 'sm' ? 30 : 32,
    padding: size === 'sm' ? '0 6px' : '0 8px',
    font: 'inherit',
  };
}

function setHoverBg(e: MouseEvent<HTMLButtonElement>, on: boolean) {
  e.currentTarget.style.background = on ? 'var(--bg-subtle)' : 'transparent';
}

export function ActionButton({
  'aria-label': label,
  onClick,
  color,
  size = 'md',
  children,
}: ActionButtonProps) {
  return (
    <button
      type="button"
      aria-label={label}
      onClick={onClick}
      onMouseEnter={(e) => setHoverBg(e, true)}
      onMouseLeave={(e) => setHoverBg(e, false)}
      style={{ ...baseStyle(size), ...(color !== undefined ? { color } : {}) }}
    >
      {children}
    </button>
  );
}
