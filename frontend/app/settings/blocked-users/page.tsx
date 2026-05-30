'use client';

import { useState } from 'react';
import { Avatar } from '@/components/shared/Avatar';
import { EmptyState } from '@/components/shared/EmptyState';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { SettingsPanel } from '@/components/settings/SettingsPanel';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { useBlockedUsers, useUnblock } from '@/hooks/useBlockedUsers';
import type { BlockedUser } from '@/lib/services/blocks';

export default function BlockedUsersSettingsPage() {
  const { data: blocked, isLoading } = useBlockedUsers();
  const unblock = useUnblock();
  const [target, setTarget] = useState<BlockedUser | null>(null);

  function confirmUnblock() {
    if (!target) return;
    const id = target.id;
    unblock.mutate(id, {
      onSettled: () => {
        // Close once the mutation resolves (success removes the row optimistically).
        setTarget(null);
      },
    });
  }

  return (
    <SettingsPanel
      title="Blocked accounts"
      description="People you've blocked can't see your posts, follow you, or message you."
    >
      <div style={{ maxWidth: 640 }}>
        {isLoading ? (
          <SkeletonList />
        ) : !blocked || blocked.length === 0 ? (
          <EmptyState
            title="No blocked accounts"
            subtitle="When you block someone, they'll show up here so you can manage them."
          />
        ) : (
          <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
            {blocked.map((user) => (
              <BlockedUserRow
                key={user.id}
                user={user}
                onUnblock={() => setTarget(user)}
                disabled={unblock.isPending}
              />
            ))}
          </ul>
        )}
      </div>

      <UnblockConfirmModal
        user={target}
        busy={unblock.isPending}
        onConfirm={confirmUnblock}
        onClose={() => { if (!unblock.isPending) setTarget(null); }}
      />
    </SettingsPanel>
  );
}

function UnblockConfirmModal({
  user,
  busy,
  onConfirm,
  onClose,
}: {
  user: BlockedUser | null;
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
            Unblock {user ? `@${user.username}` : ''}?
          </h2>
          <p
            style={{
              fontSize: 14,
              color: 'var(--text-secondary)',
              margin: '8px 0 0',
              lineHeight: 1.5,
            }}
          >
            They will be able to see your posts, follow you, and message you again.
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
            {busy ? 'Working…' : 'Unblock'}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function BlockedUserRow({
  user,
  onUnblock,
  disabled,
}: {
  user: BlockedUser;
  onUnblock: () => void;
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
        onClick={onUnblock}
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
        Unblock
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
