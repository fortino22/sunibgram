'use client';

import { useState, useEffect } from 'react';
import { Plus, MessageCircle, Search, X } from 'lucide-react';
import { ConversationItem } from './ConversationItem';
import { Avatar } from '@/components/shared/Avatar';
import { useChatRooms } from '@/hooks/chat/useChatRooms';
import { useExplorePeople } from '@/hooks/useExplore';
import { useCreatePrivateRoom } from '@/hooks/chat/useCreateChatRoom';
import { useMuteRoom } from '@/hooks/chat/useMuteRoom';
import { useAuthStore } from '@/stores/authStore';
import type { ExplorePerson } from '@/lib/services/explore';

interface ConversationListProps {
  activeRoomId: number | null;
  onRoomSelect: (roomId: number) => void;
  onNewChat: () => void;
  typingUsersByRoom?: Record<number, number[]>;
}

export function ConversationList({ activeRoomId, onRoomSelect, onNewChat, typingUsersByRoom = {} }: ConversationListProps) {
  const { data, isLoading } = useChatRooms();
  const rooms = data?.data ?? [];
  const currentUser = useAuthStore((s) => s.user);

  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [creatingFor, setCreatingFor] = useState<number | null>(null);

  const createPrivate = useCreatePrivateRoom();
  const muteRoom = useMuteRoom();

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(t);
  }, [search]);

  const hasSearch = debouncedSearch.length > 0;
  const exploreQuery = useExplorePeople(debouncedSearch);

  // user_ids that already have a private room with the current user
  const existingPrivateUserIds = new Set(
    rooms
      .filter((r) => r.type === 'private' && r.other_user)
      .map((r) => r.other_user!.user_id)
  );

  const sortedRooms = [...rooms].sort((a, b) => {
    const aUnread = (a.unread_count ?? 0) > 0;
    const bUnread = (b.unread_count ?? 0) > 0;
    if (aUnread !== bUnread) return aUnread ? -1 : 1;
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  });

  const filteredRooms = hasSearch
    ? sortedRooms.filter((r) => {
        const q = debouncedSearch.toLowerCase();
        if (r.type === 'private') {
          return (
            r.other_user?.name?.toLowerCase().includes(q) ||
            r.other_user?.username?.toLowerCase().includes(q)
          );
        }
        return r.name?.toLowerCase().includes(q);
      })
    : sortedRooms;

  // People with no existing room and not self
  const newPeople: ExplorePerson[] = hasSearch
    ? (exploreQuery.data ?? []).filter(
        (p) => !existingPrivateUserIds.has(p.userId) && p.userId !== currentUser?.id
      )
    : [];

  const handleStartChat = (person: ExplorePerson) => {
    if (createPrivate.isPending) return;
    setCreatingFor(person.userId);
    createPrivate.mutate(person.userId, {
      onSuccess: (res) => {
        setSearch('');
        onRoomSelect(res.room_id);
      },
      onSettled: () => setCreatingFor(null),
    });
  };

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        borderRight: '1px solid var(--border)',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '16px 16px 12px',
          borderBottom: '1px solid var(--border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}
      >
        <span
          style={{
            fontWeight: 600,
            fontSize: 16,
            color: 'var(--text)',
            letterSpacing: '-0.02em',
          }}
        >
          Messages
        </span>
        <button
          onClick={onNewChat}
          title="New message"
          style={{
            width: 30,
            height: 30,
            borderRadius: 8,
            background: 'var(--accent)',
            border: 'none',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Plus size={16} color="#fff" />
        </button>
      </div>

      {/* Search bar */}
      <div
        style={{
          padding: '10px 12px',
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
            padding: '7px 12px',
          }}
        >
          <Search size={14} style={{ color: 'var(--text-muted)', flexShrink: 0 }} />
          <input
            type="text"
            placeholder="Search people or conversations…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{
              flex: 1,
              background: 'none',
              border: 'none',
              outline: 'none',
              fontSize: 13,
              color: 'var(--text)',
              fontFamily: 'inherit',
            }}
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: 0,
                display: 'flex',
                color: 'var(--text-muted)',
                flexShrink: 0,
              }}
            >
              <X size={13} />
            </button>
          )}
        </div>
      </div>

      {/* List */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '8px' }}>
        {isLoading ? (
          Array.from({ length: 5 }, (_, i) => (
            <div
              key={i}
              style={{
                height: 64,
                borderRadius: 10,
                background: 'var(--skeleton, oklch(16% 0.005 260))',
                marginBottom: 4,
                animation: 'pulse 1.5s ease-in-out infinite',
              }}
            />
          ))
        ) : hasSearch ? (
          <>
            {/* Matching existing conversations */}
            {filteredRooms.length > 0 && (
              <>
                <div
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: 'var(--text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.06em',
                    padding: '4px 8px 6px',
                  }}
                >
                  Conversations
                </div>
                {filteredRooms.map((room) => (
                  <ConversationItem
                    key={room.room_id}
                    room={room}
                    active={room.room_id === activeRoomId}
                    isTyping={(typingUsersByRoom[room.room_id]?.length ?? 0) > 0}
                    onClick={() => {
                      setSearch('');
                      onRoomSelect(room.room_id);
                    }}
                    onMute={() => muteRoom.mutate({ roomId: room.room_id })}
                  />
                ))}
              </>
            )}

            {/* People not yet in a conversation */}
            {exploreQuery.isLoading && !exploreQuery.data ? (
              Array.from({ length: 3 }, (_, i) => (
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
            ) : newPeople.length > 0 ? (
              <>
                <div
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: 'var(--text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.06em',
                    padding: '8px 8px 6px',
                  }}
                >
                  People
                </div>
                {newPeople.map((p) => {
                  const isCreating = creatingFor === p.userId;
                  return (
                    <button
                      key={p.userId}
                      onClick={() => handleStartChat(p)}
                      disabled={createPrivate.isPending}
                      style={{
                        width: '100%',
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        padding: '10px 12px',
                        background: 'none',
                        border: 'none',
                        borderRadius: 10,
                        cursor: createPrivate.isPending ? 'not-allowed' : 'pointer',
                        textAlign: 'left',
                        marginBottom: 2,
                        opacity: createPrivate.isPending && !isCreating ? 0.5 : 1,
                        transition: 'background 0.1s ease',
                      }}
                      onMouseEnter={(e) => {
                        if (!createPrivate.isPending)
                          e.currentTarget.style.background = 'var(--sidebar-hover)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'none';
                      }}
                    >
                      <Avatar
                        username={p.username}
                        avatarUrl={p.userProfilePic || null}
                        size={40}
                      />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div
                          style={{
                            fontWeight: 500,
                            fontSize: 14,
                            color: 'var(--text)',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          }}
                        >
                          {isCreating ? 'Opening…' : (p.name || p.username)}
                        </div>
                        <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                          @{p.username}
                        </div>
                      </div>
                    </button>
                  );
                })}
              </>
            ) : filteredRooms.length === 0 ? (
              <div
                style={{
                  textAlign: 'center',
                  padding: '32px 20px',
                  color: 'var(--text-muted)',
                  fontSize: 13,
                }}
              >
                No results for &ldquo;{debouncedSearch}&rdquo;
              </div>
            ) : null}
          </>
        ) : sortedRooms.length === 0 ? (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '60px 20px',
              gap: 10,
            }}
          >
            <MessageCircle size={52} strokeWidth={1.2} style={{ color: 'var(--text-muted)' }} />
            <div style={{ fontWeight: 600, fontSize: 16, color: 'var(--text)' }}>Empty inbox</div>
            <div style={{ fontSize: 13, color: 'var(--text-muted)', textAlign: 'center' }}>
              Message someone
            </div>
          </div>
        ) : (
          sortedRooms.map((room) => (
            <ConversationItem
              key={room.room_id}
              room={room}
              active={room.room_id === activeRoomId}
              isTyping={(typingUsersByRoom[room.room_id]?.length ?? 0) > 0}
              onClick={() => onRoomSelect(room.room_id)}
              onMute={() => muteRoom.mutate({ roomId: room.room_id })}
            />
          ))
        )}
      </div>
    </div>
  );
}
