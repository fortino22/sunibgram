'use client';

import { useRouter } from 'next/navigation';
import type { TrendingHashtag } from '@/lib/services/explore';

interface TrendingSectionProps {
  hashtags: TrendingHashtag[];
}

export function TrendingSection({ hashtags }: TrendingSectionProps) {
  const router = useRouter();

  if (hashtags.length === 0) return null;

  return (
    <div>
      <div
        style={{
          padding: '16px 20px 8px',
          fontWeight: 700,
          fontSize: 20,
          color: 'var(--text)',
          letterSpacing: '-0.02em',
        }}
      >
        Trending
      </div>
      {hashtags.map((tag, idx) => (
        <button
          key={tag.name}
          onClick={() => router.push(`/explore?q=%23${tag.name}`)}
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 2,
            width: '100%',
            padding: '14px 20px',
            background: 'transparent',
            border: 'none',
            borderBottom: idx < hashtags.length - 1 ? '1px solid var(--border)' : 'none',
            cursor: 'pointer',
            textAlign: 'left',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--bg-subtle)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent';
          }}
        >
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Trending</span>
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)' }}>#{tag.name}</span>
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
            {tag.threadCount.toLocaleString()} posts
          </span>
        </button>
      ))}
    </div>
  );
}
