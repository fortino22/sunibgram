'use client';

import { useState } from 'react';
import { useFollowUser } from '@/hooks/useFollowUser';

interface FollowButtonProps {
  userId: number;
  initialFollowing: boolean;
}

export function FollowButton({ userId, initialFollowing }: FollowButtonProps) {
  const [following, setFollowing] = useState(initialFollowing);
  const [hovered, setHovered] = useState(false);
  const { follow, unfollow } = useFollowUser(userId);

  function handleClick(e: React.MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    if (following) {
      setFollowing(false);
      unfollow.mutate(undefined, { onError: () => setFollowing(true) });
    } else {
      setFollowing(true);
      follow.mutate(undefined, { onError: () => setFollowing(false) });
    }
  }

  const label = following ? (hovered ? 'Unfollow' : 'Following') : 'Follow';

  const followingStyle: React.CSSProperties = hovered
    ? { background: 'transparent', color: '#f4212e', border: '1px solid rgba(244, 33, 46, 0.4)' }
    : { background: 'transparent', color: 'var(--text)', border: '1px solid var(--border)' };

  const notFollowingStyle: React.CSSProperties = {
    background: 'var(--text)',
    color: 'var(--bg)',
    border: '1px solid var(--text)',
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      className="shrink-0 rounded-full text-[14px] font-bold transition-colors"
      style={{ padding: '6px 16px', ...(following ? followingStyle : notFollowingStyle) }}
    >
      {label}
    </button>
  );
}
