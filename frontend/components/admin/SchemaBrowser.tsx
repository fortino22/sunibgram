'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import {
  Search,
  KeyRound,
  Link2,
  Hash,
  Table2,
  ChevronRight,
  Copy,
  Check,
} from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { getSchema } from '@/lib/services/admin/schema';
import type { SchemaColumn, SchemaTable } from '@/lib/services/admin/types';

const ERD_BLUE = 'var(--erd-blue, #1d9bf0)';

// Presentational grouping only — buckets tables into named subsystems for the
// sidebar. Any table not listed (e.g. a brand-new one) falls into "Other", so
// the browser always reflects the live schema.
const DOMAINS: Record<string, string[]> = {
  'User & Auth': ['users', 'roles', 'verification_codes', 'user_stats', 'user_faculties'],
  Profile: ['user_profiles', 'user_settings', 'profile_field_changes'],
  Social: ['user_relationships', 'user_blocks', 'notification_user_mutes'],
  Content: ['threads', 'thread_media', 'thread_polls', 'thread_poll_options', 'thread_poll_votes', 'thread_summaries'],
  Interactions: ['likes', 'bookmarks', 'reposts', 'thread_pins'],
  Hashtags: ['hashtags', 'thread_hashtags'],
  Events: ['events', 'thread_events', 'event_participants', 'event_faculty_locations', 'event_reviews', 'event_pins', 'event_thread_pins', 'event_action_confirmations'],
  Communities: ['communities', 'community_members', 'community_rules', 'community_requests', 'community_thread_pins', 'community_pins', 'community_action_confirmations'],
  'Community Events': ['community_events', 'community_event_invitees', 'community_event_comments', 'community_event_reactions'],
  Chat: ['chat_rooms', 'chat_room_members', 'chats', 'chat_media', 'chat_reads', 'chat_reactions'],
  Moderation: ['reports', 'violations', 'trend_reports', 'trend_report_reasons'],
  Premium: ['premium_subscriptions', 'orders'],
  Bookmarks: ['bookmark_collections', 'bookmark_collection_items'],
  'Fox AI': ['fox_sessions', 'fox_histories'],
  Analytics: ['trending_user_snapshots'],
  Lookups: ['faculties', 'faculty_locations', 'categories'],
  Feedback: ['feedbacks', 'feedback_votes', 'feedback_media'],
};

function domainOf(table: string): string {
  for (const [domain, tables] of Object.entries(DOMAINS)) {
    if (tables.includes(table)) return domain;
  }
  return 'Other';
}

function columnBadge(c: SchemaColumn): { label: string; color: string } | null {
  if (c.is_pk) return { label: 'PK', color: ERD_BLUE };
  if (c.is_fk) return { label: 'FK', color: '#e0a800' };
  if (c.is_unique) return { label: 'UK', color: '#10b981' };
  return null;
}

// Map a Postgres type family to an accent color for the type label / dot.
function typeColor(type: string): string {
  const t = type.toLowerCase();
  if (/(int|serial|numeric|decimal|real|double|float)/.test(t)) return '#3b82f6';
  if (/(text|char|varchar)/.test(t)) return '#10b981';
  if (/(time|date)/.test(t)) return '#a855f7';
  if (/bool/.test(t)) return '#e0a800';
  if (/uuid/.test(t)) return '#ec4899';
  if (/(json|jsonb)/.test(t)) return '#06b6d4';
  return 'var(--text-tertiary)';
}


export function SchemaBrowser() {
  const token = useAuthStore((s) => s.token);
  const router = useRouter();
  const searchParams = useSearchParams();
  // URL is the source of truth for the selected table — no separate state.
  const selected = searchParams.get('table');

  const [filter, setFilter] = useState('');
  const [colFilter, setColFilter] = useState('');
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
  const searchRef = useRef<HTMLInputElement | null>(null);

  const q = useQuery({
    queryKey: ['admin', 'schema'],
    queryFn: () => getSchema(token!),
    enabled: !!token,
  });

  const tables = useMemo(() => q.data?.tables ?? [], [q.data]);
  const byName = useMemo(() => {
    const m: Record<string, SchemaTable> = {};
    for (const t of tables) m[t.name] = t;
    return m;
  }, [tables]);

  const stats = useMemo(() => {
    let cols = 0;
    let fks = 0;
    for (const t of tables) {
      cols += t.columns.length;
      fks += t.columns.filter((c) => c.is_fk).length;
    }
    return { tables: tables.length, cols, fks };
  }, [tables]);

  const select = useCallback(
    (name: string | null) => {
      setColFilter('');
      const params = new URLSearchParams(Array.from(searchParams.entries()));
      if (name) params.set('table', name);
      else params.delete('table');
      router.replace(`?${params.toString()}`, { scroll: false });
    },
    [router, searchParams],
  );

  const searching = filter.trim().length > 0;
  // Groups start collapsed; an active search force-opens them, otherwise fall
  // back to the user's manual toggle (default closed).
  const isCollapsed = useCallback(
    (domain: string) => (searching ? false : (collapsed[domain] ?? true)),
    [searching, collapsed],
  );

  // Group + filter table list for the sidebar.
  const grouped = useMemo(() => {
    const f = filter.trim().toLowerCase();
    const buckets: Record<string, string[]> = {};
    for (const t of tables) {
      if (f && !t.name.toLowerCase().includes(f)) continue;
      const d = domainOf(t.name);
      (buckets[d] = buckets[d] || []).push(t.name);
    }
    for (const d in buckets) buckets[d].sort();
    const order = [...Object.keys(DOMAINS), 'Other'];
    return order.filter((d) => buckets[d]?.length).map((d) => ({ domain: d, tables: buckets[d] }));
  }, [tables, filter]);

  // All matched tables in one sorted flat list (used for the search view).
  const flatMatches = useMemo(() => grouped.flatMap((g) => g.tables).sort(), [grouped]);

  // Flat ordered list of currently-visible tables for arrow-key navigation —
  // the flat search list when searching, else expanded groups.
  const flat = useMemo(
    () => (searching ? flatMatches : grouped.flatMap((g) => (isCollapsed(g.domain) ? [] : g.tables))),
    [searching, flatMatches, grouped, isCollapsed],
  );

  // Keyboard: "/" focuses search, arrows move selection, Enter keeps it.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const el = document.activeElement;
      const typing = el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement;
      if (e.key === '/' && !typing) {
        e.preventDefault();
        searchRef.current?.focus();
        return;
      }
      if (typing && el !== searchRef.current) return;
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        if (!flat.length) return;
        e.preventDefault();
        const idx = selected ? flat.indexOf(selected) : -1;
        const nextIdx =
          e.key === 'ArrowDown'
            ? Math.min(flat.length - 1, idx + 1)
            : Math.max(0, idx <= 0 ? 0 : idx - 1);
        select(flat[nextIdx]);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [flat, selected, select]);

  const active = selected ? byName[selected] : null;

  if (q.isLoading) return <SchemaSkeleton />;
  if (q.isError) {
    return (
      <div style={{ padding: 24, color: 'var(--text-secondary)' }}>
        Failed to load schema. {(q.error as Error)?.message}
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      <ShimmerStyle />
      {/* Stat chips */}
      <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
        <StatChip icon={<Table2 size={14} />} label="tables" value={stats.tables} />
        <StatChip icon={<Hash size={14} />} label="columns" value={stats.cols} />
        <StatChip icon={<Link2 size={14} />} label="relationships" value={stats.fks} />
        <span style={{ flex: 1 }} />
        <span style={{ alignSelf: 'center', fontSize: 11.5, color: 'var(--text-tertiary)' }}>
          press <kbd style={kbdStyle}>/</kbd> to search · <kbd style={kbdStyle}>↑↓</kbd> to navigate
        </span>
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '290px 1fr',
          gap: 16,
          height: 'calc(100vh - 290px)',
          minHeight: 420,
        }}
      >
        {/* Sidebar */}
        <aside
          style={{
            border: '1px solid var(--border)',
            borderRadius: 10,
            background: 'var(--bg-card)',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}
        >
          <div style={{ position: 'relative', padding: 10, borderBottom: '1px solid var(--border)' }}>
            <Search size={15} style={{ position: 'absolute', left: 20, top: 19, color: 'var(--text-tertiary)' }} />
            <input
              ref={searchRef}
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="Search tables…  ( / )"
              style={{
                width: '100%',
                padding: '8px 10px 8px 32px',
                fontSize: 13,
                borderRadius: 7,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
              }}
            />
          </div>

          <div style={{ overflowY: 'auto', flex: 1, padding: '8px 8px' }}>
            {/* While searching, show a flat list of matched tables (no group
                headers). Otherwise show collapsible domain groups. */}
            {searching ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {flatMatches.map((name) => (
                  <TableRow key={name} name={name} selected={selected === name} indent={12} onSelect={select} />
                ))}
              </div>
            ) : (
              grouped.map((g) => {
                const groupCollapsed = isCollapsed(g.domain);
                return (
                  <div key={g.domain} style={{ marginBottom: 6 }}>
                    <button
                      type="button"
                      onClick={() => setCollapsed((c) => ({ ...c, [g.domain]: !(c[g.domain] ?? true) }))}
                      className="schema-group-header"
                      style={{
                        ...groupHeaderStyle,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 6,
                        width: '100%',
                        padding: '8px 12px',
                        border: 'none',
                        borderRadius: 7,
                        background: 'transparent',
                        cursor: 'pointer',
                      }}
                    >
                      <span
                        style={{
                          display: 'inline-flex',
                          transition: 'transform .15s',
                          transform: groupCollapsed ? 'rotate(0deg)' : 'rotate(90deg)',
                          color: 'var(--text-tertiary)',
                        }}
                      >
                        <ChevronRight size={13} />
                      </span>
                      <span style={{ flex: 1, textAlign: 'left' }}>{g.domain}</span>
                      <span
                        style={{
                          fontSize: 10.5,
                          fontWeight: 600,
                          color: 'var(--text-tertiary)',
                          background: 'var(--bg-subtle)',
                          borderRadius: 999,
                          padding: '1px 7px',
                        }}
                      >
                        {g.tables.length}
                      </span>
                    </button>
                    {!groupCollapsed && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 1, marginTop: 2 }}>
                        {g.tables.map((name) => (
                          <TableRow key={name} name={name} selected={selected === name} indent={26} onSelect={select} />
                        ))}
                      </div>
                    )}
                  </div>
                );
              })
            )}
            {!grouped.length && (
              <div style={{ padding: 16, fontSize: 13, color: 'var(--text-tertiary)' }}>No tables match.</div>
            )}
          </div>

          <div style={{ padding: '8px 14px', borderTop: '1px solid var(--border)', fontSize: 11.5, color: 'var(--text-tertiary)' }}>
            {tables.length} tables · live from database
          </div>
        </aside>

        {/* Detail */}
        <section
          style={{
            border: '1px solid var(--border)',
            borderRadius: 10,
            background: 'var(--bg-card)',
            overflowY: 'auto',
            padding: active ? 22 : 0,
          }}
        >
          {!active ? (
            <div
              style={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 10,
                color: 'var(--text-tertiary)',
              }}
            >
              <Table2 size={32} strokeWidth={1.5} />
              <p style={{ fontSize: 14 }}>Select a table to view its columns and relationships.</p>
            </div>
          ) : (
            <TableDetail
              table={active}
              colFilter={colFilter}
              onColFilter={setColFilter}
              onSelect={select}
            />
          )}
        </section>
      </div>
    </div>
  );
}

function TableDetail({
  table,
  colFilter,
  onColFilter,
  onSelect,
}: {
  table: SchemaTable;
  colFilter: string;
  onColFilter: (v: string) => void;
  onSelect: (name: string) => void;
}) {
  const fks = table.columns.filter((c) => c.is_fk);
  const f = colFilter.trim().toLowerCase();
  const cols = f ? table.columns.filter((c) => c.name.toLowerCase().includes(f)) : table.columns;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
        <Table2 size={20} style={{ color: ERD_BLUE }} />
        <CopyableName name={table.name} style={{ fontSize: 20, fontWeight: 700, color: 'var(--text)' }} />
        <span style={{ fontSize: 12.5, color: 'var(--text-tertiary)' }}>{table.columns.length} columns</span>
        <span style={{ flex: 1 }} />
        <div style={{ position: 'relative' }}>
          <Search size={13} style={{ position: 'absolute', left: 9, top: 8, color: 'var(--text-tertiary)' }} />
          <input
            value={colFilter}
            onChange={(e) => onColFilter(e.target.value)}
            placeholder="Filter columns…"
            style={{
              padding: '5px 10px 5px 28px',
              fontSize: 12.5,
              borderRadius: 6,
              border: '1px solid var(--border)',
              background: 'var(--bg-subtle)',
              color: 'var(--text)',
              width: 180,
            }}
          />
        </div>
      </div>

      {/* Columns */}
      <div style={{ border: '1px solid var(--border)', borderRadius: 8, overflow: 'hidden' }}>
        <div style={{ ...colGrid, ...colHeaderStyle }}>
          <span>Column</span>
          <span>Type</span>
          <span>Key</span>
          <span>Nullable</span>
        </div>
        {cols.map((c) => {
          const badge = columnBadge(c);
          return (
            <div key={c.name} style={{ ...colGrid, ...colRowStyle }} className="schema-col-row">
              <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontWeight: badge ? 600 : 400 }}>
                {c.is_pk && <KeyRound size={13} style={{ color: ERD_BLUE }} />}
                <CopyableName name={c.name} />
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <span style={{ width: 7, height: 7, borderRadius: 2, background: typeColor(c.type), flexShrink: 0 }} />
                <span style={{ color: 'var(--text-tertiary)', fontStyle: 'italic' }}>{c.type}</span>
              </span>
              <span>
                {badge && (
                  <span
                    style={{
                      fontSize: 10.5,
                      fontWeight: 700,
                      padding: '1px 6px',
                      borderRadius: 4,
                      color: badge.color,
                      border: `1px solid ${badge.color}`,
                    }}
                  >
                    {badge.label}
                  </span>
                )}
                {c.is_fk && c.fk_table && (
                  <button
                    type="button"
                    onClick={() => onSelect(c.fk_table!)}
                    title={`Go to ${c.fk_table}`}
                    style={{
                      marginLeft: 6,
                      fontSize: 11,
                      color: ERD_BLUE,
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      padding: 0,
                    }}
                  >
                    → {c.fk_table}
                  </button>
                )}
              </span>
              <span style={{ color: 'var(--text-tertiary)', fontSize: 12 }}>{c.nullable ? 'yes' : '—'}</span>
            </div>
          );
        })}
        {!cols.length && (
          <div style={{ padding: 14, fontSize: 13, color: 'var(--text-tertiary)' }}>No columns match.</div>
        )}
      </div>

      <RelationList
        icon={<Link2 size={15} />}
        title="References →"
        empty="No outgoing foreign keys."
        rows={fks.map((c) => ({
          key: c.name,
          label: c.name,
          target: c.fk_table ?? '',
          targetCol: c.fk_column ?? '',
        }))}
        onSelect={onSelect}
      />

      <RelationList
        icon={<Hash size={15} />}
        title="Referenced by ←"
        empty="No tables reference this one."
        rows={table.referenced_by.map((r) => ({
          key: `${r.from_table}.${r.from_column}`,
          label: `${r.from_table}.${r.from_column}`,
          target: r.from_table,
          targetCol: r.to_column,
        }))}
        onSelect={onSelect}
      />
    </div>
  );
}

interface RelRow {
  key: string;
  label: string;
  target: string;
  targetCol: string;
}

function RelationList({
  icon,
  title,
  empty,
  rows,
  onSelect,
}: {
  icon: React.ReactNode;
  title: string;
  empty: string;
  rows: RelRow[];
  onSelect: (name: string) => void;
}) {
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 7, marginBottom: 8, color: 'var(--text-secondary)' }}>
        {icon}
        <h3 style={{ fontSize: 14, fontWeight: 700 }}>{title}</h3>
        <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>({rows.length})</span>
      </div>
      {!rows.length ? (
        <p style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>{empty}</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {rows.map((r) => (
            <button
              key={r.key}
              type="button"
              onClick={() => onSelect(r.target)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '7px 12px',
                fontSize: 13,
                borderRadius: 7,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                cursor: 'pointer',
                textAlign: 'left',
              }}
            >
              <span style={{ fontWeight: 600 }}>{r.label}</span>
              <span style={{ color: 'var(--text-tertiary)' }}>→</span>
              <span style={{ color: ERD_BLUE, fontWeight: 600 }}>
                {r.target}.{r.targetCol}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

// Click-to-copy text with a transient confirmation tick.
function CopyableName({ name, style }: { name: string; style?: React.CSSProperties }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard?.writeText(name).then(
      () => {
        setCopied(true);
        setTimeout(() => setCopied(false), 1200);
      },
      () => {},
    );
  };
  return (
    <button
      type="button"
      onClick={copy}
      title="Click to copy"
      className="schema-copyable"
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 5,
        background: 'none',
        border: 'none',
        padding: 0,
        cursor: 'pointer',
        color: 'inherit',
        font: 'inherit',
        ...style,
      }}
    >
      {name}
      {copied ? (
        <Check size={12} style={{ color: '#10b981' }} />
      ) : (
        <Copy size={11} className="schema-copy-icon" style={{ color: 'var(--text-tertiary)', opacity: 0 }} />
      )}
    </button>
  );
}

function TableRow({
  name,
  selected,
  indent,
  onSelect,
}: {
  name: string;
  selected: boolean;
  indent: number;
  onSelect: (name: string) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => onSelect(name)}
      className="schema-table-row"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 9,
        width: '100%',
        textAlign: 'left',
        padding: `7px 12px 7px ${indent}px`,
        fontSize: 13,
        cursor: 'pointer',
        borderRadius: 7,
        background: selected ? 'var(--bg-hover)' : 'transparent',
        color: selected ? 'var(--text)' : 'var(--text-secondary)',
        border: 'none',
        boxShadow: selected ? `inset 2px 0 0 ${ERD_BLUE}` : 'none',
        fontWeight: selected ? 600 : 400,
        transition: 'background .1s',
      }}
    >
      <Table2 size={14} style={{ flexShrink: 0, color: selected ? ERD_BLUE : 'var(--text-tertiary)' }} />
      {name}
    </button>
  );
}

function StatChip({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 7,
        padding: '6px 12px',
        borderRadius: 8,
        border: '1px solid var(--border)',
        background: 'var(--bg-card)',
        fontSize: 13,
        color: 'var(--text-secondary)',
      }}
    >
      <span style={{ color: ERD_BLUE, display: 'inline-flex' }}>{icon}</span>
      <strong style={{ color: 'var(--text)', fontWeight: 700 }}>{value}</strong>
      {label}
    </span>
  );
}

function SchemaSkeleton() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      <ShimmerStyle />
      <div style={{ display: 'flex', gap: 10 }}>
        {[0, 1, 2].map((i) => (
          <div key={i} className="schema-shimmer" style={{ width: 130, height: 34, borderRadius: 8 }} />
        ))}
      </div>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '290px 1fr',
          gap: 16,
          height: 'calc(100vh - 290px)',
          minHeight: 420,
        }}
      >
        <div style={{ border: '1px solid var(--border)', borderRadius: 10, padding: 12, display: 'flex', flexDirection: 'column', gap: 8 }}>
          {Array.from({ length: 10 }).map((_, i) => (
            <div key={i} className="schema-shimmer" style={{ height: 22, borderRadius: 6, width: `${60 + ((i * 7) % 35)}%` }} />
          ))}
        </div>
        <div style={{ border: '1px solid var(--border)', borderRadius: 10, padding: 22, display: 'flex', flexDirection: 'column', gap: 10 }}>
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="schema-shimmer" style={{ height: 26, borderRadius: 6 }} />
          ))}
        </div>
      </div>
    </div>
  );
}

function ShimmerStyle() {
  return (
    <style>{`
      .schema-shimmer{background:linear-gradient(90deg,var(--bg-subtle) 25%,var(--bg-hover) 37%,var(--bg-subtle) 63%);background-size:400% 100%;animation:schemaShimmer 1.3s ease infinite}
      @keyframes schemaShimmer{0%{background-position:100% 0}100%{background-position:-100% 0}}
      .schema-col-row:hover{background:var(--bg-subtle)}
      .schema-copyable:hover .schema-copy-icon{opacity:1 !important}
      .schema-group-header:hover{background:var(--bg-subtle)}
      .schema-table-row:hover{background:var(--bg-subtle)}
    `}</style>
  );
}

const groupHeaderStyle: React.CSSProperties = {
  padding: '6px 12px 2px',
  fontSize: 11,
  fontWeight: 700,
  textTransform: 'uppercase',
  letterSpacing: '0.04em',
  color: 'var(--text-tertiary)',
};

const colGrid: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: '1fr 200px 130px 70px',
  alignItems: 'center',
};

const colHeaderStyle: React.CSSProperties = {
  padding: '8px 14px',
  fontSize: 11.5,
  fontWeight: 700,
  textTransform: 'uppercase',
  letterSpacing: '0.03em',
  color: 'var(--text-tertiary)',
  background: 'var(--bg-subtle)',
  borderBottom: '1px solid var(--border)',
};

const colRowStyle: React.CSSProperties = {
  padding: '7px 14px',
  fontSize: 13,
  borderBottom: '1px solid var(--border)',
  color: 'var(--text)',
};

const kbdStyle: React.CSSProperties = {
  padding: '1px 5px',
  borderRadius: 4,
  border: '1px solid var(--border)',
  background: 'var(--bg-subtle)',
  fontSize: 10.5,
  fontFamily: 'inherit',
};
