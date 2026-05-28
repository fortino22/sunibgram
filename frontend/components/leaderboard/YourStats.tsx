'use client';

import { useMemo, useState, type CSSProperties, type ReactNode } from 'react';
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useMyTrendingCurrent } from '@/hooks/useMyTrendingCurrent';
import { useMyTrendingHistory } from '@/hooks/useMyTrendingHistory';
import { SnapshotModal } from './SnapshotModal';

function formatWeekLabel(weekStart: string): string {
  const d = new Date(weekStart);
  if (Number.isNaN(d.getTime())) return weekStart;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
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
        border: '1px solid',
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

export function YourStats() {
  const currentQuery = useMyTrendingCurrent();
  const [selectedYear, setSelectedYear] = useState<number | undefined>();
  const [selectedMonth, setSelectedMonth] = useState<number | undefined>();
  const historyQuery = useMyTrendingHistory(selectedYear, selectedMonth);
  const [selectedSnapshotId, setSelectedSnapshotId] = useState<number | null>(null);

  const current = currentQuery.data;
  const result = historyQuery.data;
  const history = useMemo(() => result?.data ?? [], [result]);
  const availableMonths = useMemo(() => result?.availableMonths ?? [], [result]);
  const activeYear = selectedYear ?? result?.year;
  const activeMonth = selectedMonth ?? result?.month;

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
          border: '1px solid',
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
          </div>
          {availableMonths.length > 0 && (
            <select
              value={activeYear && activeMonth ? `${activeYear}-${activeMonth}` : ''}
              onChange={(e) => {
                const [y, m] = e.target.value.split('-').map(Number);
                setSelectedYear(y);
                setSelectedMonth(m);
              }}
              style={{
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                border: '1px solid var(--border)',
                borderRadius: 8,
                padding: '6px 10px',
                fontSize: 12,
                fontWeight: 600,
                cursor: 'pointer',
                outline: 'none',
              }}
            >
              {availableMonths.map((m) => (
                <option key={`${m.year}-${m.month}`} value={`${m.year}-${m.month}`}>
                  {MONTH_NAMES[m.month - 1]} {m.year}
                </option>
              ))}
            </select>
          )}
        </div>
        {chartData.length === 0 ? (
          <div
            style={{
              padding: '32px 8px',
              textAlign: 'center',
              color: 'var(--text-tertiary)',
              fontSize: 13,
            }}
          >
            No history yet — your first snapshot drops next Monday.
          </div>
        ) : (
          <div style={{ width: '100%', height: 260 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 12, right: 16, left: 0, bottom: 4 }}>
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
            </ResponsiveContainer>
          </div>
        )}
      </HoverWrap>

      {selectedSnapshotId !== null && (
        <SnapshotModal
          snapshotId={selectedSnapshotId}
          onClose={() => setSelectedSnapshotId(null)}
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
        border: '1px solid',
        borderColor: hover ? 'var(--text-tertiary)' : 'var(--border)',
        borderRadius: 12,
        transition: 'background 0.15s ease, border-color 0.15s ease',
      }}
    >
      {children}
    </div>
  );
}
