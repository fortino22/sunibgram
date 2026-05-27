'use client';

import Link from 'next/link';
import { CalendarDays } from 'lucide-react';
import { useTrendingEvents } from '@/hooks/useEventsFeed';

const DEFAULT_VISIBLE_COUNT = 3;
const HEADING = "Today's Events";

function formatEventTime(startTime: string | null): string {
  if (!startTime) return '';
  const date = new Date(startTime);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function TrendingEvents() {
  const { data, isLoading, isError } = useTrendingEvents();

  if (isLoading) {
    return (
      <section
        className="overflow-hidden rounded-2xl"
        style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
      >
        <h2 className="text-[18px] font-extrabold" style={{ padding: '14px 16px 8px', color: 'var(--text)' }}>{HEADING}</h2>
        <div>
          {Array.from({ length: DEFAULT_VISIBLE_COUNT }, (_, i) => (
            <div key={i} className="flex items-start gap-3" style={{ padding: '12px 16px' }}>
              <div className="flex-1" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                <div className="h-2.5 w-20 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
                <div className="h-3.5 w-full animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
                <div className="h-2.5 w-16 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
              </div>
              <div className="h-14 w-14 shrink-0 animate-pulse rounded-xl" style={{ background: 'var(--bg-subtle)' }} />
            </div>
          ))}
        </div>
      </section>
    );
  }

  if (isError) return null;

  const events = data ?? [];
  if (events.length === 0) return null;

  const visible = events.slice(0, DEFAULT_VISIBLE_COUNT);

  return (
    <section
      className="overflow-hidden rounded-2xl"
      style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
    >
      <h2 className="text-[18px] font-extrabold" style={{ padding: '14px 16px 8px', color: 'var(--text)' }}>{HEADING}</h2>

      <div>
        {visible.map((event) => (
          <Link
            key={event.id}
            href={`/events/${event.id}`}
            className="flex items-start gap-3 transition-colors"
            style={{ padding: '12px 16px', color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0 flex-1">
              <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>
                {event.type === 'official' ? 'Official' : 'Campus'} · Event
              </p>
              <p className="line-clamp-2 text-[15px] font-bold leading-snug" style={{ marginTop: 2, color: 'var(--text)' }}>
                {event.title}
              </p>
              {event.startTime && (
                <p className="text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
                  {formatEventTime(event.startTime)}
                </p>
              )}
            </div>

            {event.banner ? (
              <img
                src={event.banner}
                alt=""
                className="h-14 w-14 shrink-0 rounded-xl object-cover"
              />
            ) : (
              <div
                className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl"
                style={{ background: 'var(--bg-subtle)' }}
              >
                <CalendarDays size={20} style={{ color: 'var(--text-tertiary)' }} />
              </div>
            )}
          </Link>
        ))}
      </div>

      <Link
        href="/events"
        className="block w-full text-[15px] transition-colors"
        style={{ padding: '12px 16px', color: 'var(--accent)' }}
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
      >
        Show more
      </Link>
    </section>
  );
}
