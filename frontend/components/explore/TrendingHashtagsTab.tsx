'use client';

import { useRouter } from 'next/navigation';
import { MoreHorizontal } from 'lucide-react';
import { useTrendingHashtags } from '@/hooks/useHashtags';
import { EmptyState } from '@/components/shared/EmptyState';

export function TrendingHashtagsTab() {
  const router = useRouter();
  const { data, isLoading, isError } = useTrendingHashtags();

  if (isLoading) {
    return (
      <div>
        {Array.from({ length: 10 }, (_, i) => (
          <div key={i} style={{ padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 6, borderBottom: '1px solid var(--border)' }}>
            <div className="h-2.5 w-24 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3.5 w-40 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-2.5 w-16 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
          </div>
        ))}
      </div>
    );
  }

  if (isError) {
    return <EmptyState title="Couldn't load trends" subtitle="Try again in a moment." />;
  }

  const hashtags = data ?? [];
  if (hashtags.length === 0) {
    return <EmptyState title="No trends yet" subtitle="Trending hashtags from campus will appear here." />;
  }

  return (
    <div>
      {hashtags.map((hashtag, idx) => (
        <button
          key={hashtag.name}
          type="button"
          onClick={() => router.push(`/explore?q=${encodeURIComponent(`#${hashtag.name}`)}`)}
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'space-between',
            gap: 12,
            width: '100%',
            padding: '12px 16px',
            background: 'transparent',
            border: 'none',
            borderBottom: '1px solid var(--border)',
            cursor: 'pointer',
            textAlign: 'left',
            fontFamily: 'inherit',
          }}
          onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
        >
          <div style={{ minWidth: 0 }}>
            <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>
              {idx + 1} · Trending · Campus
            </p>
            <p className="text-[15px] font-bold" style={{ marginTop: 2, color: 'var(--text)' }}>
              #{hashtag.name}
            </p>
            {hashtag.threadCount != null && (
              <p className="text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
                {hashtag.threadCount.toLocaleString()} {hashtag.threadCount === 1 ? 'post' : 'posts'}
              </p>
            )}
          </div>
          <span
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors"
            style={{ marginRight: -4, marginTop: -2, color: 'var(--text-tertiary)' }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--x-blue-soft)'; e.currentTarget.style.color = 'var(--x-blue)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-tertiary)'; }}
          >
            <MoreHorizontal size={18} />
          </span>
        </button>
      ))}
    </div>
  );
}
