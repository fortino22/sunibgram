'use client';

import { ArrowLeft } from 'lucide-react';
import { useRouter } from 'next/navigation';

interface ProfilePageHeaderProps {
  name: string;
  threadCount: number;
}

export function ProfilePageHeader({ name, threadCount }: ProfilePageHeaderProps) {
  const router = useRouter();

  return (
    <header
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 20,
        background: 'oklch(10.5% 0.005 260 / 0.92)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid var(--border)',
        padding: '10px 16px',
        display: 'flex',
        alignItems: 'center',
        gap: 14,
      }}
    >
      <button
        type="button"
        aria-label="Go back"
        onClick={() => router.back()}
        style={{
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          color: 'var(--text)',
          padding: 6,
          borderRadius: '50%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <ArrowLeft size={20} strokeWidth={1.8} />
      </button>
      <div>
        <div
          style={{
            fontWeight: 700,
            fontSize: 17,
            color: 'var(--text)',
            letterSpacing: '-0.02em',
            lineHeight: 1.2,
          }}
        >
          {name}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
          {threadCount} thread{threadCount !== 1 ? 's' : ''}
        </div>
      </div>
    </header>
  );
}
