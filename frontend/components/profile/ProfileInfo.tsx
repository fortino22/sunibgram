'use client';

import { BookOpen, Calendar, MoreHorizontal } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';

interface ProfileUser {
  name: string;
  username: string;
  bio?: string | null;
  faculty?: string | null;
  joinedMonth?: string;
  following: number;
  followers: number;
  avatarUrl?: string | null;
}

interface ProfileInfoProps {
  user: ProfileUser;
  isOwn: boolean;
  isFollowing: boolean;
  onFollowToggle: () => void;
  followPending?: boolean;
  onEditClick?: () => void;
  onFollowersClick?: () => void;
  onFollowingClick?: () => void;
}

export function ProfileInfo({
  user,
  isOwn,
  isFollowing,
  onFollowToggle,
  followPending = false,
  onEditClick,
  onFollowersClick,
  onFollowingClick,
}: ProfileInfoProps) {
  return (
    <div style={{ position: 'relative', padding: '0 20px' }}>
      {/* Avatar overlaps banner */}
      <Avatar
        username={user.name}
        avatarUrl={user.avatarUrl}
        size="lg"
        style={{
          position: 'absolute',
          top: -44,
          left: 20,
          border: '4px solid var(--bg-card)',
          zIndex: 1,
        }}
      />

      {/* Action buttons row */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, paddingTop: 10 }}>
        {isOwn ? (
          <button
            type="button"
            onClick={onEditClick}
            style={{
              background: 'none',
              border: '1px solid var(--border-strong)',
              borderRadius: 20,
              padding: '7px 18px',
              fontSize: 14,
              fontWeight: 600,
              color: 'var(--text)',
              cursor: 'pointer',
              fontFamily: 'inherit',
              letterSpacing: '-0.01em',
            }}
          >
            Edit profile
          </button>
        ) : (
          <>
            <button
              type="button"
              aria-label="More options"
              style={{
                background: 'none',
                border: '1px solid var(--border-strong)',
                borderRadius: '50%',
                width: 36,
                height: 36,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                color: 'var(--text)',
              }}
            >
              <MoreHorizontal size={17} strokeWidth={1.8} />
            </button>
            <button
              type="button"
              onClick={onFollowToggle}
              disabled={followPending}
              style={{
                background: isFollowing ? 'none' : 'var(--text)',
                border: isFollowing ? '1px solid var(--border-strong)' : '1px solid transparent',
                borderRadius: 20,
                padding: '7px 20px',
                fontSize: 14,
                fontWeight: 700,
                color: isFollowing ? 'var(--text)' : 'var(--bg)',
                fontFamily: 'inherit',
                letterSpacing: '-0.01em',
                transition: 'all 0.15s ease',
                opacity: followPending ? 0.7 : 1,
                cursor: followPending ? 'not-allowed' : 'pointer',
              }}
            >
              {isFollowing ? 'Unfollow' : 'Follow'}
            </button>
          </>
        )}
      </div>

      {/* User info */}
      <div style={{ marginTop: 10, paddingBottom: 16 }}>
        <div style={{ fontWeight: 700, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.03em', lineHeight: 1.2 }}>
          {user.name}
        </div>
        <div style={{ fontSize: 14, color: 'var(--text-tertiary)', marginTop: 2 }}>
          @{user.username}
        </div>

        {user.bio && (
          <div style={{ fontSize: 15, color: 'var(--text)', lineHeight: 1.65, marginTop: 12 }}>
            {user.bio}
          </div>
        )}

        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px 20px', marginTop: 12 }}>
          {user.faculty && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, color: 'var(--text-tertiary)' }}>
              <BookOpen size={15} strokeWidth={1.6} />
              {user.faculty}
            </span>
          )}
          {user.joinedMonth && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, color: 'var(--text-tertiary)' }}>
              <Calendar size={15} strokeWidth={1.6} />
              Joined {user.joinedMonth}
            </span>
          )}
        </div>

        <div style={{ display: 'flex', gap: 20, marginTop: 14 }}>
          <button type="button" onClick={onFollowingClick} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontFamily: 'inherit', display: 'flex', alignItems: 'baseline', gap: 5 }}>
            <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              {user.following.toLocaleString()}
            </span>
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Following</span>
          </button>
          <button type="button" onClick={onFollowersClick} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontFamily: 'inherit', display: 'flex', alignItems: 'baseline', gap: 5 }}>
            <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              {user.followers.toLocaleString()}
            </span>
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Followers</span>
          </button>
        </div>
      </div>
    </div>
  );
}
