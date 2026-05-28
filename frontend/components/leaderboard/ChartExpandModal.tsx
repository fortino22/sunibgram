'use client';

import { useEffect, useMemo, useState } from 'react';
import { createPortal } from 'react-dom';
import { X, Search, Plus } from 'lucide-react';
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useMeasuredSize } from '@/hooks/useMeasuredSize';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useMyTrendingHistory } from '@/hooks/useMyTrendingHistory';
import { useUserTrendingHistory } from '@/hooks/useUserTrendingHistory';
import { useMutualCompareSearch } from '@/hooks/useMutualCompareSearch';
import type { MutualCompareUser, MyTrendingHistoryItem } from '@/lib/services/trendingUsers';
import { SnapshotModal, type SnapshotOwnerHint } from './SnapshotModal';

const ALL_TIME_START = '1900-01-01';
const ALL_TIME_END = '2999-12-31';

const COMPARE_COLORS = [
  '#f59e0b',
  '#10b981',
  '#3b82f6',
  '#ec4899',
  '#a855f7',
  '#ef4444',
];

interface ChartExpandModalProps {
  currentUserName: string;
  onClose: () => void;
}

function formatWeekLabel(weekStart: string): string {
  const d = new Date(weekStart);
  if (Number.isNaN(d.getTime())) return weekStart;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: '2-digit' });
}

interface ChartPoint {
  weekStart: string;
  weekLabel: string;
  self?: number;
  selfSnapshotId?: number;
  [userKey: string]: string | number | undefined;
}

function mergePoints(
  selfHistory: MyTrendingHistoryItem[],
  compareSeries: { user: MutualCompareUser; history: MyTrendingHistoryItem[] }[],
): ChartPoint[] {
  const byWeek = new Map<string, ChartPoint>();

  for (const h of selfHistory) {
    byWeek.set(h.weekStart, {
      weekStart: h.weekStart,
      weekLabel: formatWeekLabel(h.weekStart),
      self: Number(h.score.toFixed(1)),
      selfSnapshotId: h.snapshotId,
    });
  }

  for (const series of compareSeries) {
    for (const h of series.history) {
      const existing = byWeek.get(h.weekStart);
      const key = `user_${series.user.userId}`;
      const snapKey = `${key}_snapshotId`;
      if (existing) {
        existing[key] = Number(h.score.toFixed(1));
        existing[snapKey] = h.snapshotId;
      } else {
        byWeek.set(h.weekStart, {
          weekStart: h.weekStart,
          weekLabel: formatWeekLabel(h.weekStart),
          [key]: Number(h.score.toFixed(1)),
          [snapKey]: h.snapshotId,
        });
      }
    }
  }

  return Array.from(byWeek.values()).sort((a, b) =>
    a.weekStart < b.weekStart ? -1 : a.weekStart > b.weekStart ? 1 : 0,
  );
}

interface SelectedSnapshot {
  snapshotId: number;
  ownerHint?: SnapshotOwnerHint;
}

export function ChartExpandModal({ currentUserName, onClose }: ChartExpandModalProps) {
  const [selectedSnapshot, setSelectedSnapshot] = useState<SelectedSnapshot | null>(null);
  const [compareUsers, setCompareUsers] = useState<MutualCompareUser[]>([]);
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchInput, setSearchInput] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(searchInput.trim()), 250);
    return () => clearTimeout(t);
  }, [searchInput]);

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        if (selectedSnapshot !== null) setSelectedSnapshot(null);
        else onClose();
      }
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose, selectedSnapshot]);

  const selfQuery = useMyTrendingHistory(ALL_TIME_START, ALL_TIME_END);
  const selfHistory = useMemo(() => selfQuery.data?.data ?? [], [selfQuery.data]);

  const [compareHistories, setCompareHistories] = useState<Record<number, MyTrendingHistoryItem[]>>({});

  function setCompareHistory(userId: number, history: MyTrendingHistoryItem[]) {
    setCompareHistories((prev) => {
      if (prev[userId] === history) return prev;
      return { ...prev, [userId]: history };
    });
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setCompareHistories((prev) => {
      const next: Record<number, MyTrendingHistoryItem[]> = {};
      for (const u of compareUsers) {
        if (prev[u.userId]) next[u.userId] = prev[u.userId];
      }
      return next;
    });
  }, [compareUsers]);

  const compareSeries = useMemo(
    () =>
      compareUsers.map((user) => ({
        user,
        history: compareHistories[user.userId] ?? [],
      })),
    [compareUsers, compareHistories],
  );

  const chartData = useMemo(
    () => mergePoints(selfHistory, compareSeries),
    [selfHistory, compareSeries],
  );

  const search = useMutualCompareSearch(debouncedQuery, searchOpen);

  const [chartRef, { w: chartW, h: chartH }] = useMeasuredSize<HTMLDivElement>();

  function addCompareUser(u: MutualCompareUser) {
    if (compareUsers.some((c) => c.userId === u.userId)) return;
    if (compareUsers.length >= COMPARE_COLORS.length) return;
    setCompareUsers((prev) => [...prev, u]);
    setSearchInput('');
    setSearchOpen(false);
  }

  function removeCompareUser(userId: number) {
    setCompareUsers((prev) => prev.filter((u) => u.userId !== userId));
  }

  const colorFor = (idx: number) => COMPARE_COLORS[idx % COMPARE_COLORS.length];

  return (
    <div
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 60,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
        animation: 'modal-backdrop-in 0.18s ease',
        padding: 16,
      }}
    >
      <div
        style={{
          position: 'relative',
          width: '100%',
          maxWidth: 1100,
          maxHeight: '92vh',
          overflowY: 'auto',
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          padding: 24,
          animation: 'modal-panel-in 0.22s cubic-bezier(0.16, 1, 0.3, 1)',
        }}
      >
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          style={{
            position: 'absolute',
            top: 12,
            right: 12,
            zIndex: 1,
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border)',
            borderRadius: '50%',
            width: 32,
            height: 32,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            color: 'var(--text)',
          }}
        >
          <X size={18} strokeWidth={2} />
        </button>

        <div style={{ marginBottom: 18 }}>
          <div
            style={{
              fontSize: 11,
              color: 'var(--text-tertiary)',
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
              marginBottom: 4,
            }}
          >
            All-Time History
          </div>
          <div style={{ fontSize: 20, fontWeight: 800, color: 'var(--text)' }}>
            Weekly Score — {currentUserName}
          </div>
        </div>

        <CompareBar
          compareUsers={compareUsers}
          colorFor={colorFor}
          onRemove={removeCompareUser}
          onOpenSearch={() => setSearchOpen((v) => !v)}
          searchOpen={searchOpen}
          atLimit={compareUsers.length >= COMPARE_COLORS.length}
        />

        {searchOpen && (
          <SearchPanel
            value={searchInput}
            onChange={setSearchInput}
            results={search.data ?? []}
            isLoading={search.isLoading}
            excludeIds={compareUsers.map((u) => u.userId)}
            onPick={addCompareUser}
          />
        )}

        {selfQuery.isLoading ? (
          <div
            style={{
              padding: '40px 0',
              textAlign: 'center',
              color: 'var(--text-tertiary)',
              fontSize: 13,
            }}
          >
            Loading history…
          </div>
        ) : chartData.length === 0 ? (
          <div
            style={{
              padding: '40px 0',
              textAlign: 'center',
              color: 'var(--text-tertiary)',
              fontSize: 13,
            }}
          >
            No snapshots yet.
          </div>
        ) : (
          <div
            ref={chartRef}
            style={{
              width: '100%',
              height: 460,
              minWidth: 0,
              position: 'relative',
              marginTop: 12,
            }}
          >
            {chartW > 0 && chartH > 0 && (
            <LineChart width={chartW} height={chartH} data={chartData} margin={{ top: 16, right: 24, left: 0, bottom: 4 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis
                  dataKey="weekLabel"
                  stroke="var(--text-tertiary)"
                  fontSize={11}
                  tickLine={false}
                  interval="preserveStartEnd"
                  minTickGap={32}
                />
                <YAxis
                  stroke="var(--text-tertiary)"
                  fontSize={11}
                  tickLine={false}
                  width={36}
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
                <Legend
                  wrapperStyle={{ fontSize: 12, color: 'var(--text)' }}
                  formatter={(value) => <span style={{ color: 'var(--text)' }}>{value}</span>}
                />
                <Line
                  type="monotone"
                  dataKey="self"
                  name={currentUserName}
                  stroke="var(--text)"
                  strokeWidth={2}
                  connectNulls
                  dot={(props: {
                    cx?: number;
                    cy?: number;
                    payload?: ChartPoint;
                    index?: number;
                  }) => {
                    const { cx, cy, payload, index } = props;
                    if (
                      cx === undefined ||
                      cy === undefined ||
                      !payload ||
                      payload.selfSnapshotId === undefined
                    ) {
                      return <g key={`dot-empty-${index ?? 0}`} />;
                    }
                    return (
                      <circle
                        key={`dot-${payload.selfSnapshotId}`}
                        cx={cx}
                        cy={cy}
                        r={4}
                        fill="var(--text)"
                        stroke="var(--text)"
                        style={{ cursor: 'pointer' }}
                        onClick={() =>
                          setSelectedSnapshot({
                            snapshotId: payload.selfSnapshotId as number,
                          })
                        }
                      />
                    );
                  }}
                  activeDot={(props: {
                    cx?: number;
                    cy?: number;
                    payload?: ChartPoint;
                  }) => {
                    const { cx, cy, payload } = props;
                    const snapId = payload?.selfSnapshotId;
                    if (cx === undefined || cy === undefined || typeof snapId !== 'number') {
                      return <g />;
                    }
                    return (
                      <circle
                        cx={cx}
                        cy={cy}
                        r={6}
                        fill="var(--text)"
                        stroke="var(--text)"
                        style={{ cursor: 'pointer' }}
                        onClick={() => setSelectedSnapshot({ snapshotId: snapId })}
                      />
                    );
                  }}
                />
                {compareUsers.map((u, i) => {
                  const color = colorFor(i);
                  const dataKey = `user_${u.userId}`;
                  const snapKey = `${dataKey}_snapshotId`;
                  return (
                    <Line
                      key={u.userId}
                      type="monotone"
                      dataKey={dataKey}
                      name={`@${u.username}`}
                      stroke={color}
                      strokeWidth={2}
                      connectNulls
                      dot={(props: {
                        cx?: number;
                        cy?: number;
                        payload?: ChartPoint;
                        index?: number;
                      }) => {
                        const { cx, cy, payload, index } = props;
                        const snapId = payload?.[snapKey];
                        if (
                          cx === undefined ||
                          cy === undefined ||
                          !payload ||
                          typeof snapId !== 'number'
                        ) {
                          return <g key={`dot-empty-${u.userId}-${index ?? 0}`} />;
                        }
                        return (
                          <circle
                            key={`dot-${u.userId}-${snapId}`}
                            cx={cx}
                            cy={cy}
                            r={3.5}
                            fill={color}
                            stroke={color}
                            style={{ cursor: 'pointer' }}
                            onClick={() =>
                              setSelectedSnapshot({
                                snapshotId: snapId,
                                ownerHint: {
                                  username: u.username,
                                  name: u.name,
                                  profile: u.profile,
                                  isPremium: u.isPremium,
                                },
                              })
                            }
                          />
                        );
                      }}
                      activeDot={(props: {
                        cx?: number;
                        cy?: number;
                        payload?: ChartPoint;
                      }) => {
                        const { cx, cy, payload } = props;
                        const snapId = payload?.[snapKey];
                        if (
                          cx === undefined ||
                          cy === undefined ||
                          typeof snapId !== 'number'
                        ) {
                          return <g />;
                        }
                        return (
                          <circle
                            cx={cx}
                            cy={cy}
                            r={5}
                            fill={color}
                            stroke={color}
                            style={{ cursor: 'pointer' }}
                            onClick={() =>
                              setSelectedSnapshot({
                                snapshotId: snapId,
                                ownerHint: {
                                  username: u.username,
                                  name: u.name,
                                  profile: u.profile,
                                  isPremium: u.isPremium,
                                },
                              })
                            }
                          />
                        );
                      }}
                    />
                  );
                })}
              </LineChart>
            )}
          </div>
        )}

        <div
          style={{
            marginTop: 14,
            fontSize: 11,
            color: 'var(--text-tertiary)',
          }}
        >
          Click your own dot to open the snapshot detail. Add mutual friends above to compare.
        </div>
      </div>

      {compareUsers.map((u) => (
        <CompareLineLoader
          key={u.userId}
          userId={u.userId}
          onLoad={(history) => setCompareHistory(u.userId, history)}
        />
      ))}

      {selectedSnapshot !== null &&
        typeof document !== 'undefined' &&
        createPortal(
          <SnapshotModal
            snapshotId={selectedSnapshot.snapshotId}
            ownerHint={selectedSnapshot.ownerHint}
            onClose={() => setSelectedSnapshot(null)}
          />,
          document.body,
        )}
    </div>
  );
}

function CompareLineLoader({
  userId,
  onLoad,
}: {
  userId: number;
  onLoad: (history: MyTrendingHistoryItem[]) => void;
}) {
  const query = useUserTrendingHistory(userId, ALL_TIME_START, ALL_TIME_END);
  useEffect(() => {
    if (query.data?.data) onLoad(query.data.data);
  }, [query.data, onLoad]);
  return null;
}

function CompareBar({
  compareUsers,
  colorFor,
  onRemove,
  onOpenSearch,
  searchOpen,
  atLimit,
}: {
  compareUsers: MutualCompareUser[];
  colorFor: (i: number) => string;
  onRemove: (userId: number) => void;
  onOpenSearch: () => void;
  searchOpen: boolean;
  atLimit: boolean;
}) {
  return (
    <div
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: 8,
        alignItems: 'center',
        marginBottom: 8,
      }}
    >
      {compareUsers.map((u, i) => (
        <div
          key={u.userId}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: '6px 10px 6px 6px',
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border)',
            borderRadius: 9999,
            fontSize: 12,
            color: 'var(--text)',
          }}
        >
          <span
            style={{
              width: 10,
              height: 10,
              borderRadius: '50%',
              background: colorFor(i),
              display: 'inline-block',
            }}
          />
          <Avatar username={u.username} avatarUrl={u.profile || null} size={20} />
          <span style={{ fontWeight: 600 }}>@{u.username}</span>
          {u.isPremium && <VerifiedBadge size={12} />}
          <button
            type="button"
            onClick={() => onRemove(u.userId)}
            aria-label={`Remove ${u.username}`}
            style={{
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-tertiary)',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <X size={14} strokeWidth={2} />
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={onOpenSearch}
        disabled={atLimit}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          padding: '6px 12px',
          background: searchOpen ? 'var(--text)' : 'var(--bg-subtle)',
          color: searchOpen ? 'var(--bg)' : 'var(--text)',
          border: '1px solid var(--border)',
          borderRadius: 9999,
          fontSize: 12,
          fontWeight: 700,
          cursor: atLimit ? 'not-allowed' : 'pointer',
          opacity: atLimit ? 0.5 : 1,
        }}
      >
        <Plus size={14} strokeWidth={2.5} />
        {atLimit ? 'Limit reached' : 'Add mutual'}
      </button>
    </div>
  );
}

function SearchResultRow({
  user,
  onPick,
}: {
  user: MutualCompareUser;
  onPick: (u: MutualCompareUser) => void;
}) {
  const [hover, setHover] = useState(false);
  return (
    <button
      key={user.userId}
      type="button"
      onClick={() => onPick(user)}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        padding: '8px 10px',
        background: hover ? 'var(--bg-hover)' : 'transparent',
        border: '1px solid',
        borderColor: hover ? 'var(--text-tertiary)' : 'transparent',
        borderRadius: 8,
        cursor: 'pointer',
        textAlign: 'left',
        color: 'var(--text)',
        transform: hover ? 'translateX(4px)' : 'translateX(0)',
        transition:
          'background 0.15s ease, border-color 0.15s ease, transform 0.15s ease',
      }}
    >
      <Avatar username={user.username} avatarUrl={user.profile || null} size={28} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 4,
            fontSize: 13,
            fontWeight: 600,
          }}
        >
          {user.name || user.username}
          {user.isPremium && <VerifiedBadge size={12} />}
        </div>
        <div style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>@{user.username}</div>
      </div>
      <Plus size={16} strokeWidth={2} color="var(--text-tertiary)" />
    </button>
  );
}

function SearchPanel({
  value,
  onChange,
  results,
  isLoading,
  excludeIds,
  onPick,
}: {
  value: string;
  onChange: (v: string) => void;
  results: MutualCompareUser[];
  isLoading: boolean;
  excludeIds: number[];
  onPick: (u: MutualCompareUser) => void;
}) {
  const [focused, setFocused] = useState(false);
  const filtered = results.filter((r) => !excludeIds.includes(r.userId));

  return (
    <div
      style={{
        marginBottom: 12,
        maxWidth: 360,
        border: '1px solid var(--border)',
        borderRadius: 12,
        background: 'var(--bg-subtle)',
        padding: 10,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          background: 'var(--bg-card)',
          border: '1px solid',
          borderColor: focused ? '#3b82f6' : 'var(--border)',
          boxShadow: focused ? '0 0 0 3px rgba(59, 130, 246, 0.25)' : 'none',
          borderRadius: 9999,
          padding: '5px 10px',
          marginBottom: 8,
          transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
        }}
      >
        <Search size={12} color={focused ? '#3b82f6' : 'var(--text-tertiary)'} />
        <input
          autoFocus
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          placeholder="Search mutuals"
          style={{
            flex: 1,
            minWidth: 0,
            background: 'transparent',
            border: 'none',
            outline: 'none',
            color: 'var(--text)',
            fontSize: 12,
          }}
        />
      </div>

      {isLoading && (
        <div style={{ padding: 10, fontSize: 12, color: 'var(--text-tertiary)' }}>
          Searching…
        </div>
      )}
      {!isLoading && filtered.length === 0 && (
        <div style={{ padding: 10, fontSize: 12, color: 'var(--text-tertiary)' }}>
          {value ? 'No mutual friends match that search.' : 'Type to search your mutuals.'}
        </div>
      )}
      {!isLoading && filtered.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4, maxHeight: 240, overflowY: 'auto' }}>
          {filtered.map((u) => (
            <SearchResultRow key={u.userId} user={u} onPick={onPick} />
          ))}
        </div>
      )}
    </div>
  );
}

