'use client';

import Link from 'next/link';
import { CalendarDays } from 'lucide-react';
import { useEventsFeed } from '@/hooks/useEventsFeed';

const DEFAULT_VISIBLE_COUNT = 3;

function formatEventTime(startTime: string | null): string {
  if (!startTime) return '';
  const date = new Date(startTime);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function TrendingEvents() {
  const { data, isLoading, isError } = useEventsFeed();

  if (isLoading) {
    return (
      <section
        className="overflow-hidden rounded-2xl"
        style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
      >
        <h2 className="px-4 pb-1 pt-4 text-xl font-extrabold" style={{ color: 'var(--text)' }}>Events</h2>
        <div>
          {Array.from({ length: DEFAULT_VISIBLE_COUNT }, (_, i) => (
            <div key={i} className="flex items-start gap-3 px-4 py-3">
              <div className="flex-1 space-y-1.5">
                <div className="h-3 w-24 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
                <div className="h-4 w-full animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
                <div className="h-3 w-20 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
              </div>
              <div className="h-16 w-16 shrink-0 animate-pulse rounded-xl" style={{ background: 'var(--bg-subtle)' }} />
            </div>
          ))}
        </div>
      </section>
    );
  }

  if (isError) return null;

  const events = data?.trending ?? [];
  if (events.length === 0) return null;

  const visible = events.slice(0, DEFAULT_VISIBLE_COUNT);

  return (
    <section
      className="overflow-hidden rounded-2xl"
      style={{ background: 'var(--bg-card)', border: '1px solid var(--border)' }}
    >
      <h2 className="px-4 pb-1 pt-4 text-xl font-extrabold" style={{ color: 'var(--text)' }}>Events</h2>

      <div>
        {visible.map((event) => (
          <Link
            key={event.id}
            href={`/events/${event.id}`}
            className="flex items-start gap-3 px-4 py-3 transition-colors"
            style={{ color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0 flex-1">
              <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>
                {event.type === 'official' ? 'Official' : 'Campus'} · Event
              </p>
              <p className="mt-0.5 line-clamp-2 text-[15px] font-bold leading-snug" style={{ color: 'var(--text)' }}>
                {event.title}
              </p>
              {event.startTime && (
                <p className="mt-0.5 text-[13px]" style={{ color: 'var(--text-tertiary)' }}>
                  {formatEventTime(event.startTime)}
                </p>
              )}
            </div>

            {event.banner ? (
              <img
                src={event.banner}
                alt=""
                className="h-16 w-16 shrink-0 rounded-xl object-cover"
              />
            ) : (
              <div
                className="flex h-16 w-16 shrink-0 items-center justify-center rounded-xl"
                style={{ background: 'var(--bg-subtle)' }}
              >
                <CalendarDays size={22} style={{ color: 'var(--text-tertiary)' }} />
              </div>
            )}
          </Link>
        ))}
      </div>

      <Link
        href="/events"
        className="block w-full px-4 py-4 text-[15px] transition-colors"
        style={{ color: 'var(--accent)' }}
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
      >
        Show more
      </Link>
    </section>
  );
}
