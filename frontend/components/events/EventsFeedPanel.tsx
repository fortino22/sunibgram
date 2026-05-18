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
  const query = useEventsFeed(active);

  const trending = query.data?.trending ?? [];
  const running = query.data?.running ?? [];
  const future = query.data?.future ?? [];

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
          + Create
        </button>
      </div>

      {dialogOpen && (
        <CreateEventDialog open={dialogOpen} onOpenChange={setDialogOpen} />
      )}

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
          <EventSection label="Trending" events={trending} />
          <EventSection
            label="Running Now"
            events={running}
            emptyState={{
              title: 'No events running now',
              subtitle: 'Check back soon or create an event for later.',
            }}
          />
          <EventSection label="Upcoming" events={future} />
        </>
      )}
    </>
  );
}
