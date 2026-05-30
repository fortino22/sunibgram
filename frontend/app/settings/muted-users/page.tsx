'use client';

import { useState } from 'react';
import { Avatar } from '@/components/shared/Avatar';
import { EmptyState } from '@/components/shared/EmptyState';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { SettingsPanel } from '@/components/settings/SettingsPanel';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { useMutedUsers, useUnmute } from '@/hooks/useMutedUsers';
import type { MutedUser } from '@/lib/services/mutes';

export default function MutedUsersSettingsPage() {
  const { data: muted, isLoading } = useMutedUsers();
  const unmute = useUnmute();
  const [target, setTarget] = useState<MutedUser | null>(null);

  function confirmUnmute() {
    if (!target) return;
    const id = target.id;
    unmute.mutate(id, {
      onSettled: () => {
        // Close once the mutation resolves (success removes the row optimistically).
        setTarget(null);
      },
    });
  }

  return (
    <SettingsPanel
      title="Muted accounts"
      description="You won't see posts or notifications from people you've muted. They aren't told."
    >
      <div style={{ maxWidth: 640 }}>
        {isLoading ? (
          <SkeletonList />
        ) : !muted || muted.length === 0 ? (
          <EmptyState
            title="No muted accounts"
            subtitle="When you mute someone, they'll show up here so you can manage them."
          />
        ) : (
          <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
            {muted.map((user) => (
              <MutedUserRow
                key={user.id}
                user={user}
                onUnmute={() => setTarget(user)}
                disabled={unmute.isPending}
              />
            ))}
          </ul>
        )}
      </div>

      <UnmuteConfirmModal
        user={target}
        busy={unmute.isPending}
        onConfirm={confirmUnmute}
        onClose={() => { if (!unmute.isPending) setTarget(null); }}
      />
    </SettingsPanel>
  );
}

function UnmuteConfirmModal({
  user,
  busy,
  onConfirm,
  onClose,
}: {
  user: MutedUser | null;
  busy: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog open={user !== null} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="!max-w-[340px] !rounded-2xl">
        <div style={{ padding: '22px 22px 18px' }}>
          <h2
            style={{
              fontSize: 18,
              fontWeight: 700,
              color: 'var(--text)',
              margin: 0,
              letterSpacing: '-0.01em',
            }}
          >
            Unmute {user ? `@${user.username}` : ''}?
          </h2>
          <p
            style={{
              fontSize: 14,
              color: 'var(--text-secondary)',
              margin: '8px 0 0',
              lineHeight: 1.5,
            }}
          >
            Their posts and notifications will show up in your feeds again.
          </p>
        </div>
        <div
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            gap: 8,
            padding: '14px 24px',
            borderTop: '1px solid var(--border)',
          }}
        >
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            style={{
              padding: '9px 18px',
              borderRadius: 9999,
              background: 'transparent',
              border: '1px solid var(--border-strong)',
              color: 'var(--text)',
              fontWeight: 600,
              fontSize: 14,
              fontFamily: 'inherit',
              cursor: busy ? 'not-allowed' : 'pointer',
              opacity: busy ? 0.6 : 1,
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={busy}
            style={{
              padding: '9px 20px',
              borderRadius: 9999,
              background: 'var(--text)',
              border: 'none',
              color: 'var(--bg)',
              fontWeight: 700,
              fontSize: 14,
              fontFamily: 'inherit',
              cursor: busy ? 'not-allowed' : 'pointer',
              opacity: busy ? 0.6 : 1,
              minWidth: 96,
            }}
          >
            {busy ? 'Working…' : 'Unmute'}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function MutedUserRow({
  user,
  onUnmute,
  disabled,
}: {
  user: MutedUser;
  onUnmute: () => void;
  disabled: boolean;
}) {
  return (
    <li
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '12px 0',
        borderBottom: '1px solid var(--border)',
      }}
    >
      <Avatar username={user.username} avatarUrl={user.avatar} size="md" />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 4,
            fontSize: 15,
            fontWeight: 600,
            color: 'var(--text)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {user.name}
          {user.is_premium && <VerifiedBadge size={15} />}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{user.username}</div>
      </div>
      <button
        type="button"
        onClick={onUnmute}
        disabled={disabled}
        style={{
          background: 'transparent',
          color: 'var(--text)',
          border: '1px solid var(--border-strong)',
          borderRadius: 9999,
          padding: '7px 16px',
          fontWeight: 700,
          fontSize: 14,
          fontFamily: 'inherit',
          cursor: disabled ? 'not-allowed' : 'pointer',
          opacity: disabled ? 0.6 : 1,
          flexShrink: 0,
        }}
      >
        Unmute
      </button>
    </li>
  );
}

function SkeletonList() {
  return (
    <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
      {[0, 1, 2].map((i) => (
        <li
          key={i}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '12px 0',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <div style={{ width: 40, height: 40, borderRadius: '50%', background: 'var(--bg-subtle)', flexShrink: 0 }} />
          <div style={{ flex: 1 }}>
            <div style={{ width: 140, height: 14, background: 'var(--bg-subtle)', borderRadius: 4, marginBottom: 6 }} />
            <div style={{ width: 90, height: 12, background: 'var(--bg-subtle)', borderRadius: 4 }} />
          </div>
          <div style={{ width: 84, height: 32, background: 'var(--bg-subtle)', borderRadius: 9999 }} />
        </li>
      ))}
    </ul>
  );
}
