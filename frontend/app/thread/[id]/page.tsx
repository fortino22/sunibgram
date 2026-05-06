'use client';

import React from 'react';
import { ThreadDetailView } from '@/components/thread/ThreadDetailView';
import { useThreadChain } from '@/hooks/useThread';

interface ThreadPageProps {
  params: Promise<{ id: string }>;
}

function LoadingSkeleton() {
  return (
    <main
      style={{
        maxWidth: 600,
        margin: '0 auto',
        minHeight: '100vh',
        background: 'var(--bg-card)',
        borderLeft: '1px solid var(--border)',
        borderRight: '1px solid var(--border)',
      }}
    >
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '14px 16px',
        }}
      >
        <div style={{ width: 120, height: 22, borderRadius: 6, background: 'var(--bg-subtle)' }} />
      </div>
      {Array.from({ length: 4 }, (_, i) => (
        <div key={i} style={{ display: 'flex', gap: 12, padding: 16, borderBottom: '1px solid var(--border)' }}>
          <div style={{ width: 44, height: 44, borderRadius: '50%', background: 'var(--bg-subtle)', flexShrink: 0 }} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 10 }}>
            <div style={{ width: '45%', height: 14, borderRadius: 5, background: 'var(--bg-subtle)' }} />
            <div style={{ width: '100%', height: 14, borderRadius: 5, background: 'var(--bg-subtle)' }} />
            <div style={{ width: '72%', height: 14, borderRadius: 5, background: 'var(--bg-subtle)' }} />
          </div>
        </div>
      ))}
    </main>
  );
}

export default function ThreadPage({ params }: ThreadPageProps) {
  const { id } = React.use(params);
  const chainQuery = useThreadChain(id);

  if (chainQuery.isLoading) {
    return <LoadingSkeleton />;
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
