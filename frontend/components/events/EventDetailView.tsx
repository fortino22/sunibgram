/* eslint-disable @next/next/no-img-element */
'use client';

import { ArrowLeft, CalendarDays, MapPin, Pencil } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import { EmptyState } from '@/components/shared/EmptyState';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { useEvent } from '@/hooks/useEvent';
import { useEventJoin } from '@/hooks/useEventJoin';
import { useEventPosts } from '@/hooks/useEventPosts';
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll';
import { useAuthStore } from '@/stores/authStore';
import { useComposeStore } from '@/stores/composeStore';

type Tab = 'official' | 'discussion';

function parseEventDate(iso: string | null): Date | null {
  if (!iso) return null;
  const date = new Date(iso.replace(' ', 'T'));
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatDateTimeRange(startIso: string | null, endIso: string | null): string {
  const start = parseEventDate(startIso);
  const end = parseEventDate(endIso);

  if (start && end) {
    const sameDay =
      start.getFullYear() === end.getFullYear() &&
      start.getMonth() === end.getMonth() &&
      start.getDate() === end.getDate();

    if (sameDay) {
      const dateLabel = start.toLocaleDateString('en-US', {
        month: 'long',
        day: 'numeric',
        year: 'numeric',
      });
      const startTimeLabel = start.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit',
      });
      const endTimeLabel = end.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit',
      });

      return `${dateLabel}, ${startTimeLabel} - ${endTimeLabel}`;
    }

    const fullOptions: Intl.DateTimeFormatOptions = {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    };

    return `${start.toLocaleString('en-US', fullOptions)} - ${end.toLocaleString('en-US', fullOptions)}`;
  }

  if (start) {
    return start.toLocaleString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  }

  if (end) {
    return end.toLocaleString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  }

  return [startIso, endIso].filter(Boolean).join(' - ');
}

function EventDetailHeader({ onBack }: { onBack: () => void }) {
  return (
    <div
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 20,
        padding: '12px 16px',
        display: 'flex',
        alignItems: 'center',
        gap: 14,
        borderBottom: '1px solid var(--border)',
        background: 'oklch(10.5% 0.005 260 / 0.92)',
        backdropFilter: 'blur(12px)',
      }}
    >
      <button
        type="button"
        aria-label="Go back"
        onClick={onBack}
        className="btn-ghost"
        style={{
          border: 'none',
          cursor: 'pointer',
          color: 'var(--text)',
          padding: 6,
          borderRadius: '50%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <ArrowLeft size={20} strokeWidth={1.8} />
      </button>
      <div style={{ fontWeight: 700, fontSize: 17, color: 'var(--text)', letterSpacing: '-0.02em' }}>
        Event
      </div>
    </div>
  );
}

function EventPostsPanel({
  eventId,
  mode,
  active,
}: {
  eventId: string;
  mode: Tab;
  active: boolean;
}) {
  const query = useEventPosts(eventId, mode, active);
  const items = useMemo(
    () => query.data?.pages.flatMap((page) => page.data) ?? [],
    [query.data]
  );

  const sentinelRef = useInfiniteScroll({
    enabled: active,
    hasNextPage: query.hasNextPage ?? false,
    isFetchingNextPage: query.isFetchingNextPage,
    onLoadMore: () => {
      if (query.hasNextPage && !query.isFetchingNextPage) void query.fetchNextPage();
    },
  });

  if (query.isLoading) {
    return (
      <>
        <ThreadCardSkeleton />
        <ThreadCardSkeleton />
      </>
    );
  }

  if (query.isError) {
    return (
      <div style={{ padding: '32px 20px', fontSize: 14, color: 'var(--coral)' }}>
        {query.error.message}
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div style={{ padding: '40px 20px' }}>
        <EmptyState
          title={mode === 'official' ? 'No official posts yet' : 'No discussion yet'}
          subtitle={
            mode === 'official'
              ? 'Organizer posts will appear here.'
              : 'Community discussion will appear here.'
          }
        />
      </div>
    );
  }

  return (
    <>
      {items.map((thread) => (
        <ThreadCard key={`${mode}-${thread.id}`} thread={thread} />
      ))}
      <div ref={sentinelRef} />
      {query.isFetchingNextPage && (
        <div style={{ padding: '20px', fontSize: 14, color: 'var(--text-tertiary)' }}>
          Loading more...
        </div>
      )}
    </>
  );
}

export function EventDetailView({ eventId }: { eventId: string }) {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<Tab>('official');
  const [joinHovered, setJoinHovered] = useState(false);
  const user = useAuthStore((s) => s.user);
  const { openOfficialEventPost } = useComposeStore();

  const query = useEvent(eventId);
  const { isJoined, join, leave } = useEventJoin(eventId);

  if (query.isLoading) {
    return (
      <div>
        <EventDetailHeader onBack={() => router.back()} />
        <div className="animate-pulse" style={{ padding: 20 }}>
          <div style={{ height: 280, borderRadius: 20, background: 'var(--bg-subtle)', marginBottom: 20 }} />
          <div style={{ height: 28, width: '60%', borderRadius: 99, background: 'var(--bg-subtle)', marginBottom: 12 }} />
          <div style={{ height: 16, width: '100%', borderRadius: 99, background: 'var(--bg-subtle)', marginBottom: 8 }} />
          <div style={{ height: 16, width: '80%', borderRadius: 99, background: 'var(--bg-subtle)' }} />
        </div>
      </div>
    );
  }

  if (query.isError || !query.data) {
    return (
      <div>
        <EventDetailHeader onBack={() => router.back()} />
        <div style={{ padding: '40px 20px', fontSize: 14, color: 'var(--coral)' }}>
          Event could not be loaded.
        </div>
      </div>
    );
  }

  const event = query.data;
  const isCreator = user?.id === event.userId;

  return (
    <div>
      <EventDetailHeader onBack={() => router.back()} />

      {/* Banner */}
      {event.banner && (
        <div style={{ borderBottom: '1px solid var(--border)' }}>
          <img
            src={event.banner}
            alt={event.title}
            style={{ width: '100%', maxHeight: 280, objectFit: 'cover', display: 'block' }}
          />
        </div>
      )}

      {/* Event info */}
      <div style={{ padding: '20px 20px 16px', borderBottom: '1px solid var(--border)', position: 'relative' }}>
        <div style={{ marginBottom: 12 }}>
          <span
            style={{
              padding: '4px 12px',
              borderRadius: 999,
              fontSize: 13,
              fontWeight: 600,
              background: event.type === 'official' ? 'var(--accent-soft)' : 'var(--bg-subtle)',
              color: event.type === 'official' ? 'var(--accent)' : 'var(--text-secondary)',
              border: `1px solid ${event.type === 'official' ? 'var(--accent-soft)' : 'var(--border)'}`,
            }}
          >
            {event.type === 'official' ? 'Official' : 'Unofficial'}
          </span>
        </div>

        <h1 style={{ fontSize: 24, fontWeight: 800, color: 'var(--text)', letterSpacing: '-0.02em', marginBottom: 10 }}>
          {event.title}
        </h1>

        <p style={{ fontSize: 15, color: 'var(--text-secondary)', lineHeight: 1.7, marginBottom: 16 }}>
          {event.description}
        </p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 18 }}>
          {event.location && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: 'var(--text-tertiary)' }}>
              <MapPin size={15} style={{ flexShrink: 0 }} />
              {event.location}
            </span>
          )}
          {(event.startTime || event.endTime) && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: 'var(--text-tertiary)' }}>
              <CalendarDays size={15} style={{ flexShrink: 0 }} />
              {formatDateTimeRange(event.startTime, event.endTime)}
            </span>
          )}
        </div>

        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
          {/* Join / Leave */}
          {isJoined ? (
            <button
              type="button"
              className="hover-dim"
              onClick={() => leave.mutate()}
              disabled={leave.isPending}
              onMouseEnter={() => setJoinHovered(true)}
              onMouseLeave={() => setJoinHovered(false)}
              style={{
                padding: '7px 18px',
                borderRadius: 999,
                border: `1px solid ${joinHovered && !leave.isPending ? 'var(--error, #e53e3e)' : '#808080'}`,
                background: 'transparent',
                color: joinHovered && !leave.isPending ? 'var(--error, #e53e3e)' : 'var(--text)',
                fontSize: 14,
                fontWeight: 600,
                cursor: leave.isPending ? 'default' : 'pointer',
                opacity: leave.isPending ? 0.6 : 1,
                transition: 'color 0.15s, border-color 0.15s',
              }}
            >
              {leave.isPending ? 'Leaving…' : joinHovered ? 'Leave' : 'Joined'}
            </button>
          ) : (
            <button
              type="button"
              className="btn-solid"
              onClick={() => join.mutate()}
              disabled={join.isPending}
              style={{
                padding: '7px 18px',
                borderRadius: 999,
                border: 'none',
                background: 'var(--text)',
                color: 'var(--bg)',
                fontSize: 14,
                fontWeight: 700,
                cursor: join.isPending ? 'default' : 'pointer',
                opacity: join.isPending ? 0.6 : 1,
              }}
            >
              {join.isPending ? 'Joining…' : 'Join'}
            </button>
          )}

          {/* Official post button — creator only */}
          {isCreator && (
            <button
              type="button"
              className="btn-ghost"
              onClick={() => openOfficialEventPost(eventId, event.title)}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 7,
                padding: '8px 18px',
                borderRadius: 999,
                border: '1.5px solid var(--border)',
                color: 'var(--text-secondary)',
                fontSize: 14,
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              <Pencil size={15} />
              Post Official
            </button>
          )}
        </div>
      </div>

      {/* Tab bar */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          display: 'flex',
          borderBottom: '1px solid var(--border)',
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
        }}
      >
        {(['official', 'discussion'] as const).map((tab) => (
          <button
            key={tab}
            type="button"
            className="hover-dim"
            onClick={() => setActiveTab(tab)}
            style={{
              flex: 1,
              padding: '13px 0',
              background: 'none',
              border: 'none',
              borderBottom: activeTab === tab ? '2px solid var(--accent)' : '2px solid transparent',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: activeTab === tab ? 700 : 400,
              color: activeTab === tab ? 'var(--text)' : 'var(--text-tertiary)',
              fontFamily: 'inherit',
              transition: 'all 0.15s',
            }}
          >
            {tab === 'official' ? 'Official' : 'Discussion'}
          </button>
        ))}
      </div>

      {activeTab === 'official' && (
        <EventPostsPanel eventId={eventId} mode="official" active />
      )}
      {activeTab === 'discussion' && (
        <EventPostsPanel eventId={eventId} mode="discussion" active />
      )}

    </div>
  );
}
