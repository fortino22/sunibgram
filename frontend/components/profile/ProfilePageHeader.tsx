'use client';

import { ArrowLeft } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';

interface ProfilePageHeaderProps {
  name: string;
  threadCount: number;
  isPremium?: boolean;
}

export function ProfilePageHeader({ name, threadCount, isPremium }: ProfilePageHeaderProps) {
  const router = useRouter();

  return (
    <header
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 20,
        background: 'var(--bg-blur)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid var(--border)',
        padding: '20px 16px',
        display: 'flex',
        alignItems: 'center',
        gap: 14,
      }}
    >
      <button
        type="button"
        aria-label="Go back"
        onClick={() => router.back()}
        className="btn-ghost"
        style={{
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
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          {name}
          {isPremium && <VerifiedBadge size={16} />}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
          {threadCount} thread{threadCount !== 1 ? 's' : ''}
        </div>
      </div>
    </header>
  );
}
