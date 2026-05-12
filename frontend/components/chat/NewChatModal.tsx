'use client';

import { useState } from 'react';
import { X, Search, MessageSquare } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useFollowing } from '@/hooks/useFollowing';
import { useCreatePrivateRoom } from '@/hooks/chat/useCreateChatRoom';
import { useAuthStore } from '@/stores/authStore';
import type { FollowUser } from '@/lib/services/follows';

interface NewChatModalProps {
  onClose: () => void;
  onRoomCreated: (roomId: number) => void;
}

export function NewChatModal({ onClose, onRoomCreated }: NewChatModalProps) {
  const user = useAuthStore((s) => s.user);
  const [search, setSearch] = useState('');
  const { data: followingData, isLoading } = useFollowing(user?.id ?? '');
  const createRoom = useCreatePrivateRoom();

  const following: FollowUser[] = followingData?.data ?? [];
  const filtered = following.filter(
    (u) =>
      u.username.toLowerCase().includes(search.toLowerCase()) ||
      (u.name ?? '').toLowerCase().includes(search.toLowerCase())
  );

  const handleSelect = (targetUserId: number) => {
    createRoom.mutate(targetUserId, {
      onSuccess: (data) => onRoomCreated(data.room_id),
    });
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          width: 420,
          maxWidth: 'calc(100vw - 32px)',
          maxHeight: '80vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '16px 20px',
            borderBottom: '1px solid var(--border)',
            flexShrink: 0,
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              fontWeight: 600,
              fontSize: 16,
              color: 'var(--text)',
            }}
          >
            <MessageSquare size={18} color="var(--accent)" />
            New Message
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: 4,
              borderRadius: 6,
              display: 'flex',
            }}
          >
            <X size={18} color="var(--text-muted)" />
          </button>
        </div>

        {/* Search */}
        <div
          style={{
            padding: '12px 20px',
            borderBottom: '1px solid var(--border)',
            flexShrink: 0,
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              background: 'var(--bg)',
              border: '1px solid var(--border)',
              borderRadius: 10,
              padding: '8px 12px',
            }}
          >
            <Search size={15} color="var(--text-muted)" />
            <input
              type="text"
              placeholder="Search people you follow..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              autoFocus
              style={{
                flex: 1,
                background: 'none',
                border: 'none',
                outline: 'none',
                fontSize: 14,
                color: 'var(--text)',
                fontFamily: 'inherit',
              }}
            />
          </div>
        </div>

        {/* User list */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '8px' }}>
          {isLoading ? (
            Array.from({ length: 4 }, (_, i) => (
              <div
                key={i}
                style={{
                  height: 56,
                  borderRadius: 10,
                  background: 'var(--skeleton, oklch(16% 0.005 260))',
                  marginBottom: 4,
                  animation: 'pulse 1.5s ease-in-out infinite',
                }}
              />
            ))
          ) : filtered.length === 0 ? (
            <div
              style={{
                textAlign: 'center',
                padding: '32px 20px',
                color: 'var(--text-muted)',
                fontSize: 13,
              }}
            >
              {following.length === 0
                ? 'Follow someone to start a conversation'
                : 'No matches found'}
            </div>
          ) : (
            filtered.map((u) => (
              <button
                key={u.user_id}
                onClick={() => handleSelect(u.user_id)}
                disabled={createRoom.isPending}
                style={{
                  width: '100%',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '10px 12px',
                  background: 'none',
                  border: 'none',
                  borderRadius: 10,
                  cursor: 'pointer',
                  textAlign: 'left',
                  transition: 'background 0.1s',
                  opacity: createRoom.isPending ? 0.6 : 1,
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--sidebar-hover)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'none';
                }}
              >
                <Avatar
                  username={u.username}
                  avatarUrl={u.avatar_url ?? null}
                  size={36}
                />
                <div>
                  <div
                    style={{ fontWeight: 500, fontSize: 14, color: 'var(--text)' }}
                  >
                    {u.name || u.username}
                  </div>
                  <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                    @{u.username}
                  </div>
                </div>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
