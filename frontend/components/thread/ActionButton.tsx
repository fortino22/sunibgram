'use client';

import { useState } from 'react';
import type { CSSProperties, MouseEvent, ReactNode } from 'react';

interface ActionButtonProps {
  'aria-label': string;
  onClick?: (e: MouseEvent<HTMLButtonElement>) => void;
  /** Identity color of this action — shown on hover and while active. */
  tint: string;
  /** Soft circular background shown behind the icon on hover. */
  softTint: string;
  /** Active state (liked / reposted / bookmarked) keeps the tint persistently. */
  active?: boolean;
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
    transition: 'background 0.12s ease, color 0.12s ease',
  };
}

export function ActionButton({
  'aria-label': label,
  onClick,
  tint,
  softTint,
  active = false,
  size = 'md',
  children,
}: ActionButtonProps) {
  const [hovered, setHovered] = useState(false);

  return (
    <button
      type="button"
      aria-label={label}
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        ...baseStyle(size),
        color: hovered || active ? tint : 'inherit',
        background: hovered ? softTint : 'transparent',
      }}
    >
      {children}
    </button>
  );
}
