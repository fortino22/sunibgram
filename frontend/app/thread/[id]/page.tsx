'use client';

import { useParams } from 'next/navigation';
import { ThreadDetailView } from '@/components/thread/ThreadDetailView';
import { ThreadDetailSkeleton } from '@/components/thread/skeletons/ThreadDetailSkeleton';
import { useThreadChain } from '@/hooks/useThread';

export default function ThreadPage() {
  const { id } = useParams<{ id: string }>();
  const chainQuery = useThreadChain(id);

  if (chainQuery.isLoading) {
    return <ThreadDetailSkeleton />;
  }

  if (chainQuery.isError || chainQuery.data === undefined) {
    return (
      <main
        style={{
          maxWidth: 600,
          margin: '0 auto',
          minHeight: '100vh',
          background: 'var(--bg-card)',
          borderLeft: '1px solid var(--border)',
          borderRight: '1px solid var(--border)',
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
      root={chainQuery.data.root}
      replies={chainQuery.data.replies}
      viewFullThreadId={chainQuery.data.rootThreadId}
    />
  );
}
