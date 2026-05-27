'use client';

import { useCallback, useRef, useState, type CSSProperties, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import Link from 'next/link';
import { Avatar } from './Avatar';
import { VerifiedBadge } from './VerifiedBadge';
import { FollowButton } from './FollowButton';
import { useProfile } from '@/hooks/useProfile';
import { useAuthStore } from '@/stores/authStore';

const OPEN_DELAY = 350;
const CLOSE_DELAY = 200;
const CARD_WIDTH = 300;
const EST_HEIGHT = 200;

interface UserHoverCardProps {
  userId: number | string;
  /** The trigger element (avatar, name link, …). */
  children: ReactNode;
  /** Styles merged onto the inline wrapper (e.g. flexShrink for a flex child). */
  style?: CSSProperties;
}

type Placement = { left: number; top: number };

/**
 * Reusable X-style profile preview popover. Wrap any trigger (avatar, name)
 * with it; on hover-intent it lazily fetches the profile and shows a card.
 */
export function UserHoverCard({ userId, children, style }: UserHoverCardProps) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState<Placement | null>(null);
  const wrapRef = useRef<HTMLSpanElement>(null);
  const openTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const closeTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const currentUserId = useAuthStore((s) => s.user?.id);
  const isSelf = currentUserId != null && String(currentUserId) === String(userId);

  // Lazy: only fetch once the card is opening.
  const { data: profile, isLoading } = useProfile(open ? userId : null);

  const computePos = useCallback(() => {
    const el = wrapRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const placeBelow = r.bottom + EST_HEIGHT + 12 < window.innerHeight;
    const top = placeBelow ? r.bottom + 8 : Math.max(8, r.top - EST_HEIGHT - 8);
    const left = Math.max(8, Math.min(r.left, window.innerWidth - CARD_WIDTH - 8));
    setPos({ left, top });
  }, []);

  const scheduleOpen = useCallback(() => {
    clearTimeout(closeTimer.current);
    openTimer.current = setTimeout(() => {
      computePos();
      setOpen(true);
    }, OPEN_DELAY);
  }, [computePos]);

  const scheduleClose = useCallback(() => {
    clearTimeout(openTimer.current);
    closeTimer.current = setTimeout(() => setOpen(false), CLOSE_DELAY);
  }, []);

  return (
    <span
      ref={wrapRef}
      onMouseEnter={scheduleOpen}
      onMouseLeave={scheduleClose}
      style={{ display: 'inline-flex', position: 'relative', ...style }}
    >
      {children}

      {open && pos && typeof document !== 'undefined' &&
        createPortal(
          <div
            onMouseEnter={() => clearTimeout(closeTimer.current)}
            onMouseLeave={scheduleClose}
            onClick={(e) => e.stopPropagation()}
            style={{
              position: 'fixed',
              left: pos.left,
              top: pos.top,
              width: CARD_WIDTH,
              zIndex: 100,
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 16,
              boxShadow: '0 12px 32px rgba(0,0,0,0.5)',
              padding: 16,
              animation: 'modal-panel-in 0.12s ease',
            }}
          >
            {!profile || isLoading ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <div className="animate-pulse rounded-full" style={{ width: 64, height: 64, background: 'var(--bg-subtle)' }} />
                <div className="animate-pulse rounded" style={{ width: 130, height: 14, background: 'var(--bg-subtle)' }} />
                <div className="animate-pulse rounded" style={{ width: '100%', height: 30, background: 'var(--bg-subtle)' }} />
              </div>
            ) : (
              <>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <Link href={`/profile/${profile.id}`} onClick={(e) => e.stopPropagation()} style={{ textDecoration: 'none' }}>
                    <Avatar username={profile.username} avatarUrl={profile.avatar ?? null} size={64} />
                  </Link>
                  {!isSelf && <FollowButton userId={profile.id} initialFollowing={profile.is_following ?? false} />}
                </div>

                <div style={{ marginTop: 8 }}>
                  <Link
                    href={`/profile/${profile.id}`}
                    onClick={(e) => e.stopPropagation()}
                    style={{ display: 'inline-flex', alignItems: 'center', gap: 4, textDecoration: 'none' }}
                  >
                    <span style={{ fontWeight: 800, fontSize: 16, color: 'var(--text)', letterSpacing: '-0.01em' }}>
                      {profile.name || profile.username}
                    </span>
                    {profile.is_premium && <VerifiedBadge size={16} />}
                  </Link>
                  <div style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>@{profile.username}</div>
                </div>

                {profile.bio && (
                  <p className="line-clamp-3" style={{ marginTop: 8, fontSize: 14, color: 'var(--text)', lineHeight: 1.4 }}>
                    {profile.bio}
                  </p>
                )}

                <div style={{ display: 'flex', gap: 18, marginTop: 10 }}>
                  <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>
                    <strong style={{ color: 'var(--text)' }}>{profile.following_count ?? 0}</strong> Following
                  </span>
                  <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>
                    <strong style={{ color: 'var(--text)' }}>{profile.followers_count ?? 0}</strong> Followers
                  </span>
                </div>
              </>
            )}
          </div>,
          document.body,
        )}
    </span>
  );
}
