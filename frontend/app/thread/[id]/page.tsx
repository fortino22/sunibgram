'use client';

import React from 'react';
import { ThreadDetailView } from '@/components/thread/ThreadDetailView';
import { ThreadDetailSkeleton } from '@/components/thread/skeletons/ThreadDetailSkeleton';
import { useThreadChain } from '@/hooks/useThread';

interface ThreadPageProps {
  params: Promise<{ id: string }>;
}

export default function ThreadPage({ params }: ThreadPageProps) {
  const { id } = React.use(params);
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

  const allItems = chainQuery.data.data;
  const numericId = Number(id);
  const root = allItems.find((item) => item.id === numericId);
  const replies = allItems.filter((item) => item.id !== numericId);
  const totalReplies = chainQuery.data.meta.total - 1;

  if (root === undefined) {
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
        Thread not found.
      </main>
    );
  }

  return (
    <ThreadDetailView
      root={root}
      replies={replies}
      totalReplies={Math.max(0, totalReplies)}
    />
  );
}
