'use client';

import { Plus, Search, X } from 'lucide-react';
import { useState } from 'react';
import { CreateEventDialog } from '@/components/events/CreateEventDialog';
import { EventCard } from '@/components/events/EventCard';
import { EventCardSkeleton } from '@/components/events/EventCardSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { useEventsFeed } from '@/hooks/useEventsFeed';
import { useMyJoinedEventItems } from '@/hooks/useMyJoinedEventItems';
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


type Filter = 'all' | 'trending' | 'running' | 'upcoming' | 'joined';

function matchesSearch(event: EventItem, q: string) {
  const lower = q.toLowerCase();
  return (
    event.title.toLowerCase().includes(lower) ||
    (event.description ?? '').toLowerCase().includes(lower)
  );
}

export function EventsFeedPanel({ active }: { active: boolean }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [filter, setFilter] = useState<Filter>('all');
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const query = useEventsFeed(active);
  const joinedQuery = useMyJoinedEventItems();

  const applySearch = (list: EventItem[]) =>
    searchQuery.trim() ? list.filter((e) => matchesSearch(e, searchQuery)) : list;

  const trending = applySearch(query.data?.trending ?? []);
  const running = applySearch(query.data?.running ?? []);
  const future = applySearch(query.data?.future ?? []);
  const joined = applySearch(joinedQuery.data ?? []);
  const showTrendingSection = filter === 'all' || filter === 'trending';
  const showRunningSection = filter === 'all' || filter === 'running';
  const showUpcomingSection = filter === 'all' || filter === 'upcoming';

  return (
    <>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        {/* Title row */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '20px 20px 0',
          }}
        >
          <div style={{ fontWeight: 700, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em' }}>
            Events
          </div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <button
              type="button"
              aria-label="Search events"
              onClick={() => {
                setSearchOpen((o) => !o);
                setSearchQuery('');
              }}
              style={{
                width: 34,
                height: 34,
                borderRadius: '50%',
                border: `1px solid ${searchOpen ? 'var(--text)' : 'var(--border)'}`,
                background: searchOpen ? 'var(--text)' : 'transparent',
                color: searchOpen ? 'var(--bg)' : 'var(--text)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Search size={16} />
            </button>
            <button
              type="button"
              onClick={() => setDialogOpen(true)}
              style={{
                padding: '7px 14px',
                borderRadius: 999,
                border: 'none',
                background: 'var(--text)',
                color: 'var(--bg)',
                fontSize: 13,
                fontWeight: 700,
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
        </div>

        {/* Search bar */}
        {searchOpen && (
          <div style={{ padding: '12px 20px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
            <div style={{ flex: 1, position: 'relative' }}>
              <Search
                size={15}
                style={{
                  position: 'absolute',
                  left: 12,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-tertiary)',
                  pointerEvents: 'none',
                }}
              />
              <input
                autoFocus
                type="text"
                placeholder="Search events…"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 36px 8px 36px',
                  borderRadius: 999,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-secondary)',
                  color: 'var(--text)',
                  fontSize: 14,
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
              {searchQuery && (
                <button
                  type="button"
                  onClick={() => setSearchQuery('')}
                  style={{
                    position: 'absolute',
                    right: 10,
                    top: '50%',
                    transform: 'translateY(-50%)',
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--text-tertiary)',
                    display: 'flex',
                    padding: 0,
                  }}
                >
                  <X size={14} />
                </button>
              )}
            </div>
          </div>
        )}

        {/* Filter chips */}
        <div
          style={{
            display: 'flex',
            gap: 8,
            padding: '12px 16px',
            overflowX: 'auto',
            scrollbarWidth: 'none',
          }}
        >
          {([
            { key: 'all', label: 'All' },
            { key: 'trending', label: 'Trending' },
            { key: 'running', label: 'Running Now' },
            { key: 'upcoming', label: 'Upcoming' },
            { key: 'joined', label: 'Joined' },
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
      </div>

      {dialogOpen && (
        <CreateEventDialog open={dialogOpen} onOpenChange={setDialogOpen} />
      )}

      {filter === 'joined' ? (
        joinedQuery.isLoading ? (
          <>
            <EventCardSkeleton />
            <EventCardSkeleton />
          </>
        ) : joinedQuery.isError ? (
          <div style={{ padding: '40px 20px', fontSize: 14, color: 'var(--coral)' }}>
            {joinedQuery.error.message}
          </div>
        ) : joined.length === 0 ? (
          <div style={{ padding: '40px 20px' }}>
            <EmptyState title="No joined events" subtitle="Join an event to see it here." />
          </div>
        ) : (
          joined.map((event) => <EventCard key={event.id} event={event} />)
        )
      ) : query.isLoading ? (
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
          {showTrendingSection && trending.length > 0 && (
            <>
              {filter === 'all' && <SectionHeader label="Trending" />}
              {trending.map((event) => <EventCard key={event.id} event={event} />)}
            </>
          )}
          {showRunningSection && (
            <>
              {filter === 'all' && running.length > 0 && <SectionHeader label="Running Now" />}
              {running.length > 0
                ? running.map((event) => <EventCard key={event.id} event={event} />)
                : filter === 'running' && (
                  <div style={{ padding: '40px 20px' }}>
                    <EmptyState title="No events running now" subtitle="Check back soon or create an event for later." />
                  </div>
                )}
            </>
          )}
          {showUpcomingSection && future.length > 0 && (
            <>
              {filter === 'all' && <SectionHeader label="Upcoming" />}
              {future.map((event) => <EventCard key={event.id} event={event} />)}
            </>
          )}
        </>
      )}
    </>
  );
}
