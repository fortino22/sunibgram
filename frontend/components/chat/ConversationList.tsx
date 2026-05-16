'use client';

import { Plus, MessageCircle } from 'lucide-react';
import { ConversationItem } from './ConversationItem';
import { useChatRooms } from '@/hooks/chat/useChatRooms';

interface ConversationListProps {
  activeRoomId: number | null;
  onRoomSelect: (roomId: number) => void;
  onNewChat: () => void;
}

export function ConversationList({ activeRoomId, onRoomSelect, onNewChat }: ConversationListProps) {
  const { data, isLoading } = useChatRooms();
  const rooms = data?.data ?? [];

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

      {/* Room list */}
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '8px',
        }}
      >
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
        ) : rooms.length === 0 ? (
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
            <MessageCircle size={52} color="var(--text-muted)" strokeWidth={1.2} />
            <div style={{ fontWeight: 600, fontSize: 16, color: 'var(--text)' }}>
              Empty inbox
            </div>
            <div style={{ fontSize: 13, color: 'var(--text-muted)', textAlign: 'center' }}>
              Message someone
            </div>
          </div>
        ) : (
          rooms.map((room) => (
            <ConversationItem
              key={room.room_id}
              room={room}
              active={room.room_id === activeRoomId}
              onClick={() => onRoomSelect(room.room_id)}
            />
          ))
        )}
      </div>

    </div>
  );
}
