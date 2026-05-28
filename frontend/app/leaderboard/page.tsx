'use client';

import { useMemo, useState } from 'react';
import { useTrendingUsers } from '@/hooks/useTrendingUsers';
import { LeaderboardPodium } from '@/components/leaderboard/LeaderboardPodium';
import { LeaderboardRow } from '@/components/leaderboard/LeaderboardRow';
import { UserStatsModal } from '@/components/leaderboard/UserStatsModal';

export default function LeaderboardPage() {
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const { data, isLoading, isError } = useTrendingUsers(1, 20);

  const users = useMemo(() => data?.data ?? [], [data]);
  const top3 = users.slice(0, 3);

  const userById = useMemo(() => {
    const map = new Map<number, { user: typeof users[number]; rank: number }>();
    users.forEach((u, i) => map.set(u.userId, { user: u, rank: i + 1 }));
    return map;
  }, [users]);

  const selected = selectedUserId !== null ? userById.get(selectedUserId) : undefined;

  return (
    <div style={{ padding: '24px 24px 80px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <header
        style={{
          display: 'flex',
          alignItems: 'flex-end',
          justifyContent: 'space-between',
          gap: 12,
          margin: '0 -24px',
          padding: '0 24px 16px',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div>
          <h1
            style={{
              fontSize: 26,
              fontWeight: 800,
              color: 'var(--text)',
              letterSpacing: '-0.02em',
            }}
          >
            Leaderboard
          </h1>
          <p style={{ marginTop: 4, fontSize: 13, color: 'var(--text-tertiary)' }}>
            Top contributors ranked by weekly engagement
          </p>
        </div>
        <span
          style={{
            padding: '6px 14px',
            borderRadius: 9999,
            background: 'var(--text)',
            color: 'var(--bg)',
            fontSize: 12,
            fontWeight: 700,
            letterSpacing: '0.02em',
          }}
        >
          This Week
        </span>
      </header>

      {isLoading && (
        <div
          style={{
            padding: 48,
            textAlign: 'center',
            color: 'var(--text-tertiary)',
            fontSize: 14,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 16,
          }}
        >
          Loading leaderboard…
        </div>
      )}

      {isError && !isLoading && (
        <div
          style={{
            padding: 48,
            textAlign: 'center',
            color: 'var(--coral)',
            fontSize: 14,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 16,
          }}
        >
          Failed to load leaderboard. Try again later.
        </div>
      )}

      {!isLoading && !isError && users.length === 0 && (
        <div
          style={{
            padding: 48,
            textAlign: 'center',
            color: 'var(--text-tertiary)',
            fontSize: 14,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 16,
          }}
        >
          No trending users this week. Be the first to post or engage.
        </div>
      )}

      {!isLoading && users.length > 0 && (
        <>
          <LeaderboardPodium top3={top3} onSelect={setSelectedUserId} />

          <div
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 16,
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                padding: '14px 16px',
                borderBottom: '1px solid var(--border)',
                fontWeight: 700,
                fontSize: 14,
                color: 'var(--text)',
              }}
            >
              All ranks
            </div>
            {users.map((user, i) => (
              <LeaderboardRow
                key={user.userId}
                rank={i + 1}
                user={user}
                highlighted={i === 0}
                isLast={i === users.length - 1}
                onSelect={setSelectedUserId}
              />
            ))}
          </div>
        </>
      )}

      {selected && (
        <UserStatsModal
          user={selected.user}
          rank={selected.rank}
          onClose={() => setSelectedUserId(null)}
        />
      )}
    </div>
  );
}
