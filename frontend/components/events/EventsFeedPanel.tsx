'use client';

import { ArrowLeft } from 'lucide-react';
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

export function EventsFeedPanel({ active }: { active: boolean }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [trendingFilterActive, setTrendingFilterActive] = useState(false);
  const [runningFilterActive, setRunningFilterActive] = useState(false);
  const [upcomingFilterActive, setUpcomingFilterActive] = useState(false);
  const query = useEventsFeed(active);

  const trending = query.data?.trending ?? [];
  const running = query.data?.running ?? [];
  const future = query.data?.future ?? [];
  const showAllEvents = !trendingFilterActive && !runningFilterActive && !upcomingFilterActive;
  const showTrendingSection = showAllEvents || trendingFilterActive;
  const showRunningSection = showAllEvents || runningFilterActive;
  const showUpcomingSection = showAllEvents || upcomingFilterActive;

  function resetFilters() {
    setTrendingFilterActive(false);
    setRunningFilterActive(false);
    setUpcomingFilterActive(false);
  }

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
            padding: '7px 18px',
            borderRadius: 999,
            background: 'var(--accent)',
            color: '#fff',
            fontSize: 14,
            fontWeight: 700,
            border: 'none',
            cursor: 'pointer',
          }}
        >
          Create
        </button>
      </div>

      {dialogOpen && (
        <CreateEventDialog open={dialogOpen} onOpenChange={setDialogOpen} />
      )}

      <div
        style={{
          display: 'flex',
          gap: 10,
          padding: '14px 20px 10px',
          overflowX: 'auto',
          borderBottom: '1px solid var(--border)',
          background: 'var(--bg-card)',
        }}
      >
        <button
          type="button"
          onClick={resetFilters}
          aria-pressed={showAllEvents}
          style={{
            padding: '8px 14px',
            borderRadius: 999,
            border: `1px solid ${showAllEvents ? 'var(--accent)' : 'var(--border)'}`,
            background: showAllEvents ? 'var(--accent)' : 'transparent',
            color: showAllEvents ? '#fff' : 'var(--text-secondary)',
            fontSize: 13,
            fontWeight: 700,
            whiteSpace: 'nowrap',
            cursor: 'pointer',
            transition: 'all 0.15s ease',
          }}
        >
          See All
        </button>
        <button
          type="button"
          onClick={() => setTrendingFilterActive((current) => !current)}
          aria-pressed={trendingFilterActive}
          style={{
            padding: '8px 14px',
            borderRadius: 999,
            border: `1px solid ${trendingFilterActive ? 'var(--accent)' : 'var(--border)'}`,
            background: trendingFilterActive ? 'var(--accent)' : 'transparent',
            color: trendingFilterActive ? '#fff' : 'var(--text-secondary)',
            fontSize: 13,
            fontWeight: 700,
            whiteSpace: 'nowrap',
            cursor: 'pointer',
            transition: 'all 0.15s ease',
          }}
        >
          Trending
        </button>
        <button
          type="button"
          onClick={() => setRunningFilterActive((current) => !current)}
          aria-pressed={runningFilterActive}
          style={{
            padding: '8px 14px',
            borderRadius: 999,
            border: `1px solid ${runningFilterActive ? 'var(--accent)' : 'var(--border)'}`,
            background: runningFilterActive ? 'var(--accent)' : 'transparent',
            color: runningFilterActive ? '#fff' : 'var(--text-secondary)',
            fontSize: 13,
            fontWeight: 700,
            whiteSpace: 'nowrap',
            cursor: 'pointer',
            transition: 'all 0.15s ease',
          }}
        >
          Running Now
        </button>
        <button
          type="button"
          onClick={() => setUpcomingFilterActive((current) => !current)}
          aria-pressed={upcomingFilterActive}
          style={{
            padding: '8px 14px',
            borderRadius: 999,
            border: `1px solid ${upcomingFilterActive ? 'var(--accent)' : 'var(--border)'}`,
            background: upcomingFilterActive ? 'var(--accent)' : 'transparent',
            color: upcomingFilterActive ? '#fff' : 'var(--text-secondary)',
            fontSize: 13,
            fontWeight: 700,
            whiteSpace: 'nowrap',
            cursor: 'pointer',
            transition: 'all 0.15s ease',
          }}
        >
          Upcoming
        </button>
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
