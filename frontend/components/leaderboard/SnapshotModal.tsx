'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { X } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useMyTrendingSnapshot } from '@/hooks/useMyTrendingSnapshot';

export interface SnapshotOwnerHint {
  username: string;
  name?: string;
  profile?: string | null;
  isPremium?: boolean;
}

interface SnapshotModalProps {
  snapshotId: number;
  ownerHint?: SnapshotOwnerHint;
  onClose: () => void;
}

function formatWeekRange(weekStart: string): string {
  const start = new Date(weekStart);
  if (Number.isNaN(start.getTime())) return weekStart;
  const end = new Date(start);
  end.setUTCDate(start.getUTCDate() + 6);
  const fmt = (d: Date) =>
    d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  return `${fmt(start)} → ${fmt(end)}`;
}

function HoverCell({ children, padding = '10px 12px', radius = 10 }: { children: ReactNode; padding?: string; radius?: number }) {
  const [hover, setHover] = useState(false);
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        background: hover ? 'var(--bg-hover)' : 'var(--bg-subtle)',
        border: '1px solid',
        borderColor: hover ? 'var(--text-tertiary)' : 'var(--border)',
        borderRadius: radius,
        padding,
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
        transition: 'background 0.15s ease, border-color 0.15s ease, transform 0.15s ease',
        transform: hover ? 'translateY(-1px)' : 'translateY(0)',
        cursor: 'default',
      }}
    >
      {children}
    </div>
  );
}

function StatCell({ label, value }: { label: string; value: number }) {
  return (
    <HoverCell>
      <span
        style={{
          fontSize: 11,
          color: 'var(--text-tertiary)',
          textTransform: 'uppercase',
          letterSpacing: '0.04em',
        }}
      >
        {label}
      </span>
      <span style={{ fontSize: 18, fontWeight: 700, color: 'var(--text)' }}>{value}</span>
    </HoverCell>
  );
}

export function SnapshotModal({ snapshotId, ownerHint, onClose }: SnapshotModalProps) {
  const { data, isLoading, isError } = useMyTrendingSnapshot(snapshotId);

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose]);

  return (
    <div
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 70,
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
          maxWidth: 460,
          maxHeight: '85vh',
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

        {isLoading && (
          <div style={{ padding: '24px 0', color: 'var(--text-tertiary)', fontSize: 14 }}>
            Loading snapshot…
          </div>
        )}
        {isError && !isLoading && (
          <div style={{ padding: '24px 0', color: 'var(--coral)', fontSize: 14 }}>
            Failed to load snapshot.
          </div>
        )}
        {data && !isLoading && (
          <>
            {ownerHint && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  marginBottom: 14,
                  padding: '8px 10px',
                  background: 'var(--bg-subtle)',
                  border: '1px solid var(--border)',
                  borderRadius: 12,
                }}
              >
                <Avatar
                  username={ownerHint.username}
                  avatarUrl={ownerHint.profile ?? null}
                  size={32}
                />
                <div style={{ minWidth: 0 }}>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 4,
                      fontSize: 13,
                      fontWeight: 700,
                      color: 'var(--text)',
                    }}
                  >
                    {ownerHint.name || ownerHint.username}
                    {ownerHint.isPremium && <VerifiedBadge size={12} />}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                    @{ownerHint.username}
                  </div>
                </div>
              </div>
            )}

            <div style={{ marginBottom: 16 }}>
              <div
                style={{
                  fontSize: 11,
                  color: 'var(--text-tertiary)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.06em',
                  marginBottom: 4,
                }}
              >
                Week
              </div>
              <div style={{ fontSize: 18, fontWeight: 800, color: 'var(--text)' }}>
                {formatWeekRange(data.weekStart)}
              </div>
            </div>

            <div
              style={{
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: 10,
                marginBottom: 18,
              }}
            >
              <HoverCell padding="12px 14px" radius={12}>
                <div
                  style={{
                    fontSize: 11,
                    color: 'var(--text-tertiary)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.04em',
                  }}
                >
                  Rank
                </div>
                <div
                  style={{
                    fontSize: 22,
                    fontWeight: 800,
                    color: 'var(--text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  #{data.rank}
                </div>
              </HoverCell>
              <HoverCell padding="12px 14px" radius={12}>
                <div
                  style={{
                    fontSize: 11,
                    color: 'var(--text-tertiary)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.04em',
                  }}
                >
                  Score
                </div>
                <div
                  style={{
                    fontSize: 22,
                    fontWeight: 800,
                    color: 'var(--text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {data.score.toFixed(1)}
                </div>
              </HoverCell>
            </div>

            <div style={{ marginBottom: 16 }}>
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
                <StatCell label="Likes" value={data.received.likes} />
                <StatCell label="Comments" value={data.received.comments} />
                <StatCell label="Bookmarks" value={data.received.bookmarks} />
                <StatCell label="Reposts" value={data.received.reposts} />
                <StatCell label="Threads" value={data.received.threadCount} />
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
                <StatCell label="Likes" value={data.given.likes} />
                <StatCell label="Comments" value={data.given.comments} />
                <StatCell label="Bookmarks" value={data.given.bookmarks} />
                <StatCell label="Posts" value={data.given.posts} />
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
