'use client';

import { useParams } from 'next/navigation';
import { ThreadDetailView } from '@/components/thread/ThreadDetailView';
import { ThreadDetailSkeleton } from '@/components/thread/skeletons/ThreadDetailSkeleton';
import { useThreadChain } from '@/hooks/useThread';

export default function ThreadPage() {
  const { id } = useParams<{ id: string }>();
  const chain = useThreadChain(id);

  if (chain.isLoading) {
    return <ThreadDetailSkeleton />;
  }

  if (chain.isError || chain.root === undefined) {
    return (
      <main
        style={{
          maxWidth: 600,
          margin: '0 auto',
          minHeight: '100vh',
          background: 'var(--bg-card)',
          color: 'var(--text)',
          padding: 24,
        }}
      >
        Thread could not be loaded.
      </main>
    );
  }

  return (
    <ThreadDetailView
      root={chain.root}
      replies={chain.replies}
      hasNextPage={chain.hasNextPage}
      fetchNextPage={chain.fetchNextPage}
      isFetchingNextPage={chain.isFetchingNextPage}
    />
  );
}
