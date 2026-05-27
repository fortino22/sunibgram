'use client';

import Link from 'next/link';
import { CalendarDays, MapPin } from 'lucide-react';
import type { EventItem } from '@/lib/services/events';

function parseEventDate(iso: string | null): Date | null {
  if (!iso) return null;
  const date = new Date(iso.replace(' ', 'T'));
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatEventTimeRange(startIso: string | null, endIso: string | null): string {
  const start = parseEventDate(startIso);
  const end = parseEventDate(endIso);

  if (start && end) {
    const sameDay =
      start.getFullYear() === end.getFullYear() &&
      start.getMonth() === end.getMonth() &&
      start.getDate() === end.getDate();

    if (sameDay) {
      const dateLabel = start.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
      const startT = start.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
      const endT = end.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
      return `${dateLabel}, ${startT} – ${endT}`;
    }

    const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' };
    return `${start.toLocaleString('en-US', opts)} – ${end.toLocaleString('en-US', opts)}`;
  }

  const single = start ?? end;
  if (single) {
    return single.toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' });
  }

  return [startIso, endIso].filter(Boolean).join(' – ');
}

export function EventCard({ event, live = false }: { event: EventItem; live?: boolean }) {
  const isOfficial = event.type === 'official';
  return (
    <Link
      href={`/events/${event.id}`}
      className="block"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
        background: 'transparent',
        textDecoration: 'none',
        transition: 'background 0.15s ease',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
    >
      {/* Thumbnail */}
      <div
        style={{
          width: 56,
          height: 56,
          borderRadius: 10,
          background: event.banner
            ? `url(${event.banner}) center/cover no-repeat`
            : 'var(--bg-subtle)',
          border: '1px solid var(--border)',
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {!event.banner && <CalendarDays size={20} style={{ color: 'var(--text-tertiary)' }} />}
      </div>

      {/* Info */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 2 }}>
          <span
            style={{
              fontSize: 15,
              fontWeight: 700,
              color: 'var(--text)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {event.title}
          </span>
          <span
            style={{
              fontSize: 11,
              fontWeight: 600,
              color: isOfficial ? 'var(--accent)' : 'var(--text-tertiary)',
              background: isOfficial ? 'var(--accent-soft)' : 'transparent',
              border: `1px solid ${isOfficial ? 'var(--accent-soft)' : 'var(--border)'}`,
              borderRadius: 999,
              padding: '1px 7px',
              flexShrink: 0,
            }}
          >
            {isOfficial ? 'Official' : 'Unofficial'}
          </span>
          {live && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 11,
                fontWeight: 700,
                color: 'var(--green)',
                flexShrink: 0,
              }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  background: 'var(--green)',
                  animation: 'pulse 1.6s ease-in-out infinite',
                }}
              />
              Live
            </span>
          )}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          {(event.startTime || event.endTime) && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 13,
                color: 'var(--text-tertiary)',
              }}
            >
              <CalendarDays size={12} />
              {formatEventTimeRange(event.startTime, event.endTime)}
            </span>
          )}
          {event.location && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 13,
                color: 'var(--text-tertiary)',
              }}
            >
              <MapPin size={12} />
              {event.location}
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
