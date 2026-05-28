'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useAuthStore } from '@/stores/authStore';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { FollowButton } from '@/components/shared/FollowButton';
import { UserHoverCard } from '@/components/shared/UserHoverCard';
import type { TrendingUser } from '@/lib/services/trendingUsers';

interface LeaderboardRowProps {
  rank: number;
  user: TrendingUser;
  highlighted?: boolean;
  isLast?: boolean;
  onSelect: (userId: number) => void;
}

export function LeaderboardRow({
  rank,
  user,
  highlighted = false,
  isLast = false,
  onSelect,
}: LeaderboardRowProps) {
  const [hovered, setHovered] = useState(false);
  const currentUserId = useAuthStore((s) => s.user?.id);
  const isSelf = currentUserId === user.userId;

  const bg = hovered
    ? 'var(--bg-hover)'
    : highlighted
      ? 'var(--bg-subtle)'
      : 'transparent';

  function handleRowClick() {
    onSelect(user.userId);
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={handleRowClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleRowClick();
        }
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'grid',
        gridTemplateColumns: '40px 1fr auto auto',
        alignItems: 'center',
        gap: 12,
        width: '100%',
        padding: '14px 16px',
        background: bg,
        borderBottom: isLast ? 'none' : '1px solid var(--border)',
        transition: 'background 0.15s ease',
        cursor: 'pointer',
      }}
    >
      <span
        style={{
          fontWeight: 700,
          fontSize: 14,
          color: hovered ? 'var(--text)' : 'var(--text-tertiary)',
          textAlign: 'center',
          transition: 'color 0.15s ease',
        }}
      >
        {rank}
      </span>

      <UserHoverCard userId={user.userId} style={{ justifySelf: 'start', maxWidth: '100%' }}>
        <Link
          href={`/profile/${user.userId}`}
          onClick={(e) => e.stopPropagation()}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 12,
            minWidth: 0,
            maxWidth: '100%',
            textDecoration: 'none',
            color: 'inherit',
          }}
        >
          <Avatar username={user.name} avatarUrl={user.profile || null} size="md" />
          <div style={{ minWidth: 0 }}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 4,
                fontWeight: 600,
                fontSize: 14,
                color: 'var(--text)',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              <span
                className="hover-underline"
                style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}
              >
                {user.name}
              </span>
              {user.isPremium && <VerifiedBadge size={13} />}
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{user.username}</div>
          </div>
        </Link>
      </UserHoverCard>

      <span
        style={{
          fontSize: 15,
          fontWeight: 700,
          color: 'var(--text)',
          fontVariantNumeric: 'tabular-nums',
          padding: '4px 8px',
        }}
      >
        {user.score.toFixed(1)}
      </span>

      {isSelf ? (
        <span style={{ width: 1 }} />
      ) : (
        <div onClick={(e) => e.stopPropagation()}>
          <FollowButton userId={user.userId} initialFollowing={user.isFollowing} />
        </div>
      )}
    </div>
  );
}
