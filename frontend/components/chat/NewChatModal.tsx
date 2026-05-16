'use client';

import { useState, useEffect } from 'react';
import { X, Search, MessageSquare } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useExplorePeople, useSuggestedPeople } from '@/hooks/useExplore';
import { useCreatePrivateRoom } from '@/hooks/chat/useCreateChatRoom';
import type { ExplorePerson } from '@/lib/services/explore';

interface NewChatModalProps {
  onClose: () => void;
  onRoomCreated: (roomId: number) => void;
}

export function NewChatModal({ onClose, onRoomCreated }: NewChatModalProps) {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const createRoom = useCreatePrivateRoom();

  // 300ms debounce to avoid hammering the API on every keystroke
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(t);
  }, [search]);

  const hasQuery = debouncedSearch.length > 0;
  const searchQuery = useExplorePeople(debouncedSearch);
  const suggestedQuery = useSuggestedPeople(10);

  const isLoading = hasQuery ? searchQuery.isLoading : suggestedQuery.isLoading;
  const people: ExplorePerson[] = hasQuery
    ? (searchQuery.data ?? [])
    : (suggestedQuery.data ?? []);

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
              placeholder="Search anyone by name or username…"
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

        {/* Section label */}
        <div
          style={{
            padding: '8px 20px 4px',
            fontSize: 11,
            fontWeight: 600,
            color: 'var(--text-muted)',
            textTransform: 'uppercase',
            letterSpacing: '0.06em',
            flexShrink: 0,
          }}
        >
          {hasQuery ? 'Results' : 'Suggested'}
        </div>

        {/* User list */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '4px 8px 8px' }}>
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
          ) : people.length === 0 ? (
            <div
              style={{
                textAlign: 'center',
                padding: '32px 20px',
                color: 'var(--text-muted)',
                fontSize: 13,
              }}
            >
              {hasQuery ? 'No users found' : 'No suggestions available'}
            </div>
          ) : (
            people.map((u) => (
              <button
                key={u.userId}
                onClick={() => handleSelect(u.userId)}
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
                  cursor: createRoom.isPending ? 'not-allowed' : 'pointer',
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
                  avatarUrl={u.userProfilePic ?? null}
                  size={36}
                />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text)', display: 'flex', alignItems: 'center', gap: 4 }}>
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {u.name || u.username}
                    </span>
                    {u.isPremium && (
                      <span style={{ fontSize: 10, color: 'var(--accent)', fontWeight: 700, flexShrink: 0 }}>★</span>
                    )}
                  </div>
                  <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>@{u.username}</div>
                </div>
                {u.isFollowing && (
                  <span style={{ fontSize: 11, color: 'var(--text-muted)', flexShrink: 0 }}>Following</span>
                )}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
