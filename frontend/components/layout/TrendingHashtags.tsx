'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { useTrendingHashtags } from '@/hooks/useHashtags';

const DEFAULT_VISIBLE_COUNT = 5;

interface TrendingHashtagsProps {
  initialVisibleCount?: number;
  title?: string;
}

function getHashtagHref(name: string) {
  return `/explore?q=${encodeURIComponent(`#${name}`)}`;
}

function TrendingHashtagsLoading({ title }: { title: string }) {
  return (
    <div className="overflow-hidden rounded-2xl bg-(--bg-card)">
      <h2 className="px-4 pb-0 pt-3 text-xl font-extrabold text-(--text)">{title}</h2>
      <div>
        {Array.from({ length: DEFAULT_VISIBLE_COUNT }, (_, index) => (
          <div key={index} className="px-4 py-3 space-y-1.5">
            <div className="h-3 w-28 animate-pulse rounded bg-(--bg-subtle)" />
            <div className="h-4 w-40 animate-pulse rounded bg-(--bg-subtle)" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function TrendingHashtags({
  initialVisibleCount = DEFAULT_VISIBLE_COUNT,
  title = 'Trending',
}: TrendingHashtagsProps) {
  const [expanded, setExpanded] = useState(false);
  const trendingQuery = useTrendingHashtags();

  if (trendingQuery.isLoading) {
    return <TrendingHashtagsLoading title={title} />;
  }

  if (trendingQuery.isError) {
    return (
      <div className="overflow-hidden rounded-2xl bg-(--bg-card)">
        <h2 className="px-4 pb-1 pt-3 text-xl font-extrabold text-(--text)">{title}</h2>
        <p className="px-4 pb-4 text-sm text-(--text-tertiary)">
          Unable to load trending hashtags right now.
        </p>
      </div>
    );
  }

  const hashtags = trendingQuery.data ?? [];

  if (hashtags.length === 0) {
    return null;
  }

  const canToggle = hashtags.length > initialVisibleCount;
  const visibleHashtags = expanded ? hashtags : hashtags.slice(0, initialVisibleCount);

  return (
    <section className="overflow-hidden rounded-2xl bg-(--bg-card)">
      <h2 className="px-4 pb-1 pt-3 text-xl font-extrabold text-(--text)">{title}</h2>

      <div>
        {visibleHashtags.map((hashtag, index) => (
          <Link
            key={hashtag.name}
            href={getHashtagHref(hashtag.name)}
            className="flex items-start justify-between px-4 py-3 transition-colors hover:bg-(--bg-subtle)"
            style={index < visibleHashtags.length - 1 ? { borderBottom: '1px solid var(--border)' } : undefined}
          >
            <div className="min-w-0">
              <p className="text-xs text-(--text-tertiary)">Campus · Trending</p>
              <p className="mt-0.5 text-[15px] font-bold text-(--text)">#{hashtag.name}</p>
            </div>
          </Link>
        ))}
      </div>

      {canToggle && (
        <div className="border-t px-4 py-3" style={{ borderColor: 'var(--border)' }}>
          <Button
            type="button"
            variant="ghost"
            className="w-full justify-start rounded-lg px-0 text-sm text-(--accent) hover:bg-transparent hover:text-(--accent) hover:underline"
            onClick={() => setExpanded((current) => !current)}
          >
            {expanded ? 'Show less' : 'Show more'}
          </Button>
        </div>
      )}
    </section>
  );
}
