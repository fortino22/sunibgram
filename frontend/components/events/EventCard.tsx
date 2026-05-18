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

export function EventCard({ event }: { event: EventItem }) {
  return (
    <Link
      href={`/events/${event.id}`}
      className="block transition-opacity hover:opacity-95"
      style={{ borderBottom: '1px solid var(--border)', background: 'var(--bg-card)' }}
    >
      <article style={{ padding: '16px 20px' }}>
        <div style={{ display: 'flex', gap: 8, marginBottom: 10, flexWrap: 'wrap' }}>
          <span
            style={{
              padding: '3px 10px',
              borderRadius: 999,
              fontSize: 12,
              fontWeight: 600,
              background: event.type === 'official' ? 'var(--accent-soft)' : 'var(--bg-subtle)',
              color: event.type === 'official' ? 'var(--accent)' : 'var(--text-secondary)',
              border: `1px solid ${event.type === 'official' ? 'var(--accent-soft)' : 'var(--border)'}`,
            }}
          >
            {event.type === 'official' ? 'Official' : 'Unofficial'}
          </span>
        </div>

        <h3
          style={{
            fontSize: 17,
            fontWeight: 700,
            color: 'var(--text)',
            marginBottom: 6,
            lineHeight: 1.35,
            letterSpacing: '-0.01em',
          }}
        >
          {event.title}
        </h3>

        <p
          style={{
            fontSize: 14,
            color: 'var(--text-secondary)',
            lineHeight: 1.6,
            marginBottom: 12,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
          }}
        >
          {event.description}
        </p>

        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
          {(event.startTime || event.endTime) && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 5,
                fontSize: 13,
                color: 'var(--text-tertiary)',
              }}
            >
              <CalendarDays size={13} />
              {formatEventTimeRange(event.startTime, event.endTime)}
            </span>
          )}
          {event.location && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 5,
                fontSize: 13,
                color: 'var(--text-tertiary)',
              }}
            >
              <MapPin size={13} />
              {event.location}
            </span>
          )}
        </div>
      </article>
    </Link>
  );
}
