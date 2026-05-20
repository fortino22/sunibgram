'use client';

import { EmptyState } from '@/components/shared/EmptyState';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { useMyBookmarks } from '@/hooks/useMyBookmarks';
import type { ThreadItem } from '@/lib/services/threads';

export default function BookmarksPage() {
  const bookmarksQuery = useMyBookmarks();
  const bookmarks = bookmarksQuery.data?.data ?? [];

  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '20px 20px',
        }}
      >
        <div style={{ fontWeight: 500, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em' }}>
          Bookmarks
        </div>
      </div>

      {bookmarksQuery.isLoading ? (
        Array.from({ length: 3 }, (_, index) => <ThreadCardSkeleton key={index} />)
      ) : bookmarksQuery.isError ? (
        <EmptyState title="Bookmarks unavailable" subtitle="Your bookmarked posts could not be loaded right now." />
      ) : bookmarks.length === 0 ? (
        <EmptyState title="No bookmarks yet" subtitle="Posts you save will appear here." />
      ) : (
        bookmarks.map((thread: ThreadItem) => <ThreadCard key={thread.id} thread={thread} />)
      )}
    </div>
  );
}
