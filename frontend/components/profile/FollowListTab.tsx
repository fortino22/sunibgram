'use client';

import { useRef, useEffect, type MouseEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { EmptyState } from '@/components/shared/EmptyState';
import { useFollowers } from '@/hooks/useFollowers';
import { useFollowing } from '@/hooks/useFollowing';
import { useFollowUser } from '@/hooks/useFollowUser';
import { useProfile } from '@/hooks/useProfile';
import type { FollowUser } from '@/lib/services/follows';
import { useAuthStore } from '@/stores/authStore';

type FollowListType = 'followers' | 'following';

interface FollowListTabProps {
  userId: number | string;
  type: FollowListType;
  filterPremium?: boolean;
}

interface FollowUserCardProps {
  user: FollowUser;
  initialIsFollowing?: boolean;
}

export function FollowUserCard({ user, initialIsFollowing }: FollowUserCardProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { data: userProfile } = useProfile(user.user_id);
  const { follow, unfollow } = useFollowUser(user.user_id);
  const isOwnProfile = currentUserId === user.user_id;
  const hasResolvedFollowState = initialIsFollowing !== undefined || userProfile !== undefined;
  const isFollowing = initialIsFollowing ?? Boolean(userProfile?.is_following);
  const isPending = follow.isPending || unfollow.isPending;

  function handleFollowToggle(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation();
    if (isPending) return;
    if (isFollowing) {
      unfollow.mutate();
    } else {
      follow.mutate();
    }
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => router.push(`/profile/${user.user_id}`)}
      onKeyDown={(event) => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          router.push(`/profile/${user.user_id}`);
        }
      }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
        cursor: 'pointer',
      }}
    >
      <Avatar username={user.username} avatarUrl={user.avatar_url} size="md" />

      <div style={{ minWidth: 0, flex: 1 }}>
        <div
          style={{
            fontSize: 15,
            fontWeight: 700,
            color: 'var(--text)',
            letterSpacing: '-0.02em',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {user.name ?? user.username}
        </div>
        <div
          style={{
            fontSize: 14,
            color: 'var(--text-tertiary)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          @{user.username}
        </div>
      </div>

      {!isOwnProfile && (
        <button
          type="button"
          onClick={handleFollowToggle}
          disabled={isPending || !hasResolvedFollowState}
          style={{
            background: isFollowing ? 'none' : 'var(--text)',
            border: isFollowing ? '1px solid var(--border-strong)' : '1px solid transparent',
            borderRadius: 999,
            padding: '7px 16px',
            minWidth: 92,
            fontSize: 14,
            fontWeight: 700,
            color: isFollowing ? 'var(--text)' : 'var(--bg)',
            cursor: isPending || !hasResolvedFollowState ? 'not-allowed' : 'pointer',
            fontFamily: 'inherit',
            letterSpacing: '-0.01em',
            opacity: isPending || !hasResolvedFollowState ? 0.7 : 1,
            transition: 'all 0.15s ease',
          }}
        >
          {hasResolvedFollowState ? (isFollowing ? 'Unfollow' : 'Follow') : 'Loading…'}
        </button>
      )}
    </div>
  );
}

export function FollowListTab({ userId, type, filterPremium = false }: FollowListTabProps) {
  const currentUserId = useAuthStore((state) => state.user?.id);
  const followersQuery = useFollowers(userId);
  const followingQuery = useFollowing(userId);
  const query = type === 'followers' ? followersQuery : followingQuery;
  const sentinelRef = useRef<HTMLDivElement>(null);

  const allUsers = query.data?.pages.flatMap((p) => p.data) ?? [];
  const users = filterPremium ? allUsers.filter((u) => u.is_premium) : allUsers;
  const isOwnFollowingList = type === 'following' && currentUserId === Number(userId);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && query.hasNextPage && !query.isFetchingNextPage) {
          void query.fetchNextPage();
        }
      },
      { rootMargin: '200px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [query]);

  if (query.isLoading) {
    return (
      <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
        Loading…
      </div>
    );
  }

  if (query.isError) {
    return (
      <div style={{ padding: 24, textAlign: 'center', color: 'var(--coral)', fontSize: 14 }}>
        {query.error.message}
      </div>
    );
  }

  if (users.length === 0) {
    return (
      <EmptyState
        title={filterPremium ? 'No verified followers' : `No ${type} yet`}
        subtitle={
          filterPremium
            ? 'Followers with premium accounts will appear here.'
            : type === 'followers'
              ? 'When people follow this profile, they will appear here.'
              : 'Accounts followed by this profile will appear here.'
        }
      />
    );
  }

  return (
    <div>
      {users.map((user) => (
        <FollowUserCard
          key={user.user_id}
          user={user}
          initialIsFollowing={isOwnFollowingList ? true : undefined}
        />
      ))}
      <div ref={sentinelRef} style={{ height: 1 }} />
      {query.isFetchingNextPage && (
        <div style={{ padding: 16, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
          Loading…
        </div>
      )}
    </div>
  );
}
