'use client';

import { useState } from 'react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import type { TrendingUser } from '@/lib/services/trendingUsers';

interface LeaderboardPodiumProps {
  top3: TrendingUser[];
  onSelect: (userId: number) => void;
}

interface PodiumColumnProps {
  user: TrendingUser | undefined;
  rank: 1 | 2 | 3;
  onSelect: (userId: number) => void;
}

const RANK_HEIGHTS: Record<1 | 2 | 3, number> = { 1: 110, 2: 80, 3: 64 };
const RANK_LABELS: Record<1 | 2 | 3, string> = { 1: '1', 2: '2', 3: '3' };

function PodiumColumn({ user, rank, onSelect }: PodiumColumnProps) {
  const [hovered, setHovered] = useState(false);
  const isFirst = rank === 1;
  const avatarSize = isFirst ? 96 : 76;

  if (!user) {
    return (
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          opacity: 0.25,
        }}
      >
        <div style={{ width: avatarSize, height: avatarSize, borderRadius: '50%', background: 'var(--bg-subtle)' }} />
        <div style={{ marginTop: 8, height: 16 }} />
        <div
          style={{
            marginTop: 12,
            height: RANK_HEIGHTS[rank],
            width: '100%',
            maxWidth: 110,
            borderRadius: '12px 12px 0 0',
            background: 'var(--bg-subtle)',
          }}
        />
      </div>
    );
  }

  const pedestalBg = isFirst
    ? 'var(--text)'
    : hovered
      ? 'var(--bg-hover)'
      : 'var(--bg-subtle)';
  const pedestalColor = isFirst ? 'var(--bg)' : 'var(--text)';
  const ringColor = isFirst
    ? hovered
      ? 'var(--text)'
      : 'var(--border-strong)'
    : hovered
      ? 'var(--text)'
      : 'var(--border)';

  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <button
        type="button"
        onClick={() => onSelect(user.userId)}
        aria-label={`Show stats for ${user.name}`}
        style={{
          background: 'none',
          border: 'none',
          padding: 0,
          cursor: 'pointer',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 10,
          transform: hovered ? 'translateY(-2px)' : 'translateY(0)',
          transition: 'transform 0.18s ease',
        }}
      >
        <div
          style={{
            padding: 3,
            borderRadius: '50%',
            background: ringColor,
            transition: 'background 0.18s ease',
          }}
        >
          <Avatar username={user.name} avatarUrl={user.profile || null} size={avatarSize} />
        </div>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 4,
            fontWeight: 700,
            fontSize: 14,
            color: 'var(--text)',
            maxWidth: 140,
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{user.name}</span>
          {user.isPremium && <VerifiedBadge size={13} />}
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{user.username}</div>
      </button>
      <div
        style={{
          marginTop: 12,
          height: RANK_HEIGHTS[rank],
          width: '100%',
          maxWidth: 110,
          borderRadius: '12px 12px 0 0',
          background: pedestalBg,
          border: '1px solid var(--border)',
          borderBottom: 'none',
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'center',
          paddingTop: 14,
          color: pedestalColor,
          fontWeight: 800,
          fontSize: 28,
          transition: 'background 0.18s ease',
        }}
      >
        {RANK_LABELS[rank]}
      </div>
    </div>
  );
}

export function LeaderboardPodium({ top3, onSelect }: LeaderboardPodiumProps) {
  const [first, second, third] = top3;

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'flex-end',
        gap: 12,
        padding: '28px 16px 0',
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 16,
        overflow: 'hidden',
      }}
    >
      <PodiumColumn user={second} rank={2} onSelect={onSelect} />
      <PodiumColumn user={first} rank={1} onSelect={onSelect} />
      <PodiumColumn user={third} rank={3} onSelect={onSelect} />
    </div>
  );
}
