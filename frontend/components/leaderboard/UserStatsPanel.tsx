'use client';

import { useState } from 'react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import type { TrendingUser } from '@/lib/services/trendingUsers';

interface UserStatsPanelProps {
  user: TrendingUser;
  rank?: number;
}

interface StatCellProps {
  label: string;
  value: number;
}

function ScoreBlock({ score }: { score: number }) {
  const [hovered, setHovered] = useState(false);
  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 2,
        padding: '12px 0',
        background: hovered ? 'var(--bg-hover)' : 'var(--bg-subtle)',
        borderRadius: 12,
        border: `1px solid ${hovered ? 'var(--border-strong)' : 'var(--border)'}`,
        transition: 'background 0.15s ease, border-color 0.15s ease, transform 0.15s ease',
        transform: hovered ? 'translateY(-1px)' : 'translateY(0)',
        cursor: 'default',
      }}
    >
      <span style={{ fontSize: 11, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
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
        {score.toFixed(1)}
      </span>
    </div>
  );
}

function StatCell({ label, value }: StatCellProps) {
  const [hovered, setHovered] = useState(false);
  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: hovered ? 'var(--bg-hover)' : 'var(--bg-subtle)',
        border: `1px solid ${hovered ? 'var(--border-strong)' : 'var(--border)'}`,
        borderRadius: 10,
        padding: '10px 12px',
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
        transition: 'background 0.15s ease, border-color 0.15s ease, transform 0.15s ease',
        transform: hovered ? 'translateY(-1px)' : 'translateY(0)',
        cursor: 'default',
      }}
    >
      <span style={{ fontSize: 11, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>
        {label}
      </span>
      <span style={{ fontSize: 18, fontWeight: 700, color: 'var(--text)' }}>{value}</span>
    </div>
  );
}

export function UserStatsPanel({ user, rank }: UserStatsPanelProps) {
  return (
    <div
      style={{
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 16,
        padding: 20,
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
        <Avatar username={user.name} avatarUrl={user.profile || null} size={64} />
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
              {user.name}
            </span>
            {user.isPremium && <VerifiedBadge size={15} />}
          </div>
          <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{user.username}</div>
          {rank && (
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
              Rank #{rank}
            </span>
          )}
        </div>
      </div>

      <ScoreBlock score={user.score} />
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
          Engagement Received
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
          <StatCell label="Likes" value={user.received.likes} />
          <StatCell label="Comments" value={user.received.comments} />
          <StatCell label="Bookmarks" value={user.received.bookmarks} />
          <StatCell label="Reposts" value={user.received.reposts} />
          <StatCell label="Threads" value={user.received.threadCount} />
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
          <StatCell label="Likes" value={user.given.likes} />
          <StatCell label="Comments" value={user.given.comments} />
          <StatCell label="Bookmarks" value={user.given.bookmarks} />
          <StatCell label="Posts" value={user.given.posts} />
        </div>
      </div>
    </div>
  );
}
