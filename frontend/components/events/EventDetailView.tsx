/* eslint-disable @next/next/no-img-element */
'use client';

import { ArrowLeft, Bookmark, CalendarDays, Clock, Link2, LoaderCircle, MapPin } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import { EmptyState } from '@/components/shared/EmptyState';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { useEvent } from '@/hooks/useEvent';
import { useEventActions } from '@/hooks/useEventActions';
import { useEventPosts } from '@/hooks/useEventPosts';
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll';

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
        style={{
          background: 'none',
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
  const [shareNotice, setShareNotice] = useState<string | null>(null);
  const [bookmarked, setBookmarked] = useState(false);

  const query = useEvent(eventId);
  const { bookmark, share } = useEventActions(eventId);

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

  async function handleShare() {
    try {
      const result = await share.mutateAsync();
      if (navigator.share) {
        await navigator.share({ title: event.title, text: event.description, url: result.shareableLink });
      } else if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(result.shareableLink);
      }
      setShareNotice('Link copied!');
      setTimeout(() => setShareNotice(null), 3000);
    } catch (err) {
      setShareNotice(err instanceof Error ? err.message : 'Could not share.');
    }
  }

  function handleBookmark() {
    if (bookmark.isPending) return;
    bookmark.mutate(bookmarked, {
      onSuccess: () => setBookmarked((b) => !b),
    });
  }

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
      <div style={{ padding: '20px 20px 16px', borderBottom: '1px solid var(--border)' }}>
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
          {event.status && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: 'var(--text-tertiary)' }}>
              <Clock size={15} style={{ flexShrink: 0 }} />
              {event.status}
            </span>
          )}
        </div>

        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
          <button
            type="button"
            onClick={handleBookmark}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 7,
              padding: '8px 18px',
              borderRadius: 999,
              border: '1.5px solid var(--border)',
              background: bookmarked ? 'var(--accent)' : 'transparent',
              color: bookmarked ? '#fff' : 'var(--text-secondary)',
              fontSize: 14,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            {bookmark.isPending ? <LoaderCircle size={15} className="animate-spin" /> : <Bookmark size={15} />}
            {bookmarked ? 'Saved' : 'Save'}
          </button>

          <button
            type="button"
            onClick={() => void handleShare()}
            disabled={share.isPending}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 7,
              padding: '8px 18px',
              borderRadius: 999,
              border: '1.5px solid var(--border)',
              background: 'transparent',
              color: 'var(--text-secondary)',
              fontSize: 14,
              fontWeight: 600,
              cursor: share.isPending ? 'not-allowed' : 'pointer',
              opacity: share.isPending ? 0.6 : 1,
            }}
          >
            {share.isPending ? <LoaderCircle size={15} className="animate-spin" /> : <Link2 size={15} />}
            Share
          </button>
        </div>

        {shareNotice && (
          <p style={{ marginTop: 10, fontSize: 13, color: 'var(--text-tertiary)' }}>{shareNotice}</p>
        )}
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
