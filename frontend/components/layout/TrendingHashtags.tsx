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
      <h2 className="px-4 pb-1 pt-4 text-xl font-extrabold" style={{ color: 'var(--text)' }}>{title}</h2>
      <div>
        {Array.from({ length: DEFAULT_VISIBLE_COUNT }, (_, i) => (
          <div key={i} className="px-4 py-3 space-y-1.5">
            <div className="h-3 w-24 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-4 w-36 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
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
        <h2 className="px-4 pb-1 pt-4 text-xl font-extrabold" style={{ color: 'var(--text)' }}>{title}</h2>
        <p className="px-4 pb-4 text-sm" style={{ color: 'var(--text-tertiary)' }}>Unable to load trending hashtags.</p>
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
      <h2 className="px-4 pb-1 pt-4 text-xl font-extrabold" style={{ color: 'var(--text)' }}>{title}</h2>

      <div>
        {visibleHashtags.map((hashtag) => (
          <Link
            key={hashtag.name}
            href={getHashtagHref(hashtag.name)}
            className="group flex items-start justify-between px-4 py-3 transition-colors"
            style={{ color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0">
              <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>Campus · Trending</p>
              <p className="mt-0.5 text-[15px] font-bold" style={{ color: 'var(--text)' }}>#{hashtag.name}</p>
              {hashtag.post_count != null && (
                <p className="mt-0.5 text-[13px]" style={{ color: 'var(--text-tertiary)' }}>
                  {hashtag.post_count.toLocaleString()} posts
                </p>
              )}
            </div>
            <button
              type="button"
              onClick={(e) => e.preventDefault()}
              className="ml-2 mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full opacity-0 transition-opacity group-hover:opacity-100"
              style={{ color: 'var(--text-tertiary)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
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
          className="w-full px-4 py-4 text-left text-[15px] transition-colors"
          style={{ color: 'var(--accent)' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
        >
          {expanded ? 'Show less' : 'Show more'}
        </button>
      )}
    </section>
  );
}
