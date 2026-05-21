'use client';

import { useEffect, useRef, useCallback } from 'react';
import { EmptyState } from '@/components/shared/EmptyState';
import { NotificationCard } from '@/components/notifications/NotificationCard';
import { NotificationSkeleton } from '@/components/notifications/NotificationSkeleton';
import { useNotifications } from '@/hooks/useNotifications';
import { useDeleteNotification } from '@/hooks/useDeleteNotification';
import { useNotificationStore } from '@/stores/notificationStore';

export default function NotificationsPage() {
  const { data, isLoading, isError, fetchNextPage, hasNextPage, isFetchingNextPage } = useNotifications();
  const { mutate: deleteNotif, isPending: isDeleting, variables: deletingId } = useDeleteNotification();
  const clearUnread = useNotificationStore((s) => s.clearUnread);
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    clearUnread();
  }, [clearUnread]);

  const handleFetchNext = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  // Infinite scroll via IntersectionObserver
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting) handleFetchNext(); },
      { rootMargin: '200px' },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [handleFetchNext]);

  const allNotifications = data?.pages.flatMap((p) => p.data) ?? [];

  return (
    <div>
      {/* Header */}
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
          Notifications
        </div>
      </div>

      {/* Body */}
      {isLoading ? (
        <NotificationSkeleton count={8} />
      ) : isError ? (
        <div style={{ padding: '60px 24px', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
          Failed to load notifications. Please try again.
        </div>
      ) : allNotifications.length === 0 ? (
        <EmptyState
          title="No notifications yet"
          subtitle="When someone likes, comments on, or reposts your threads you'll see it here."
        />
      ) : (
        <>
          {allNotifications.map((notification) => (
            <NotificationCard
              key={notification.id}
              notification={notification}
              onDelete={deleteNotif}
              isDeleting={isDeleting && deletingId === notification.id}
            />
          ))}

          {/* Infinite scroll sentinel */}
          <div ref={sentinelRef} style={{ height: 1 }} />

          {isFetchingNextPage && (
            <NotificationSkeleton count={3} />
          )}

          {!hasNextPage && allNotifications.length > 0 && (
            <div style={{ padding: '24px', textAlign: 'center', fontSize: 13, color: 'var(--text-tertiary)' }}>
              You&apos;re all caught up
            </div>
          )}
        </>
      )}
    </div>
  );
}
