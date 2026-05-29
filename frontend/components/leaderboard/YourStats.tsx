'use client';

import {
  useEffect,
  useMemo,
  useState,
  type CSSProperties,
  type ReactNode,
} from 'react';
import {
  CartesianGrid,
  Line,
  LineChart,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useMeasuredSize } from '@/hooks/useMeasuredSize';
import { Maximize2 } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useMyTrendingCurrent } from '@/hooks/useMyTrendingCurrent';
import { useMyTrendingHistory } from '@/hooks/useMyTrendingHistory';
import { SnapshotModal } from './SnapshotModal';
import { ChartExpandModal } from './ChartExpandModal';

const MONTH_NAMES = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
];

type RangeMode = 'month' | 'quarter' | 'range';

const MODE_LABEL: Record<RangeMode, string> = {
  month: 'Month',
  quarter: 'Quarter',
  range: 'Custom',
};

function pad2(n: number): string {
  return n.toString().padStart(2, '0');
}

function toISODate(d: Date): string {
  return `${d.getUTCFullYear()}-${pad2(d.getUTCMonth() + 1)}-${pad2(d.getUTCDate())}`;
}

function monthRange(year: number, month: number): { start: string; end: string } {
  const start = new Date(Date.UTC(year, month - 1, 1));
  const end = new Date(Date.UTC(year, month, 0));
  return { start: toISODate(start), end: toISODate(end) };
}

function quarterRange(year: number, quarter: number): { start: string; end: string } {
  const startMonth = (quarter - 1) * 3 + 1;
  const start = new Date(Date.UTC(year, startMonth - 1, 1));
  const end = new Date(Date.UTC(year, startMonth + 2, 0));
  return { start: toISODate(start), end: toISODate(end) };
}

function formatWeekLabel(weekStart: string): string {
  const d = new Date(weekStart);
  if (Number.isNaN(d.getTime())) return weekStart;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

function formatRangeLabel(start: string, end: string): string {
  if (!start || !end) return '';
  const s = new Date(start);
  const e = new Date(end);
  const fmt = (d: Date) =>
    d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  return `${fmt(s)} → ${fmt(e)}`;
}

function HoverWrap({
  children,
  style,
  bare = false,
}: {
  children: ReactNode;
  style?: CSSProperties;
  bare?: boolean;
}) {
  const [hover, setHover] = useState(false);
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        transition: 'background 0.15s ease, border-color 0.15s ease, transform 0.15s ease',
        background: bare
          ? hover
            ? 'var(--bg-subtle)'
            : 'transparent'
          : hover
            ? 'var(--bg-hover)'
            : 'var(--bg-card)',
        borderWidth: bare ? 0 : 1,
        borderStyle: bare ? 'none' : 'solid',
        borderColor: bare ? 'transparent' : hover ? 'var(--text-tertiary)' : 'var(--border)',
        ...style,
      }}
    >
      {children}
    </div>
  );
}

function StatCell({ label, value }: { label: string; value: number }) {
  const [hover, setHover] = useState(false);
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        background: hover ? 'var(--bg-hover)' : 'var(--bg-subtle)',
        borderWidth: 1,
        borderStyle: 'solid',
        borderColor: hover ? 'var(--text-tertiary)' : 'var(--border)',
        borderRadius: 10,
        padding: '10px 12px',
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
        transition: 'background 0.15s ease, border-color 0.15s ease, transform 0.15s ease',
        transform: hover ? 'translateY(-1px)' : 'translateY(0)',
        cursor: 'default',
      }}
    >
      <span
        style={{
          fontSize: 11,
          color: 'var(--text-tertiary)',
          textTransform: 'uppercase',
          letterSpacing: '0.04em',
        }}
      >
        {label}
      </span>
      <span style={{ fontSize: 18, fontWeight: 700, color: 'var(--text)' }}>{value}</span>
    </div>
  );
}

function ModeButton({
  active,
  label,
  onClick,
}: {
  active: boolean;
  label: string;
  onClick: () => void;
}) {
  const [hover, setHover] = useState(false);
  return (
    <button
      type="button"
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        flex: 1,
        padding: '8px 12px',
        borderRadius: 8,
        border: 'none',
        background: active ? 'var(--text)' : hover ? 'var(--bg-hover)' : 'transparent',
        color: active ? 'var(--bg)' : 'var(--text)',
        fontSize: 12,
        fontWeight: 700,
        letterSpacing: '0.02em',
        cursor: 'pointer',
        transition: 'background 0.15s ease, color 0.15s ease',
      }}
    >
      {label}
    </button>
  );
}

function selectStyle(): CSSProperties {
  return {
    background: 'var(--bg-subtle)',
    color: 'var(--text)',
    border: '1px solid var(--border)',
    borderRadius: 8,
    padding: '6px 10px',
    fontSize: 12,
    fontWeight: 600,
    cursor: 'pointer',
    outline: 'none',
  };
}

function inputStyle(): CSSProperties {
  return {
    background: 'var(--bg-subtle)',
    color: 'var(--text)',
    border: '1px solid var(--border)',
    borderRadius: 8,
    padding: '6px 10px',
    fontSize: 12,
    fontWeight: 600,
    outline: 'none',
    colorScheme: 'dark',
  };
}

interface RangeState {
  mode: RangeMode;
  start?: string;
  end?: string;
  year?: number;
  month?: number;
  quarter?: number;
}

function readUrlState(): RangeState {
  const params = new URLSearchParams(window.location.search);
  const modeRaw = params.get('mode');
  const mode: RangeMode =
    modeRaw === 'range' || modeRaw === 'quarter' ? modeRaw : 'month';
  const year = params.get('year') ? Number(params.get('year')) : undefined;
  const month = params.get('month') ? Number(params.get('month')) : undefined;
  const quarter = params.get('quarter') ? Number(params.get('quarter')) : undefined;
  return {
    mode,
    start: params.get('start') || undefined,
    end: params.get('end') || undefined,
    year,
    month,
    quarter,
  };
}

function syncUrl(state: RangeState) {
  if (typeof window === 'undefined') return;
  const next = new URLSearchParams(window.location.search);
  next.set('tab', 'stats');
  next.set('mode', state.mode);
  if (state.year !== undefined) next.set('year', String(state.year));
  else next.delete('year');
  if (state.month !== undefined) next.set('month', String(state.month));
  else next.delete('month');
  if (state.quarter !== undefined) next.set('quarter', String(state.quarter));
  else next.delete('quarter');
  if (state.start) next.set('start', state.start);
  else next.delete('start');
  if (state.end) next.set('end', state.end);
  else next.delete('end');
  const qs = next.toString();
  window.history.replaceState(null, '', qs ? `?${qs}` : window.location.pathname);
}

export function YourStats() {
  const [rangeState, setRangeState] = useState<RangeState>({ mode: 'month' });

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setRangeState(readUrlState());
  }, []);

  const [selectedSnapshotId, setSelectedSnapshotId] = useState<number | null>(null);
  const [expandOpen, setExpandOpen] = useState(false);

  const [chartRef, { w: chartW, h: chartH }] = useMeasuredSize<HTMLDivElement>();

  const computedRange = useMemo<{ start?: string; end?: string }>(() => {
    if (rangeState.mode === 'range' && rangeState.start && rangeState.end) {
      return { start: rangeState.start, end: rangeState.end };
    }
    if (rangeState.mode === 'month' && rangeState.year && rangeState.month) {
      return monthRange(rangeState.year, rangeState.month);
    }
    if (rangeState.mode === 'quarter' && rangeState.year && rangeState.quarter) {
      return quarterRange(rangeState.year, rangeState.quarter);
    }
    return {};
  }, [rangeState]);

  const currentQuery = useMyTrendingCurrent();
  const historyQuery = useMyTrendingHistory(computedRange.start, computedRange.end);

  const current = currentQuery.data;
  const result = historyQuery.data;
  const history = useMemo(() => result?.data ?? [], [result]);
  const availableMonths = useMemo(() => result?.availableMonths ?? [], [result]);

  // Backfill year/month from backend default on first load (no URL params present).
  useEffect(() => {
    if (rangeState.mode !== 'month') return;
    if (rangeState.year || rangeState.month) return;
    if (!result?.startDate) return;
    const d = new Date(result.startDate);
    const next: RangeState = {
      ...rangeState,
      year: d.getUTCFullYear(),
      month: d.getUTCMonth() + 1,
    };
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setRangeState(next);
    syncUrl(next);
  }, [rangeState, result?.startDate]);

  const chartData = useMemo(
    () =>
      history.map((h) => ({
        snapshotId: h.snapshotId,
        weekLabel: formatWeekLabel(h.weekStart),
        score: Number(h.score.toFixed(1)),
        rank: h.rank,
      })),
    [history],
  );

  const isLoading = currentQuery.isLoading || historyQuery.isLoading;

  const availableYears = useMemo(() => {
    const set = new Set<number>();
    availableMonths.forEach((m) => set.add(m.year));
    return Array.from(set).sort((a, b) => b - a);
  }, [availableMonths]);

  const monthsForYear = (year: number): number[] => {
    const set = new Set<number>();
    availableMonths.forEach((m) => {
      if (m.year === year) set.add(m.month);
    });
    return Array.from(set).sort((a, b) => a - b);
  };

  function applyState(next: RangeState) {
    setRangeState(next);
    syncUrl(next);
  }

  function changeMode(nextMode: RangeMode) {
    const base: RangeState = { mode: nextMode };
    if (nextMode === 'month' && availableMonths.length > 0) {
      base.year = availableMonths[0].year;
      base.month = availableMonths[0].month;
    } else if (nextMode === 'quarter' && availableMonths.length > 0) {
      const m = availableMonths[0];
      base.year = m.year;
      base.quarter = Math.floor((m.month - 1) / 3) + 1;
    } else if (nextMode === 'range' && result?.startDate && result?.endDate) {
      base.start = result.startDate;
      base.end = result.endDate;
    }
    applyState(base);
  }

  function setMonthSelection(year: number, month: number) {
    applyState({ mode: 'month', year, month });
  }

  function setQuarterSelection(year: number, quarter: number) {
    applyState({ mode: 'quarter', year, quarter });
  }

  function setRangeSelection(start: string, end: string) {
    applyState({ mode: 'range', start, end });
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <HoverWrap
        bare
        style={{
          borderRadius: 16,
          padding: 20,
        }}
      >
        {isLoading && !current ? (
          <div style={{ color: 'var(--text-tertiary)', fontSize: 14 }}>Loading your stats…</div>
        ) : current ? (
          <>
            <div style={{ display: 'flex', alignItems: 'center', gap: 14, marginBottom: 16 }}>
              <Avatar
                username={current.name || current.username}
                avatarUrl={current.profile || null}
                size={64}
              />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <span
                    style={{
                      fontWeight: 700,
                      fontSize: 16,
                      color: 'var(--text)',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {current.name || current.username}
                  </span>
                  {current.isPremium && <VerifiedBadge size={15} />}
                </div>
                <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
                  @{current.username}
                </div>
                <span
                  style={{
                    display: 'inline-block',
                    marginTop: 6,
                    padding: '2px 10px',
                    borderRadius: 9999,
                    background: 'var(--text)',
                    color: 'var(--bg)',
                    fontSize: 11,
                    fontWeight: 700,
                  }}
                >
                  {current.rank !== null ? `Rank #${current.rank}` : 'Unranked this week'}
                </span>
              </div>
            </div>

            <StatBlock>
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 2,
                  padding: '12px 0',
                  borderRadius: 12,
                }}
              >
                <span
                  style={{
                    fontSize: 11,
                    color: 'var(--text-tertiary)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.06em',
                  }}
                >
                  Weekly Score
                </span>
                <span
                  style={{
                    fontSize: 32,
                    fontWeight: 800,
                    color: 'var(--text)',
                    lineHeight: 1.1,
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {current.score.toFixed(1)}
                </span>
              </div>
            </StatBlock>

            <div style={{ marginTop: 16, marginBottom: 16 }}>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: 'var(--text-secondary)',
                  marginBottom: 8,
                  textTransform: 'uppercase',
                  letterSpacing: '0.04em',
                }}
              >
                Engagement Received
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                <StatCell label="Likes" value={current.received.likes} />
                <StatCell label="Comments" value={current.received.comments} />
                <StatCell label="Bookmarks" value={current.received.bookmarks} />
                <StatCell label="Reposts" value={current.received.reposts} />
                <StatCell label="Threads" value={current.received.threadCount} />
              </div>
            </div>

            <div>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: 'var(--text-secondary)',
                  marginBottom: 8,
                  textTransform: 'uppercase',
                  letterSpacing: '0.04em',
                }}
              >
                Activity Given
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                <StatCell label="Likes" value={current.given.likes} />
                <StatCell label="Comments" value={current.given.comments} />
                <StatCell label="Bookmarks" value={current.given.bookmarks} />
                <StatCell label="Posts" value={current.given.posts} />
              </div>
            </div>
          </>
        ) : (
          <div style={{ color: 'var(--text-tertiary)', fontSize: 14 }}>
            No activity yet this week. Post or engage to start ranking.
          </div>
        )}
      </HoverWrap>

      <HoverWrap
        style={{
          borderRadius: 16,
          padding: 20,
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 12,
            marginBottom: 12,
            flexWrap: 'wrap',
          }}
        >
          <div
            style={{
              fontSize: 12,
              fontWeight: 600,
              color: 'var(--text-secondary)',
              textTransform: 'uppercase',
              letterSpacing: '0.04em',
            }}
          >
            History ({chartData.length || 0} {chartData.length === 1 ? 'week' : 'weeks'})
            {result?.startDate && result?.endDate && (
              <span
                style={{
                  marginLeft: 10,
                  fontSize: 11,
                  fontWeight: 500,
                  color: 'var(--text-tertiary)',
                  textTransform: 'none',
                  letterSpacing: 0,
                }}
              >
                {formatRangeLabel(result.startDate, result.endDate)}
              </span>
            )}
          </div>
          <button
            type="button"
            onClick={() => setExpandOpen(true)}
            aria-label="Expand chart"
            className="btn-ghost"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 12px',
              background: 'var(--bg-subtle)',
              border: '1px solid var(--border)',
              borderRadius: 8,
              cursor: 'pointer',
              color: 'var(--text)',
              fontSize: 12,
              fontWeight: 600,
            }}
          >
            <Maximize2 size={14} strokeWidth={2} />
            Expand
          </button>
        </div>

        <div
          style={{
            display: 'flex',
            gap: 4,
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border)',
            borderRadius: 10,
            padding: 3,
            marginBottom: 12,
          }}
        >
          {(['month', 'quarter', 'range'] as RangeMode[]).map((m) => (
            <ModeButton
              key={m}
              label={MODE_LABEL[m]}
              active={rangeState.mode === m}
              onClick={() => changeMode(m)}
            />
          ))}
        </div>

        {rangeState.mode === 'month' && availableMonths.length > 0 && (
          <div style={{ display: 'flex', gap: 8, marginBottom: 14, flexWrap: 'wrap' }}>
            <select
              value={rangeState.year ?? availableYears[0] ?? ''}
              onChange={(e) => {
                const y = Number(e.target.value);
                const months = monthsForYear(y);
                const m = rangeState.month && months.includes(rangeState.month) ? rangeState.month : months[0];
                setMonthSelection(y, m);
              }}
              style={selectStyle()}
            >
              {availableYears.map((y) => (
                <option key={y} value={y}>
                  {y}
                </option>
              ))}
            </select>
            <select
              value={rangeState.month ?? ''}
              onChange={(e) =>
                setMonthSelection(rangeState.year ?? availableYears[0], Number(e.target.value))
              }
              style={selectStyle()}
            >
              {monthsForYear(rangeState.year ?? availableYears[0] ?? 0).map((m) => (
                <option key={m} value={m}>
                  {MONTH_NAMES[m - 1]}
                </option>
              ))}
            </select>
          </div>
        )}

        {rangeState.mode === 'quarter' && availableMonths.length > 0 && (
          <div style={{ display: 'flex', gap: 8, marginBottom: 14, flexWrap: 'wrap' }}>
            <select
              value={rangeState.year ?? availableYears[0] ?? ''}
              onChange={(e) =>
                setQuarterSelection(Number(e.target.value), rangeState.quarter ?? 1)
              }
              style={selectStyle()}
            >
              {availableYears.map((y) => (
                <option key={y} value={y}>
                  {y}
                </option>
              ))}
            </select>
            <select
              value={rangeState.quarter ?? 1}
              onChange={(e) =>
                setQuarterSelection(rangeState.year ?? availableYears[0], Number(e.target.value))
              }
              style={selectStyle()}
            >
              {[1, 2, 3, 4].map((q) => (
                <option key={q} value={q}>
                  Q{q} ({MONTH_NAMES[(q - 1) * 3].slice(0, 3)}–{MONTH_NAMES[(q - 1) * 3 + 2].slice(0, 3)})
                </option>
              ))}
            </select>
          </div>
        )}

        {rangeState.mode === 'range' && (
          <RangeInput
            initialStart={rangeState.start ?? result?.startDate ?? ''}
            initialEnd={rangeState.end ?? result?.endDate ?? ''}
            onApply={(s, e) => {
              setRangeSelection(s, e);
              historyQuery.refetch();
            }}
          />
        )}

        {chartData.length === 0 ? (
          <div
            style={{
              padding: '32px 8px',
              textAlign: 'center',
              color: 'var(--text-tertiary)',
              fontSize: 13,
            }}
          >
            No snapshots in this range.
          </div>
        ) : (
          <div
            ref={chartRef}
            style={{ width: '100%', height: 260, minWidth: 0, position: 'relative' }}
          >
            {chartW > 0 && chartH > 0 && (
            <LineChart width={chartW} height={chartH} data={chartData} margin={{ top: 12, right: 16, left: 0, bottom: 4 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis
                  dataKey="weekLabel"
                  stroke="var(--text-tertiary)"
                  fontSize={11}
                  tickLine={false}
                  interval="preserveStartEnd"
                  minTickGap={24}
                />
                <YAxis
                  stroke="var(--text-tertiary)"
                  fontSize={11}
                  tickLine={false}
                  width={32}
                />
                <Tooltip
                  contentStyle={{
                    background: 'var(--bg-card)',
                    border: '1px solid var(--border)',
                    borderRadius: 10,
                    fontSize: 12,
                  }}
                  labelStyle={{ color: 'var(--text)' }}
                  itemStyle={{ color: 'var(--text)' }}
                />
                <Line
                  type="monotone"
                  dataKey="score"
                  stroke="var(--text)"
                  strokeWidth={2}
                  dot={(props: {
                    cx?: number;
                    cy?: number;
                    payload?: { snapshotId: number };
                    index?: number;
                  }) => {
                    const { cx, cy, payload, index } = props;
                    if (cx === undefined || cy === undefined || !payload) {
                      return <g key={`dot-empty-${index ?? 0}`} />;
                    }
                    return (
                      <circle
                        key={`dot-${payload.snapshotId}`}
                        cx={cx}
                        cy={cy}
                        r={3.5}
                        fill="var(--text)"
                        stroke="var(--text)"
                        style={{ cursor: 'pointer' }}
                        onClick={() => setSelectedSnapshotId(payload.snapshotId)}
                      />
                    );
                  }}
                  activeDot={{
                    r: 6,
                    style: { cursor: 'pointer' },
                    onClick: (_, e) => {
                      const target = (e as unknown as { payload?: { snapshotId?: number } })
                        ?.payload;
                      if (target?.snapshotId !== undefined) {
                        setSelectedSnapshotId(target.snapshotId);
                      }
                    },
                  }}
                />
              </LineChart>
            )}
          </div>
        )}
      </HoverWrap>

      {selectedSnapshotId !== null && (
        <SnapshotModal
          snapshotId={selectedSnapshotId}
          onClose={() => setSelectedSnapshotId(null)}
        />
      )}

      {expandOpen && (
        <ChartExpandModal
          currentUserName={current?.name || current?.username || 'You'}
          onClose={() => setExpandOpen(false)}
        />
      )}
    </div>
  );
}

function StatBlock({ children }: { children: ReactNode }) {
  const [hover, setHover] = useState(false);
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        background: hover ? 'var(--bg-hover)' : 'var(--bg-subtle)',
        borderWidth: 1,
        borderStyle: 'solid',
        borderColor: hover ? 'var(--text-tertiary)' : 'var(--border)',
        borderRadius: 12,
        transition: 'background 0.15s ease, border-color 0.15s ease',
      }}
    >
      {children}
    </div>
  );
}

function RangeInput({
  initialStart,
  initialEnd,
  onApply,
}: {
  initialStart: string;
  initialEnd: string;
  onApply: (start: string, end: string) => void;
}) {
  const [start, setStart] = useState(initialStart);
  const [end, setEnd] = useState(initialEnd);

  useEffect(() => {
    setStart(initialStart);
    setEnd(initialEnd);
  }, [initialStart, initialEnd]);

  const valid = start && end && start <= end;
  const [hover, setHover] = useState(false);

  return (
    <div
      style={{
        display: 'flex',
        gap: 8,
        marginBottom: 14,
        flexWrap: 'wrap',
        alignItems: 'center',
      }}
    >
      <input
        type="date"
        value={start}
        onChange={(e) => setStart(e.target.value)}
        style={inputStyle()}
      />
      <span style={{ color: 'var(--text-tertiary)', fontSize: 12 }}>→</span>
      <input
        type="date"
        value={end}
        onChange={(e) => setEnd(e.target.value)}
        style={inputStyle()}
      />
      <button
        type="button"
        disabled={!valid}
        onClick={() => valid && onApply(start, end)}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        style={{
          padding: '6px 14px',
          borderRadius: 8,
          border: '1px solid var(--border)',
          background: !valid
            ? 'var(--bg-subtle)'
            : hover
              ? 'var(--text-secondary)'
              : 'var(--text)',
          color: !valid ? 'var(--text-tertiary)' : 'var(--bg)',
          fontSize: 12,
          fontWeight: 700,
          cursor: valid ? 'pointer' : 'not-allowed',
          transition: 'background 0.15s ease, color 0.15s ease',
        }}
      >
        Apply
      </button>
    </div>
  );
}
