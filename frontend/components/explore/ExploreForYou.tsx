'use client';

import { useEffect, useRef } from 'react';
import { TrendingEvents } from '@/components/layout/TrendingEvents';
import { TrendingHashtags } from '@/components/layout/TrendingHashtags';
import { WhoToFollow } from '@/components/layout/WhoToFollow';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { useThreads } from '@/hooks/useThreads';
import type { ThreadItem } from '@/lib/services/threads';

export function ExploreForYou() {
  const feed = useThreads();
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && feed.hasNextPage && !feed.isFetchingNextPage) {
          void feed.fetchNextPage();
        }
      },
      { rootMargin: '300px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [feed]);

  const threads = feed.data?.pages.flatMap((p) => p.data) ?? [];

  return (
    <div>
      {/* Flat full-width sections like X — no card boxes */}
      <TrendingEvents detailed bare />
      <TrendingHashtags title="What's happening" bare />
      <WhoToFollow bare />

      {/* Posts for you — home feed */}
      <div
        className="text-[20px] font-extrabold"
        style={{ padding: '16px 16px 8px', color: 'var(--text)', letterSpacing: '-0.02em' }}
      >
        Posts for you
      </div>

      {feed.isLoading ? (
        <div>{Array.from({ length: 4 }, (_, i) => <ThreadCardSkeleton key={i} />)}</div>
      ) : threads.length === 0 ? (
        <EmptyState title="Nothing here yet" subtitle="Posts from campus will appear here." />
      ) : (
        <div>
          {threads.map((thread: ThreadItem) => (
            <ThreadCard key={thread.id} thread={thread} />
          ))}
          <div ref={sentinelRef} style={{ height: 1 }} />
          {feed.isFetchingNextPage && (
            <div>{Array.from({ length: 2 }, (_, i) => <ThreadCardSkeleton key={i} />)}</div>
          )}
        </div>
      )}
    </div>
  );
}
