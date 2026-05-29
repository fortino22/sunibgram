'use client';

import { CalendarDays, ChevronDown, LayoutGrid, List, MapPin, Plus, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import { CreateEventDialog } from '@/components/events/CreateEventDialog';
import { EventCardSkeleton } from '@/components/events/EventCardSkeleton';
import { EventGridCard } from '@/components/events/EventGridCard';
import { EmptyState } from '@/components/shared/EmptyState';
import { useEventsFeed } from '@/hooks/useEventsFeed';
import { useMyEvents } from '@/hooks/useMyEvents';
import { useMyJoinedEventItems } from '@/hooks/useMyJoinedEventItems';
import type { EventItem, MyEventItem } from '@/lib/services/events';

// ── Types ───────────────────────────────────────────────────────────────────

type Tab = 'trending' | 'running' | 'future';
type Scope = 'all' | 'joined' | 'mine';
type ViewMode = 'grid' | 'list';
type DatePreset = 'any' | 'today' | 'week' | 'month' | 'quarter' | 'custom';

const TABS: { key: Tab; label: string }[] = [
  { key: 'trending', label: 'Trending' },
  { key: 'running',  label: 'Running' },
  { key: 'future',   label: 'Future' },
];

const SCOPE_OPTIONS: { key: Scope; label: string }[] = [
  { key: 'all',    label: 'All Events' },
  { key: 'joined', label: 'Joined' },
  { key: 'mine',   label: 'My Events' },
];

const DATE_PRESETS: { key: DatePreset; label: string }[] = [
  { key: 'any',     label: 'Any date' },
  { key: 'today',   label: 'Today' },
  { key: 'week',    label: 'This Week' },
  { key: 'month',   label: 'This Month' },
  { key: 'quarter', label: 'Next 3 Months' },
  { key: 'custom',  label: 'Custom range…' },
];

const VIEW_STORAGE_KEY = 'events:viewMode';

// ── Helpers ─────────────────────────────────────────────────────────────────

function parseDate(iso: string | null): Date | null {
  if (!iso) return null;
  const d = new Date(iso.replace(' ', 'T'));
  return Number.isNaN(d.getTime()) ? null : d;
}

function matchesSearch(event: EventItem, q: string) {
  if (!q.trim()) return true;
  const lower = q.toLowerCase();
  return (
    event.title.toLowerCase().includes(lower) ||
    (event.description ?? '').toLowerCase().includes(lower) ||
    (event.location ?? '').toLowerCase().includes(lower)
  );
}

function startOfDay(d: Date) { const x = new Date(d); x.setHours(0, 0, 0, 0); return x; }
function endOfDay(d: Date)   { const x = new Date(d); x.setHours(23, 59, 59, 999); return x; }

function presetRange(preset: DatePreset, custom: { from: string; to: string }): { from: Date | null; to: Date | null } {
  const now = new Date();
  if (preset === 'any') return { from: null, to: null };
  if (preset === 'today') return { from: startOfDay(now), to: endOfDay(now) };
  if (preset === 'week') {
    const day = now.getDay() === 0 ? 6 : now.getDay() - 1; // Monday-start
    const from = startOfDay(new Date(now.getFullYear(), now.getMonth(), now.getDate() - day));
    const to = endOfDay(new Date(from.getFullYear(), from.getMonth(), from.getDate() + 6));
    return { from, to };
  }
  if (preset === 'month') {
    const from = startOfDay(new Date(now.getFullYear(), now.getMonth(), 1));
    const to = endOfDay(new Date(now.getFullYear(), now.getMonth() + 1, 0));
    return { from, to };
  }
  if (preset === 'quarter') {
    const from = startOfDay(now);
    const to = endOfDay(new Date(now.getFullYear(), now.getMonth() + 3, now.getDate()));
    return { from, to };
  }
  // custom
  const f = custom.from ? startOfDay(new Date(custom.from)) : null;
  const t = custom.to ? endOfDay(new Date(custom.to)) : null;
  return { from: f && !Number.isNaN(f.getTime()) ? f : null, to: t && !Number.isNaN(t.getTime()) ? t : null };
}

function inDateRange(event: EventItem, range: { from: Date | null; to: Date | null }) {
  if (!range.from && !range.to) return true;
  const start = parseDate(event.startTime);
  if (!start) return false;
  if (range.from && start < range.from) return false;
  if (range.to && start > range.to) return false;
  return true;
}

function statusBadge(status: string): { label: string; tone: 'success' | 'warning' | 'danger' | 'info' | 'default' } {
  const m: Record<string, { label: string; tone: 'success' | 'warning' | 'danger' | 'info' | 'default' }> = {
    approved: { label: 'Approved', tone: 'success' },
    pending:  { label: 'Pending',  tone: 'warning' },
    declined: { label: 'Declined', tone: 'danger' },
    revision: { label: 'Revision', tone: 'info' },
  };
  return m[status] ?? { label: status, tone: 'default' };
}

function formatRange(startIso: string | null, endIso: string | null): string {
  const start = parseDate(startIso);
  const end = parseDate(endIso);
  if (start && end) {
    const sameDay = start.toDateString() === end.toDateString();
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
  return one ? one.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '';
}

// ── Revision notes disclosure (My Events) ───────────────────────────────────

function RevisionNotesDisclosure({ event }: { event: MyEventItem }) {
  const [open, setOpen] = useState(false);
  if (!event.notes.length) return null;
  return (
    <div style={{ marginTop: 8 }}>
      <button
        type="button"
        onClick={(e) => { e.preventDefault(); e.stopPropagation(); setOpen((o) => !o); }}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          width: '100%',
          padding: '8px 12px',
          background: 'var(--bg-secondary)',
          border: '1px solid var(--border)',
          borderRadius: 10,
          cursor: 'pointer',
          color: 'var(--text-tertiary)',
          fontSize: 12,
          fontWeight: 600,
          transition: 'background 0.15s ease, border-color 0.15s ease',
        }}
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--text)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--border)'; }}
      >
        <ChevronDown
          size={14}
          style={{
            transition: 'transform 0.15s ease',
            transform: open ? 'rotate(180deg)' : 'rotate(0deg)',
          }}
        />
        Revision notes ({event.notes.length})
      </button>
      {open && (
        <div
          style={{
            marginTop: 6,
            padding: '12px 14px',
            borderRadius: 10,
            background: 'var(--bg)',
            border: '1px solid var(--border)',
            display: 'flex',
            flexDirection: 'column',
            gap: 10,
          }}
        >
          {event.notes.map((note) => (
            <div key={note.revisionNumber} style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--coral, #f87171)' }}>
                Revision #{note.revisionNumber}
              </div>
              <div style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.4 }}>
                {note.notes}
              </div>
              <div style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                {new Date(note.createdAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ── List view row card ──────────────────────────────────────────────────────

interface ListRowProps {
  event: EventItem;
  live?: boolean;
  badge?: { label: string; tone: 'success' | 'warning' | 'danger' | 'info' | 'default' };
}

function EventListRow({ event, live = false, badge }: ListRowProps) {
  const [hovered, setHovered] = useState(false);
  const isOfficial = event.type === 'official';
  const dateLabel = formatRange(event.startTime, event.endTime);

  const tones: Record<string, { bg: string; fg: string; border: string }> = {
    default: { bg: 'oklch(100% 0 0 / 0.06)', fg: 'var(--text)',         border: 'oklch(100% 0 0 / 0.15)' },
    success: { bg: 'oklch(70% 0.15 145 / 0.15)', fg: 'oklch(82% 0.15 145)', border: 'oklch(70% 0.15 145 / 0.4)' },
    warning: { bg: 'oklch(75% 0.15 80 / 0.15)',  fg: 'oklch(85% 0.15 80)',  border: 'oklch(75% 0.15 80 / 0.4)' },
    danger:  { bg: 'oklch(65% 0.2 25 / 0.15)',   fg: 'oklch(75% 0.2 25)',   border: 'oklch(65% 0.2 25 / 0.4)' },
    info:    { bg: 'oklch(70% 0.15 230 / 0.15)', fg: 'oklch(80% 0.15 230)', border: 'oklch(70% 0.15 230 / 0.4)' },
  };
  const tone = badge ? tones[badge.tone] : null;

  return (
    <Link
      href={`/events/${event.id}`}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        padding: 14,
        borderRadius: 16,
        border: '1px solid',
        borderColor: hovered ? 'var(--text)' : 'var(--border)',
        background: 'var(--bg-secondary)',
        textDecoration: 'none',
        transition: 'transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease',
        transform: hovered ? 'translateY(-2px)' : 'translateY(0)',
        boxShadow: hovered ? '0 12px 24px oklch(0% 0 0 / 0.45)' : '0 0 0 oklch(0% 0 0 / 0)',
      }}
    >
      {/* Banner */}
      <div
        style={{
          width: 140,
          height: 96,
          flexShrink: 0,
          borderRadius: 12,
          background: event.banner
            ? `url(${event.banner}) center/cover no-repeat`
            : 'linear-gradient(135deg, oklch(20% 0.01 260), oklch(14% 0.005 260))',
          border: '1px solid var(--border)',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {!event.banner && (
          <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-tertiary)' }}>
            <CalendarDays size={26} />
          </div>
        )}
        {live && (
          <span
            style={{
              position: 'absolute',
              top: 8,
              left: 8,
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
              fontSize: 10,
              fontWeight: 700,
              letterSpacing: '0.06em',
              textTransform: 'uppercase',
              color: 'var(--text)',
              background: 'oklch(0% 0 0 / 0.6)',
              backdropFilter: 'blur(6px)',
              border: '1px solid var(--green)',
              padding: '3px 7px',
              borderRadius: 999,
            }}
          >
            <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'var(--green)', animation: 'pulse 1.6s ease-in-out infinite' }} />
            Live
          </span>
        )}
      </div>

      {/* Info */}
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 6 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.01em' }}>
            {event.title}
          </span>
          <span
            style={{
              fontSize: 10,
              fontWeight: 700,
              letterSpacing: '0.06em',
              textTransform: 'uppercase',
              color: isOfficial ? 'var(--bg)' : 'var(--text-secondary)',
              background: isOfficial ? 'var(--text)' : 'transparent',
              border: isOfficial ? 'none' : '1px solid var(--border)',
              padding: '3px 8px',
              borderRadius: 999,
            }}
          >
            {isOfficial ? 'Official' : 'Unofficial'}
          </span>
        </div>

        {dateLabel && (
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-secondary)' }}>
            <CalendarDays size={13} />
            {dateLabel}
          </span>
        )}

        {event.location && (
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-tertiary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            <MapPin size={13} style={{ flexShrink: 0 }} />
            <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{event.location}</span>
          </span>
        )}
      </div>

      {/* Badge */}
      {badge && tone && (
        <span
          style={{
            flexShrink: 0,
            fontSize: 11,
            fontWeight: 700,
            letterSpacing: '0.05em',
            textTransform: 'uppercase',
            color: tone.fg,
            background: tone.bg,
            border: `1px solid ${tone.border}`,
            padding: '5px 10px',
            borderRadius: 999,
          }}
        >
          {badge.label}
        </span>
      )}
    </Link>
  );
}

// ── Main panel ──────────────────────────────────────────────────────────────

interface PanelProps { active: boolean }

export function EventsFeedPanel({ active }: PanelProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [tab, setTab] = useState<Tab>('trending');
  const [scope, setScope] = useState<Scope>('all');
  const [scopeOpen, setScopeOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [view, setView] = useState<ViewMode>(() => {
    if (typeof window === 'undefined') return 'grid';
    try {
      const stored = window.localStorage.getItem(VIEW_STORAGE_KEY);
      if (stored === 'grid' || stored === 'list') return stored;
    } catch {}
    return 'grid';
  });
  const [datePreset, setDatePreset] = useState<DatePreset>('any');
  const [dateOpen, setDateOpen] = useState(false);
  const [customRange, setCustomRange] = useState<{ from: string; to: string }>({ from: '', to: '' });

  const scopeRef = useRef<HTMLDivElement>(null);
  const dateRef = useRef<HTMLDivElement>(null);

  const query = useEventsFeed(active);
  const joinedQuery = useMyJoinedEventItems();
  const myEventsQuery = useMyEvents();

  useEffect(() => {
    try { window.localStorage.setItem(VIEW_STORAGE_KEY, view); } catch {}
  }, [view]);

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (scopeRef.current && !scopeRef.current.contains(e.target as Node)) setScopeOpen(false);
      if (dateRef.current  && !dateRef.current.contains(e.target as Node))  setDateOpen(false);
    }
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const now = useMemo(() => new Date(), []);
  const range = useMemo(() => presetRange(datePreset, customRange), [datePreset, customRange]);

  const passDate = (e: EventItem) => inDateRange(e, range);
  const passSearch = (e: EventItem) => matchesSearch(e, searchQuery);

  // Time-state predicates (mainly for joined / mine since query.data is bucketed already)
  const isRunningNow = (e: EventItem) => {
    const s = parseDate(e.startTime); const en = parseDate(e.endTime);
    return !!(s && en && s <= now && en >= now);
  };
  const isFuture = (e: EventItem) => {
    const s = parseDate(e.startTime);
    return !!(s && s > now);
  };

  const baseAll: EventItem[] = useMemo(() => {
    const d = query.data;
    if (!d) return [];
    if (tab === 'trending') return d.trending;
    if (tab === 'running')  return d.running;
    return d.future;
  }, [query.data, tab]);

  const allList = baseAll.filter(passDate).filter(passSearch);

  const joinedList = (joinedQuery.data ?? [])
    .filter((e) => (tab === 'trending' ? true : tab === 'running' ? isRunningNow(e) : isFuture(e)))
    .filter(passDate)
    .filter(passSearch);

  const myList: MyEventItem[] = (myEventsQuery.data ?? [])
    .filter((e) => (tab === 'trending' ? true : tab === 'running' ? isRunningNow(e) : isFuture(e)))
    .filter(passDate)
    .filter(passSearch);

  const currentScopeLabel = SCOPE_OPTIONS.find((o) => o.key === scope)?.label ?? 'All Events';
  const currentDateLabel = (() => {
    if (datePreset === 'custom' && (customRange.from || customRange.to)) {
      const f = customRange.from ? new Date(customRange.from).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : '…';
      const t = customRange.to   ? new Date(customRange.to).toLocaleDateString('en-US',   { month: 'short', day: 'numeric' }) : '…';
      return `${f} – ${t}`;
    }
    return DATE_PRESETS.find((p) => p.key === datePreset)?.label ?? 'Any date';
  })();

  const loading =
    (scope === 'all'    && query.isLoading) ||
    (scope === 'joined' && joinedQuery.isLoading) ||
    (scope === 'mine'   && myEventsQuery.isLoading);

  const errorMsg =
    (scope === 'all'    && query.error?.message) ||
    (scope === 'joined' && joinedQuery.error?.message) ||
    (scope === 'mine'   && myEventsQuery.error?.message) ||
    null;

  const liveFlag = tab === 'running';

  // ── Render helpers ─────────────────────────────────────────────────────

  const renderEvents = (list: EventItem[]) => {
    if (view === 'grid') {
      return (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 20 }}>
          {list.map((e) => <EventGridCard key={e.id} event={e} live={liveFlag} />)}
        </div>
      );
    }
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {list.map((e) => <EventListRow key={e.id} event={e} live={liveFlag} />)}
      </div>
    );
  };

  const renderMine = (list: MyEventItem[]) => {
    if (view === 'grid') {
      return (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 20 }}>
          {list.map((e) => (
            <div key={e.id} style={{ display: 'flex', flexDirection: 'column' }}>
              <EventGridCard event={e} badge={statusBadge(e.status)} />
              <RevisionNotesDisclosure event={e} />
            </div>
          ))}
        </div>
      );
    }
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {list.map((e) => (
          <div key={e.id}>
            <EventListRow event={e} badge={statusBadge(e.status)} />
            <RevisionNotesDisclosure event={e} />
          </div>
        ))}
      </div>
    );
  };

  const skeletonView = view === 'grid' ? (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 20 }}>
      {Array.from({ length: 6 }).map((_, i) => <EventCardSkeleton key={i} />)}
    </div>
  ) : (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="animate-pulse" style={{
          display: 'flex', gap: 16, padding: 14, borderRadius: 16,
          border: '1px solid var(--border)', background: 'var(--bg-secondary)',
        }}>
          <div style={{ width: 140, height: 96, borderRadius: 12, background: 'var(--bg-subtle)', flexShrink: 0 }} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8, justifyContent: 'center' }}>
            <div className="h-4 w-1/2 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3 w-1/3 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3 w-1/4 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
          </div>
        </div>
      ))}
    </div>
  );

  return (
    <>
      {/* ── Sticky header ─────────────────────────────────────────── */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(8% 0.005 260 / 0.85)',
          backdropFilter: 'blur(16px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        {/* Title row */}
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '20px 32px 16px', gap: 12,
        }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <div style={{ fontWeight: 800, fontSize: 24, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              Events
            </div>
            <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              Discover what&apos;s happening on campus
            </div>
          </div>

          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <button
              type="button"
              aria-label="Search events"
              onClick={() => { setSearchOpen((o) => !o); setSearchQuery(''); }}
              style={{
                width: 40, height: 40, borderRadius: '50%',
                border: `1px solid ${searchOpen ? 'var(--text)' : 'var(--border)'}`,
                background: searchOpen ? 'var(--text)' : 'transparent',
                color: searchOpen ? 'var(--bg)' : 'var(--text)',
                cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
                transition: 'background 0.15s, border-color 0.15s, transform 0.15s',
              }}
              onMouseEnter={(e) => { if (!searchOpen) (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; (e.currentTarget as HTMLElement).style.transform = 'scale(1.05)'; }}
              onMouseLeave={(e) => { if (!searchOpen) (e.currentTarget as HTMLElement).style.background = 'transparent'; (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
            >
              <Search size={16} />
            </button>

            <button
              type="button"
              onClick={() => setDialogOpen(true)}
              style={{
                padding: '0 20px', height: 40, borderRadius: 999, border: 'none',
                background: 'var(--text)', color: 'var(--bg)',
                fontSize: 14, fontWeight: 700, letterSpacing: '-0.01em',
                cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 6,
                transition: 'opacity 0.15s, transform 0.15s',
              }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.opacity = '0.88'; (e.currentTarget as HTMLElement).style.transform = 'scale(1.03)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.opacity = '1'; (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
            >
              <Plus size={16} strokeWidth={2.5} />
              Create
            </button>
          </div>
        </div>

        {/* Search bar */}
        {searchOpen && (
          <div style={{ padding: '0 32px 14px' }}>
            <div style={{ position: 'relative' }}>
              <Search size={15} style={{ position: 'absolute', left: 14, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', pointerEvents: 'none' }} />
              <input
                autoFocus
                type="text"
                placeholder="Search events, location…"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                style={{
                  width: '100%', padding: '11px 40px', borderRadius: 999,
                  border: '1px solid var(--border)', background: 'var(--bg-secondary)',
                  color: 'var(--text)', fontSize: 14, outline: 'none', boxSizing: 'border-box',
                  transition: 'border-color 0.15s',
                }}
                onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--text)'; }}
                onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border)'; }}
              />
              {searchQuery && (
                <button type="button" onClick={() => setSearchQuery('')} style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', display: 'flex', padding: 0 }}>
                  <X size={14} />
                </button>
              )}
            </div>
          </div>
        )}

        {/* Tabs + filter row */}
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          gap: 12, padding: '0 32px 16px', flexWrap: 'wrap',
        }}>
          {/* Tabs */}
          <div style={{ display: 'flex', gap: 6, background: 'var(--bg-secondary)', padding: 4, borderRadius: 999, border: '1px solid var(--border)' }}>
            {TABS.map(({ key, label }) => {
              const isActive = tab === key;
              return (
                <button
                  key={key}
                  type="button"
                  onClick={() => setTab(key)}
                  style={{
                    padding: '8px 20px', borderRadius: 999, border: 'none',
                    background: isActive ? 'var(--text)' : 'transparent',
                    color: isActive ? 'var(--bg)' : 'var(--text-secondary)',
                    fontSize: 13, fontWeight: 700, letterSpacing: '-0.01em',
                    cursor: 'pointer', transition: 'background 0.18s, color 0.18s',
                  }}
                  onMouseEnter={(e) => { if (!isActive) (e.currentTarget as HTMLElement).style.color = 'var(--text)'; }}
                  onMouseLeave={(e) => { if (!isActive) (e.currentTarget as HTMLElement).style.color = 'var(--text-secondary)'; }}
                >
                  {label}
                </button>
              );
            })}
          </div>

          {/* Right cluster: date + scope + view toggle */}
          <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
            {/* Date filter */}
            <div style={{ position: 'relative' }} ref={dateRef}>
              <button
                type="button"
                onClick={() => setDateOpen((o) => !o)}
                style={{
                  display: 'inline-flex', alignItems: 'center', gap: 6,
                  height: 38, padding: '0 14px',
                  borderRadius: 999,
                  border: `1px solid ${dateOpen ? 'var(--text)' : 'var(--border)'}`,
                  background: dateOpen || datePreset !== 'any' ? 'var(--bg-subtle)' : 'transparent',
                  color: 'var(--text)', fontSize: 13, fontWeight: 600,
                  cursor: 'pointer', transition: 'background 0.15s, border-color 0.15s',
                }}
                onMouseEnter={(e) => { if (!dateOpen) (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
                onMouseLeave={(e) => { if (!dateOpen && datePreset === 'any') (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
              >
                <CalendarDays size={14} />
                {currentDateLabel}
                <ChevronDown size={14} style={{ transition: 'transform 0.15s', transform: dateOpen ? 'rotate(180deg)' : 'rotate(0deg)' }} />
              </button>

              {dateOpen && (
                <div style={{
                  position: 'absolute', top: 'calc(100% + 6px)', right: 0, zIndex: 50,
                  background: 'var(--bg)', border: '1px solid var(--border)', borderRadius: 16,
                  padding: 8, minWidth: 240,
                  boxShadow: '0 16px 40px oklch(0% 0 0 / 0.65)',
                }}>
                  {DATE_PRESETS.map(({ key, label }) => (
                    <button
                      key={key}
                      type="button"
                      onClick={() => {
                        setDatePreset(key);
                        if (key !== 'custom') setDateOpen(false);
                      }}
                      style={{
                        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                        width: '100%', textAlign: 'left',
                        padding: '10px 12px', borderRadius: 10, border: 'none',
                        background: datePreset === key ? 'var(--bg-subtle)' : 'transparent',
                        color: datePreset === key ? 'var(--text)' : 'var(--text-secondary)',
                        fontSize: 13, fontWeight: datePreset === key ? 700 : 500,
                        cursor: 'pointer', transition: 'background 0.1s',
                      }}
                      onMouseEnter={(e) => { if (datePreset !== key) (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
                      onMouseLeave={(e) => { if (datePreset !== key) (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                    >
                      <span>{label}</span>
                      {datePreset === key && (
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      )}
                    </button>
                  ))}

                  {datePreset === 'custom' && (
                    <div style={{
                      marginTop: 8, padding: 12, borderRadius: 10,
                      background: 'var(--bg-secondary)', border: '1px solid var(--border)',
                      display: 'flex', flexDirection: 'column', gap: 8,
                    }}>
                      <label style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                        <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>From</span>
                        <input
                          type="date"
                          value={customRange.from}
                          onChange={(e) => setCustomRange((r) => ({ ...r, from: e.target.value }))}
                          style={{
                            padding: '8px 10px', borderRadius: 8,
                            border: '1px solid var(--border)', background: 'var(--bg)',
                            color: 'var(--text)', fontSize: 13, outline: 'none',
                            colorScheme: 'dark',
                          }}
                        />
                      </label>
                      <label style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                        <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>To</span>
                        <input
                          type="date"
                          value={customRange.to}
                          onChange={(e) => setCustomRange((r) => ({ ...r, to: e.target.value }))}
                          style={{
                            padding: '8px 10px', borderRadius: 8,
                            border: '1px solid var(--border)', background: 'var(--bg)',
                            color: 'var(--text)', fontSize: 13, outline: 'none',
                            colorScheme: 'dark',
                          }}
                        />
                      </label>
                      <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
                        <button
                          type="button"
                          onClick={() => { setCustomRange({ from: '', to: '' }); setDatePreset('any'); setDateOpen(false); }}
                          style={{
                            flex: 1, padding: '8px', borderRadius: 8,
                            border: '1px solid var(--border)', background: 'transparent',
                            color: 'var(--text-secondary)', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                          }}
                        >Clear</button>
                        <button
                          type="button"
                          onClick={() => setDateOpen(false)}
                          style={{
                            flex: 1, padding: '8px', borderRadius: 8,
                            border: 'none', background: 'var(--text)',
                            color: 'var(--bg)', fontSize: 12, fontWeight: 700, cursor: 'pointer',
                          }}
                        >Apply</button>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>

            {/* Scope dropdown */}
            <div style={{ position: 'relative' }} ref={scopeRef}>
              <button
                type="button"
                onClick={() => setScopeOpen((o) => !o)}
                style={{
                  display: 'inline-flex', alignItems: 'center', gap: 6,
                  height: 38, padding: '0 14px', minWidth: 140, justifyContent: 'space-between',
                  borderRadius: 999,
                  border: `1px solid ${scopeOpen ? 'var(--text)' : 'var(--border)'}`,
                  background: scopeOpen ? 'var(--bg-subtle)' : 'transparent',
                  color: 'var(--text)', fontSize: 13, fontWeight: 600,
                  cursor: 'pointer', transition: 'background 0.15s, border-color 0.15s',
                }}
                onMouseEnter={(e) => { if (!scopeOpen) (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
                onMouseLeave={(e) => { if (!scopeOpen) (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
              >
                {currentScopeLabel}
                <ChevronDown size={14} style={{ transition: 'transform 0.15s', transform: scopeOpen ? 'rotate(180deg)' : 'rotate(0deg)' }} />
              </button>

              {scopeOpen && (
                <div style={{
                  position: 'absolute', top: 'calc(100% + 6px)', right: 0, zIndex: 50,
                  background: 'var(--bg)', border: '1px solid var(--border)', borderRadius: 16,
                  padding: 8, minWidth: 180,
                  boxShadow: '0 16px 40px oklch(0% 0 0 / 0.65)',
                }}>
                  {SCOPE_OPTIONS.map(({ key, label }) => (
                    <button
                      key={key}
                      type="button"
                      onClick={() => { setScope(key); setScopeOpen(false); }}
                      style={{
                        display: 'block', width: '100%', textAlign: 'left',
                        padding: '10px 12px', borderRadius: 10, border: 'none',
                        background: scope === key ? 'var(--bg-subtle)' : 'transparent',
                        color: scope === key ? 'var(--text)' : 'var(--text-secondary)',
                        fontSize: 13, fontWeight: scope === key ? 700 : 500,
                        cursor: 'pointer', transition: 'background 0.1s',
                      }}
                      onMouseEnter={(e) => { if (scope !== key) (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
                      onMouseLeave={(e) => { if (scope !== key) (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* View toggle */}
            <div style={{
              display: 'flex', gap: 2, padding: 4, borderRadius: 999,
              background: 'var(--bg-secondary)', border: '1px solid var(--border)',
            }}>
              <button
                type="button"
                aria-label="Grid view"
                onClick={() => setView('grid')}
                style={{
                  width: 32, height: 30, borderRadius: 999, border: 'none',
                  background: view === 'grid' ? 'var(--text)' : 'transparent',
                  color: view === 'grid' ? 'var(--bg)' : 'var(--text-secondary)',
                  cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  transition: 'background 0.18s, color 0.18s',
                }}
              >
                <LayoutGrid size={14} />
              </button>
              <button
                type="button"
                aria-label="List view"
                onClick={() => setView('list')}
                style={{
                  width: 32, height: 30, borderRadius: 999, border: 'none',
                  background: view === 'list' ? 'var(--text)' : 'transparent',
                  color: view === 'list' ? 'var(--bg)' : 'var(--text-secondary)',
                  cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  transition: 'background 0.18s, color 0.18s',
                }}
              >
                <List size={14} />
              </button>
            </div>
          </div>
        </div>
      </div>

      {dialogOpen && <CreateEventDialog open={dialogOpen} onOpenChange={setDialogOpen} />}

      {/* ── Content ─────────────────────────────────────────────── */}
      <div style={{ padding: '24px 32px 48px' }}>
        {loading ? (
          skeletonView
        ) : errorMsg ? (
          <div style={{ padding: '40px 20px', fontSize: 14, color: 'var(--coral)' }}>{errorMsg}</div>
        ) : scope === 'mine' ? (
          myList.length === 0 ? (
            <div style={{ padding: '60px 20px' }}>
              <EmptyState title="No events here" subtitle={`No ${tab} events submitted by you yet.`} />
            </div>
          ) : renderMine(myList)
        ) : scope === 'joined' ? (
          joinedList.length === 0 ? (
            <div style={{ padding: '60px 20px' }}>
              <EmptyState title="No joined events" subtitle={`Join an event to see ${tab} ones here.`} />
            </div>
          ) : renderEvents(joinedList)
        ) : allList.length === 0 ? (
          <div style={{ padding: '60px 20px' }}>
            <EmptyState
              title={tab === 'trending' ? 'No trending events' : tab === 'running' ? 'No events running now' : 'No future events'}
              subtitle={tab === 'trending' ? 'Check back soon for popular events.' : tab === 'running' ? 'Nothing live right now.' : 'Create one to get started.'}
            />
          </div>
        ) : renderEvents(allList)}
      </div>
    </>
  );
}
