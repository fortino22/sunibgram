'use client';

import Link from 'next/link';
import { useState } from 'react';
import { CalendarDays, MapPin } from 'lucide-react';
import type { EventItem } from '@/lib/services/events';

function parseEventDate(iso: string | null): Date | null {
  if (!iso) return null;
  const date = new Date(iso.replace(' ', 'T'));
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatRange(startIso: string | null, endIso: string | null): string {
  const start = parseEventDate(startIso);
  const end = parseEventDate(endIso);

  if (start && end) {
    const sameDay =
      start.getFullYear() === end.getFullYear() &&
      start.getMonth() === end.getMonth() &&
      start.getDate() === end.getDate();
    if (sameDay) {
      const d = start.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
      const s = start.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
      const e = end.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
      return `${d}, ${s} – ${e}`;
    }
    const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
    return `${start.toLocaleDateString('en-US', opts)} – ${end.toLocaleDateString('en-US', opts)}`;
  }
  const one = start ?? end;
  if (one) return one.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  return '';
}

interface Props {
  event: EventItem;
  live?: boolean;
  badge?: { label: string; tone: 'default' | 'success' | 'warning' | 'danger' | 'info' };
}

export function EventGridCard({ event, live = false, badge }: Props) {
  const [hovered, setHovered] = useState(false);
  const isOfficial = event.type === 'official';
  const dateLabel = formatRange(event.startTime, event.endTime);

  const badgeTones: Record<string, { bg: string; fg: string; border: string }> = {
    default: { bg: 'var(--bg-hover)', fg: 'var(--text)',         border: 'var(--border)' },
    success: { bg: 'oklch(70% 0.15 145 / 0.18)', fg: 'oklch(82% 0.15 145)', border: 'oklch(70% 0.15 145 / 0.4)' },
    warning: { bg: 'oklch(75% 0.15 80 / 0.18)',  fg: 'oklch(85% 0.15 80)',  border: 'oklch(75% 0.15 80 / 0.4)' },
    danger:  { bg: 'oklch(65% 0.2 25 / 0.18)',   fg: 'oklch(75% 0.2 25)',   border: 'oklch(65% 0.2 25 / 0.4)' },
    info:    { bg: 'oklch(70% 0.15 230 / 0.18)', fg: 'oklch(80% 0.15 230)', border: 'oklch(70% 0.15 230 / 0.4)' },
  };
  const tone = badge ? badgeTones[badge.tone] : null;

  return (
    <Link
      href={`/events/${event.id}`}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'flex',
        flexDirection: 'column',
        textDecoration: 'none',
        borderRadius: 16,
        border: '1px solid var(--border)',
        background: 'var(--bg-secondary)',
        overflow: 'hidden',
        transition: 'transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease',
        transform: hovered ? 'translateY(-3px)' : 'translateY(0)',
        borderColor: hovered ? 'var(--text)' : 'var(--border)',
        boxShadow: hovered ? 'var(--shadow-dropdown)' : '0 0 0 oklch(0% 0 0 / 0)',
      }}
    >
      {/* Banner */}
      <div
        style={{
          position: 'relative',
          width: '100%',
          aspectRatio: '16 / 10',
          background: event.banner
            ? `url(${event.banner}) center/cover no-repeat`
            : 'linear-gradient(135deg, oklch(20% 0.01 260), oklch(14% 0.005 260))',
          overflow: 'hidden',
        }}
      >
        {/* Gradient overlay */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            background: 'linear-gradient(180deg, oklch(0% 0 0 / 0) 40%, oklch(0% 0 0 / 0.55) 100%)',
          }}
        />
        {!event.banner && (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'var(--text-tertiary)',
            }}
          >
            <CalendarDays size={36} />
          </div>
        )}
        {/* Top-left type pill */}
        <span
          style={{
            position: 'absolute',
            top: 10,
            left: 10,
            fontSize: 10,
            fontWeight: 700,
            letterSpacing: '0.06em',
            textTransform: 'uppercase',
            color: isOfficial ? 'var(--bg)' : 'var(--text)',
            background: isOfficial ? 'var(--text)' : 'oklch(0% 0 0 / 0.55)',
            backdropFilter: 'blur(6px)',
            border: isOfficial ? 'none' : '1px solid oklch(100% 0 0 / 0.2)',
            padding: '4px 8px',
            borderRadius: 999,
          }}
        >
          {isOfficial ? 'Official' : 'Unofficial'}
        </span>
        {/* Top-right live indicator only (status badge moved to body) */}
        {live && (
          <span
            style={{
              position: 'absolute',
              top: 10,
              right: 10,
              display: 'inline-flex',
              alignItems: 'center',
              gap: 5,
              fontSize: 10,
              fontWeight: 700,
              letterSpacing: '0.06em',
              textTransform: 'uppercase',
              color: 'var(--text)',
              background: 'oklch(0% 0 0 / 0.6)',
              backdropFilter: 'blur(6px)',
              border: '1px solid var(--green)',
              padding: '4px 8px',
              borderRadius: 999,
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

      {/* Body */}
      <div style={{ padding: '14px 14px 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
        {/* Status badge row (top of body, right-aligned) */}
        {badge && tone && (
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <span
              style={{
                fontSize: 10,
                fontWeight: 700,
                letterSpacing: '0.06em',
                textTransform: 'uppercase',
                color: tone.fg,
                background: tone.bg,
                border: `1px solid ${tone.border}`,
                padding: '4px 9px',
                borderRadius: 999,
              }}
            >
              {badge.label}
            </span>
          </div>
        )}
        {dateLabel && (
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 12,
              fontWeight: 600,
              color: 'var(--text-tertiary)',
              letterSpacing: '0.01em',
            }}
          >
            <CalendarDays size={13} />
            {dateLabel}
          </div>
        )}
        <div
          style={{
            fontSize: 16,
            fontWeight: 700,
            color: 'var(--text)',
            letterSpacing: '-0.01em',
            lineHeight: 1.25,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
          }}
        >
          {event.title}
        </div>
        {event.location && (
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 13,
              color: 'var(--text-secondary)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            <MapPin size={13} style={{ flexShrink: 0 }} />
            <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{event.location}</span>
          </div>
        )}
      </div>
    </Link>
  );
}
