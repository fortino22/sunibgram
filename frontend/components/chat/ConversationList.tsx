'use client';

import { useState } from 'react';
import { Plus } from 'lucide-react';
import { ConversationItem } from './ConversationItem';
import { NewChatModal } from './NewChatModal';
import { useChatRooms } from '@/hooks/chat/useChatRooms';

interface ConversationListProps {
  activeRoomId: number | null;
  onRoomSelect: (roomId: number) => void;
}

export function ConversationList({ activeRoomId, onRoomSelect }: ConversationListProps) {
  const { data, isLoading } = useChatRooms();
  const [showNewChat, setShowNewChat] = useState(false);
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
          onClick={() => setShowNewChat(true)}
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
              textAlign: 'center',
              padding: '40px 20px',
              color: 'var(--text-muted)',
              fontSize: 13,
            }}
          >
            <div style={{ marginBottom: 6 }}>No conversations yet</div>
            <div>Start one with the + button above</div>
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

      {showNewChat && (
        <NewChatModal
          onClose={() => setShowNewChat(false)}
          onRoomCreated={(roomId) => {
            setShowNewChat(false);
            onRoomSelect(roomId);
          }}
        />
      )}
    </div>
  );
}
