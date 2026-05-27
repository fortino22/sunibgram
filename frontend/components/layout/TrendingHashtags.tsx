'use client';

import { useState } from 'react';
import Link from 'next/link';
import { MoreHorizontal } from 'lucide-react';
import { useTrendingHashtags } from '@/hooks/useHashtags';

const DEFAULT_VISIBLE_COUNT = 4;

interface TrendingHashtagsProps {
  initialVisibleCount?: number;
  title?: string;
}

function getHashtagHref(name: string) {
  return `/explore?q=${encodeURIComponent(`#${name}`)}`;
}

function TrendingHashtagsLoading({ title }: { title: string }) {
  return (
    <div
      className="overflow-hidden rounded-2xl"
      style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
    >
      <h2 className="text-[18px] font-extrabold" style={{ padding: '14px 16px 8px', color: 'var(--text)' }}>{title}</h2>
      <div>
        {Array.from({ length: DEFAULT_VISIBLE_COUNT }, (_, i) => (
          <div key={i} style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
            <div className="h-2.5 w-20 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3.5 w-32 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
          </div>
        ))}
      </div>
    </div>
  );
}

export function TrendingHashtags({
  initialVisibleCount = DEFAULT_VISIBLE_COUNT,
  title = 'Trending on campus',
}: TrendingHashtagsProps) {
  const [expanded, setExpanded] = useState(false);
  const trendingQuery = useTrendingHashtags();

  if (trendingQuery.isLoading) return <TrendingHashtagsLoading title={title} />;

  if (trendingQuery.isError) {
    return (
      <div
        className="overflow-hidden rounded-2xl"
        style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
      >
        <h2 className="text-[18px] font-extrabold" style={{ padding: '14px 16px 8px', color: 'var(--text)' }}>{title}</h2>
        <p className="text-sm" style={{ padding: '4px 16px 14px', color: 'var(--text-tertiary)' }}>Unable to load trending hashtags.</p>
      </div>
    );
  }

  const hashtags = trendingQuery.data ?? [];
  if (hashtags.length === 0) return null;

  const canToggle = hashtags.length > initialVisibleCount;
  const visibleHashtags = expanded ? hashtags : hashtags.slice(0, initialVisibleCount);

  return (
    <section
      className="overflow-hidden rounded-2xl"
      style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
    >
      <h2 className="text-[18px] font-extrabold" style={{ padding: '14px 16px 8px', color: 'var(--text)' }}>{title}</h2>

      <div>
        {visibleHashtags.map((hashtag) => (
          <Link
            key={hashtag.name}
            href={getHashtagHref(hashtag.name)}
            className="group flex items-start justify-between transition-colors"
            style={{ padding: '12px 16px', color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0">
              <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>Trending · Campus</p>
              <p className="text-[15px] font-bold" style={{ marginTop: 2, color: 'var(--text)' }}>#{hashtag.name}</p>
              {hashtag.threadCount != null && (
                <p className="text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
                  {hashtag.threadCount.toLocaleString()} {hashtag.threadCount === 1 ? 'post' : 'posts'}
                </p>
              )}
            </div>
            <button
              type="button"
              onClick={(e) => e.preventDefault()}
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors"
              style={{ marginRight: -6, marginTop: -2, color: 'var(--text-tertiary)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(29,155,240,0.1)'; (e.currentTarget as HTMLElement).style.color = 'var(--accent)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; (e.currentTarget as HTMLElement).style.color = 'var(--text-tertiary)'; }}
              aria-label="More options"
            >
              <MoreHorizontal size={18} />
            </button>
          </Link>
        ))}
      </div>

      {canToggle && (
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="w-full text-left text-[15px] transition-colors"
          style={{ padding: '12px 16px', color: 'var(--accent)' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
        >
          {expanded ? 'Show less' : 'Show more'}
        </button>
      )}
    </section>
  );
}
