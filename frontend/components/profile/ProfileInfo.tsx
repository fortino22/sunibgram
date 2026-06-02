'use client';

import { useState } from 'react';
import Link from 'next/link';
import { BookOpen, Calendar, ChevronRight } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { JoinedHoverCard } from '@/components/profile/JoinedHoverCard';
import { ProfileOptionsMenu } from '@/components/profile/ProfileOptionsMenu';
import { AvatarZoomModal } from '@/components/profile/AvatarZoomModal';

interface ProfileUser {
  name: string;
  username: string;
  bio?: string | null;
  faculty?: string | null;
  joinedMonth?: string;
  following: number;
  followers: number;
  avatarUrl?: string | null;
  isPremium?: boolean;
}

interface ProfileInfoProps {
  user: ProfileUser;
  userId: number;
  isOwn: boolean;
  isFollowing: boolean;
  onFollowToggle: () => void;
  followPending?: boolean;
  onEditClick?: () => void;
  onFollowersClick?: () => void;
  onFollowingClick?: () => void;
  onReport?: () => void;
  isBlocked?: boolean;
  isBlockedBy?: boolean;
  onBlock?: () => void;
  blockPending?: boolean;
  isMuted?: boolean;
  onMute?: () => void;
}

export function ProfileInfo({
  user,
  userId,
  isOwn,
  isFollowing,
  onFollowToggle,
  followPending = false,
  onEditClick,
  onFollowersClick,
  onFollowingClick,
  onReport,
  isBlocked = false,
  isBlockedBy = false,
  onBlock,
  blockPending = false,
  isMuted = false,
  onMute,
}: ProfileInfoProps) {
  const [zoomOpen, setZoomOpen] = useState(false);

  return (
    <div style={{ position: 'relative', padding: '0 20px' }}>
      {/* Avatar overlaps banner — click to zoom */}
      <button
        type="button"
        onClick={() => setZoomOpen(true)}
        aria-label={`View ${user.name}'s profile photo`}
        style={{
          position: 'absolute',
          top: -44,
          left: 20,
          zIndex: 1,
          padding: 0,
          border: 'none',
          background: 'none',
          borderRadius: '50%',
          cursor: 'zoom-in',
          fontFamily: 'inherit',
        }}
      >
        <Avatar
          username={user.name}
          avatarUrl={user.avatarUrl}
          size="lg"
          style={{ border: '4px solid var(--bg-card)' }}
        />
      </button>

      {/* Action buttons row */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, paddingTop: 10 }}>
        {isOwn ? (
          <button
            type="button"
            onClick={onEditClick}
            className="btn-ghost"
            style={{
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
            {onReport && (
              <ProfileOptionsMenu
                username={user.username}
                userId={userId}
                isBlocked={isBlocked}
                isMuted={isMuted}
                onBlock={onBlock}
                onMute={onMute}
                onReport={onReport}
              />
            )}
            {isBlockedBy ? (
              <button
                type="button"
                disabled
                style={{
                  background: 'none',
                  border: '1px solid var(--border-strong)',
                  borderRadius: 20,
                  padding: '7px 20px',
                  fontSize: 14,
                  fontWeight: 700,
                  color: 'var(--text-tertiary)',
                  fontFamily: 'inherit',
                  letterSpacing: '-0.01em',
                  opacity: 0.6,
                  cursor: 'not-allowed',
                }}
              >
                Follow
              </button>
            ) : isBlocked ? (
              <button
                type="button"
                onClick={onBlock}
                disabled={blockPending}
                className="hover-dim"
                style={{
                  background: 'none',
                  border: '1px solid var(--coral)',
                  borderRadius: 20,
                  padding: '7px 20px',
                  fontSize: 14,
                  fontWeight: 700,
                  color: 'var(--coral)',
                  fontFamily: 'inherit',
                  letterSpacing: '-0.01em',
                  transition: 'all 0.15s ease',
                  opacity: blockPending ? 0.7 : 1,
                  cursor: blockPending ? 'not-allowed' : 'pointer',
                }}
              >
                Blocked
              </button>
            ) : (
              <button
                type="button"
                onClick={onFollowToggle}
                disabled={followPending}
                className="hover-dim"
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
            )}
          </>
        )}
      </div>

      {/* User info */}
      <div style={{ marginTop: 10, paddingBottom: 16 }}>
        <div style={{ fontWeight: 700, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.03em', lineHeight: 1.2, display: 'flex', alignItems: 'center', gap: 6 }}>
          {user.name}
          {user.isPremium && <VerifiedBadge size={19} />}
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
            <JoinedHoverCard userId={userId} joinedMonth={user.joinedMonth}>
              <Link href={`/about-profile/${userId}`} className="group" style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, color: 'var(--text-tertiary)', textDecoration: 'none' }}>
                <Calendar size={15} strokeWidth={1.6} />
                <span className="group-hover:underline">Joined {user.joinedMonth}</span>
                <ChevronRight size={14} strokeWidth={1.8} />
              </Link>
            </JoinedHoverCard>
          )}
        </div>

        <div style={{ display: 'flex', gap: 20, marginTop: 14 }}>
          <button type="button" onClick={onFollowingClick} className="hover-underline" style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontFamily: 'inherit', display: 'flex', alignItems: 'baseline', gap: 5 }}>
            <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              {user.following.toLocaleString()}
            </span>
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Following</span>
          </button>
          <button type="button" onClick={onFollowersClick} className="hover-underline" style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontFamily: 'inherit', display: 'flex', alignItems: 'baseline', gap: 5 }}>
            <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              {user.followers.toLocaleString()}
            </span>
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Followers</span>
          </button>
        </div>
      </div>

      <AvatarZoomModal
        open={zoomOpen}
        onClose={() => setZoomOpen(false)}
        username={user.name}
        avatarUrl={user.avatarUrl}
      />
    </div>
  );
}
