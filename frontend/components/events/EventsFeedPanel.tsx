'use client';

import { ArrowLeft, Plus } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { CreateEventDialog } from '@/components/events/CreateEventDialog';
import { EventCard } from '@/components/events/EventCard';
import { EventCardSkeleton } from '@/components/events/EventCardSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { useEventsFeed } from '@/hooks/useEventsFeed';
import type { EventItem } from '@/lib/services/events';

function SectionHeader({ label }: { label: string }) {
  return (
    <div
      style={{
        padding: '10px 20px 6px',
        fontSize: 13,
        fontWeight: 700,
        letterSpacing: '0.04em',
        textTransform: 'uppercase',
        color: 'var(--text-tertiary)',
        borderBottom: '1px solid var(--border)',
      }}
    >
      {label}
    </div>
  );
}

function EventSection({
  label,
  events,
  emptyState,
}: {
  label: string;
  events: EventItem[];
  emptyState?: { title: string; subtitle: string };
}) {
  return (
    <>
      <SectionHeader label={label} />
      {events.length > 0 ? (
        events.map((event) => (
          <EventCard key={event.id} event={event} />
        ))
      ) : emptyState ? (
        <div style={{ padding: '36px 20px' }}>
          <EmptyState title={emptyState.title} subtitle={emptyState.subtitle} />
        </div>
      ) : null}
    </>
  );
}

type Filter = 'all' | 'trending' | 'running' | 'upcoming';

export function EventsFeedPanel({ active }: { active: boolean }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [filter, setFilter] = useState<Filter>('all');
  const query = useEventsFeed(active);

  const trending = query.data?.trending ?? [];
  const running = query.data?.running ?? [];
  const future = query.data?.future ?? [];
  const showTrendingSection = filter === 'all' || filter === 'trending';
  const showRunningSection = filter === 'all' || filter === 'running';
  const showUpcomingSection = filter === 'all' || filter === 'upcoming';

  return (
    <>
      {/* Sub-header — matches homepage sticky style */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 9,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '14px 20px',
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button
            type="button"
            aria-label="Go back"
            onClick={() => router.back()}
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
          <span style={{ fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>
            Events
          </span>
        </div>
        <button
          type="button"
          onClick={() => setDialogOpen(true)}
          style={{
            padding: '7px 14px',
            borderRadius: 999,
            background: 'var(--text)',
            color: 'var(--bg)',
            fontSize: 13,
            fontWeight: 700,
            border: 'none',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          <Plus size={14} strokeWidth={2.5} />
          Create
        </button>
      </div>

      {dialogOpen && (
        <CreateEventDialog open={dialogOpen} onOpenChange={setDialogOpen} />
      )}

      <div
        style={{
          display: 'flex',
          gap: 8,
          padding: '4px 16px 12px',
          overflowX: 'auto',
          borderBottom: '1px solid var(--border)',
          scrollbarWidth: 'none',
        }}
      >
        {([
          { key: 'all', label: 'All' },
          { key: 'trending', label: 'Trending' },
          { key: 'running', label: 'Running Now' },
          { key: 'upcoming', label: 'Upcoming' },
        ] as { key: Filter; label: string }[]).map(({ key, label }) => (
          <button
            key={key}
            type="button"
            onClick={() => setFilter(key)}
            style={{
              padding: '6px 16px',
              borderRadius: 999,
              border: `1px solid ${filter === key ? 'var(--text)' : 'var(--border)'}`,
              background: filter === key ? 'var(--text)' : 'transparent',
              color: filter === key ? 'var(--bg)' : 'var(--text)',
              fontSize: 13,
              fontWeight: 600,
              whiteSpace: 'nowrap',
              flexShrink: 0,
              cursor: 'pointer',
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {query.isLoading ? (
        <>
          <EventCardSkeleton />
          <EventCardSkeleton />
          <EventCardSkeleton />
        </>
      ) : query.isError ? (
        <div style={{ padding: '40px 20px', fontSize: 14, color: 'var(--coral)' }}>
          {query.error.message}
        </div>
      ) : (
        <>
          {showTrendingSection && (
            <EventSection label="Trending" events={trending} />
          )}
          {showRunningSection && (
            <EventSection
              label="Running Now"
              events={running}
              emptyState={{
                title: 'No events running now',
                subtitle: 'Check back soon or create an event for later.',
              }}
            />
          )}
          {showUpcomingSection && (
            <EventSection label="Upcoming" events={future} />
          )}
        </>
      )}
    </>
  );
}
