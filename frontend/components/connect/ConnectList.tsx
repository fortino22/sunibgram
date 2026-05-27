'use client';

import { useRef, useEffect } from 'react';
import type { UseInfiniteQueryResult, InfiniteData } from '@tanstack/react-query';
import { EmptyState } from '@/components/shared/EmptyState';
import { ConnectPersonRow } from './ConnectPersonRow';
import type { ExplorePerson } from '@/lib/services/explore';

type SuggestionsQuery = UseInfiniteQueryResult<InfiniteData<ExplorePerson[]>, Error>;

interface ConnectListProps {
  query: SuggestionsQuery;
  emptyTitle: string;
  emptySubtitle: string;
}

export function ConnectList({ query, emptyTitle, emptySubtitle }: ConnectListProps) {
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && query.hasNextPage && !query.isFetchingNextPage) {
          void query.fetchNextPage();
        }
      },
      { rootMargin: '300px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [query]);

  if (query.isLoading) {
    return (
      <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
        Loading…
      </div>
    );
  }

  if (query.isError) {
    return (
      <div style={{ padding: 24, textAlign: 'center', color: 'var(--coral)', fontSize: 14 }}>
        {query.error.message}
      </div>
    );
  }

  const people = query.data?.pages.flatMap((p) => p) ?? [];

  if (people.length === 0) {
    return <EmptyState title={emptyTitle} subtitle={emptySubtitle} />;
  }

  return (
    <div>
      {people.map((person) => (
        <ConnectPersonRow key={person.userId} person={person} />
      ))}
      <div ref={sentinelRef} style={{ height: 1 }} />
      {query.isFetchingNextPage && (
        <div style={{ padding: 16, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
          Loading…
        </div>
      )}
    </div>
  );
}
