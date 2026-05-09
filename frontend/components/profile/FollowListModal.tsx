'use client';

import { useRouter } from 'next/navigation';
import { X } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useFollowers } from '@/hooks/useFollowers';
import { useFollowing } from '@/hooks/useFollowing';
import type { FollowUser } from '@/lib/services/follows';

interface FollowListModalProps {
  userId: number | string;
  type: 'followers' | 'following';
  onClose: () => void;
}

function UserRow({ user, onNavigate }: { user: FollowUser; onNavigate: (id: number) => void }) {
  return (
    <button
      type="button"
      onClick={() => onNavigate(user.user_id)}
      style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '12px 20px', width: '100%',
        background: 'none', border: 'none', borderBottom: '1px solid var(--border)',
        cursor: 'pointer', textAlign: 'left',
      }}
    >
      <Avatar username={user.name ?? user.username} avatarUrl={user.avatar_url} size="md" />
      <div>
        <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>
          {user.name ?? user.username}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{user.username}</div>
      </div>
    </button>
  );
}

export function FollowListModal({ userId, type, onClose }: FollowListModalProps) {
  const router = useRouter();
  const followersQuery = useFollowers(userId);
  const followingQuery = useFollowing(userId);

  const query = type === 'followers' ? followersQuery : followingQuery;
  const users = query.data?.data ?? [];
  const title = type === 'followers' ? 'Followers' : 'Following';

  function navigate(id: number) {
    router.push(`/profile/${id}`);
    onClose();
  }

  return (
    <div
      style={{ position: 'fixed', inset: 0, zIndex: 50, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, width: '100%', maxWidth: 400, maxHeight: '70vh', display: 'flex', flexDirection: 'column' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 20px', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
          <span style={{ fontWeight: 700, fontSize: 17, color: 'var(--text)' }}>{title}</span>
          <button type="button" onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text)', padding: 4, borderRadius: '50%', display: 'flex' }}>
            <X size={20} strokeWidth={2} />
          </button>
        </div>

        <div style={{ overflowY: 'auto', flex: 1 }}>
          {query.isLoading && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          )}
          {!query.isLoading && users.length === 0 && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              No {title.toLowerCase()} yet
            </div>
          )}
          {users.map((user) => (
            <UserRow key={user.user_id} user={user} onNavigate={navigate} />
          ))}
        </div>
      </div>
    </div>
  );
}
