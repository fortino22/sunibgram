'use client';

import Link from 'next/link';
import { CalendarDays, MapPin, MessageCircle, Plus } from 'lucide-react';
import { EmptyState } from '@/components/shared/EmptyState';
import { Button } from '@/components/ui/button';
import { useCommunityEvents } from '@/hooks/useCommunityEvents';
import type { CommunityEventListItem } from '@/lib/services/communityEvents';

interface Props {
  communityId: number;
  isAdmin: boolean;
  onCreateEvent?: () => void;
}

function formatRange(start: string, end: string): string {
  const s = new Date(start.replace(' ', 'T') + 'Z');
  const e = new Date(end.replace(' ', 'T') + 'Z');
  if (Number.isNaN(s.getTime())) return '';
  const dateOpts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
  const timeOpts: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' };
  const sameDay = s.toDateString() === e.toDateString();
  if (sameDay) {
    return `${s.toLocaleDateString(undefined, dateOpts)} · ${s.toLocaleTimeString(undefined, timeOpts)} – ${e.toLocaleTimeString(undefined, timeOpts)}`;
  }
  return `${s.toLocaleDateString(undefined, dateOpts)} – ${e.toLocaleDateString(undefined, dateOpts)}`;
}

function EventCard({ event }: { event: CommunityEventListItem }) {
  return (
    <Link
      href={`/community-events/${event.id}`}
      style={{
        display: 'flex',
        flexDirection: 'column',
        borderRadius: 14,
        border: '1px solid var(--border)',
        overflow: 'hidden',
        textDecoration: 'none',
        background: 'var(--bg-card)',
        transition: 'border-color 0.15s',
      }}
      className="ce-card"
    >
      <div className="ce-card-banner" style={{ aspectRatio: '16 / 9', background: 'var(--bg-subtle)', overflow: 'hidden' }}>
        {event.banner_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={event.banner_url} alt={event.title} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
        ) : (
          <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-tertiary)' }}>
            <CalendarDays size={32} />
          </div>
        )}
      </div>
      <div style={{ padding: 14, display: 'flex', flexDirection: 'column', gap: 6 }}>
        <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>{event.title}</h3>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-secondary)' }}>
          <CalendarDays size={14} />
          {formatRange(event.start_time, event.end_time)}
        </div>
        {event.location && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-tertiary)' }}>
            <MapPin size={14} />
            {event.location}
          </div>
        )}
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
          <MessageCircle size={14} />
          {event.comment_count} {event.comment_count === 1 ? 'comment' : 'comments'}
        </div>
      </div>
    </Link>
  );
}

export function CommunityEventsTab({ communityId, isAdmin, onCreateEvent }: Props) {
  const { data, isLoading } = useCommunityEvents(communityId);
  const events = data?.data ?? [];

  return (
    <div style={{ padding: 16 }}>
      {isLoading ? (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: 16 }}>
          {[0, 1, 2].map((i) => (
            <div key={i} style={{ height: 220, borderRadius: 14, background: 'var(--bg-subtle)' }} />
          ))}
        </div>
      ) : events.length === 0 ? (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16 }}>
          <EmptyState
            title="No events yet"
            subtitle={isAdmin ? 'Create your first community event and tag the people who should see it.' : "You'll see events here once an admin tags you in one."}
          />
          {isAdmin && onCreateEvent && (
            <Button onClick={onCreateEvent}>
              <Plus size={16} style={{ marginRight: 6 }} />
              Create event
            </Button>
          )}
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: 16 }}>
          {events.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      )}
    </div>
  );
}
