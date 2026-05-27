'use client';

import Link from 'next/link';
import { CalendarDays, MapPin } from 'lucide-react';
import { useTrendingEvents } from '@/hooks/useEventsFeed';

const DEFAULT_VISIBLE_COUNT = 3;
const HEADING = "Today's Events";

interface TrendingEventsProps {
  /** Richer rows: shows start–end range + location. */
  detailed?: boolean;
  /** Flat full-width section (no rounded card chrome) — for in-page use like X. */
  bare?: boolean;
}

function formatEventTime(time: string | null): string {
  if (!time) return '';
  const date = new Date(time);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

function formatEventRange(startTime: string | null, endTime: string | null): string {
  const start = formatEventTime(startTime);
  const end = formatEventTime(endTime);
  if (start && end && start !== end) return `${start} – ${end}`;
  return start || end;
}

export function TrendingEvents({ detailed = false, bare = false }: TrendingEventsProps) {
  const { data, isLoading, isError } = useTrendingEvents();

  const sectionClass = bare ? '' : 'overflow-hidden rounded-2xl';
  const sectionStyle = bare
    ? { borderBottom: '1px solid var(--border)' }
    : { background: 'var(--bg-card)', border: '1px solid var(--border)' };
  const headCls = bare ? 'text-[20px]' : 'text-[18px]';
  const headPad = bare ? '16px 16px 6px' : '14px 16px 8px';

  if (isLoading) {
    return (
      <section className={sectionClass} style={sectionStyle}>
        <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>{HEADING}</h2>
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
    <section className={sectionClass} style={sectionStyle}>
      <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>{HEADING}</h2>

      <div>
        {visible.map((event) => (
          <Link
            key={event.id}
            href={`/events/${event.id}`}
            className="flex items-start gap-3 transition-colors"
            style={{ padding: '12px 16px', color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0 flex-1">
              <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>
                {event.type === 'official' ? 'Official' : 'Campus'} · Event
              </p>
              <p className="line-clamp-2 text-[15px] font-bold leading-snug" style={{ marginTop: 2, color: 'var(--text)' }}>
                {event.title}
              </p>
              {detailed ? (
                <>
                  {(event.startTime || event.endTime) && (
                    <p className="flex items-center gap-1.5 text-[13px]" style={{ marginTop: 4, color: 'var(--text-tertiary)' }}>
                      <CalendarDays size={13} style={{ flexShrink: 0 }} />
                      {formatEventRange(event.startTime, event.endTime)}
                    </p>
                  )}
                  {event.location && (
                    <p className="flex items-center gap-1.5 text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
                      <MapPin size={13} style={{ flexShrink: 0 }} />
                      <span className="truncate">{event.location}</span>
                    </p>
                  )}
                </>
              ) : (
                event.startTime && (
                  <p className="text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
                    {formatEventTime(event.startTime)}
                  </p>
                )
              )}
            </div>

            {event.banner ? (
              <img src={event.banner} alt="" className="h-14 w-14 shrink-0 rounded-xl object-cover" />
            ) : (
              <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl" style={{ background: 'var(--bg-subtle)' }}>
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
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
      >
        Show more
      </Link>
    </section>
  );
}
