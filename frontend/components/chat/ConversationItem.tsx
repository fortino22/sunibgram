'use client';

import { Avatar } from '@/components/shared/Avatar';
import type { ChatRoom } from '@/lib/services/chat';

interface ConversationItemProps {
  room: ChatRoom;
  active: boolean;
  onClick: () => void;
}

function formatTime(iso: string): string {
  const date = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / 86_400_000);

  if (diffDays === 0) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return date.toLocaleDateString([], { weekday: 'short' });
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
}

export function ConversationItem({ room, active, onClick }: ConversationItemProps) {
  const isPrivate = room.type === 'private';
  const other = room.other_user;

  const displayName = isPrivate
    ? (other?.name || other?.username || 'Direct Message')
    : (room.name ?? 'Group Chat');

  const subtitle = isPrivate
    ? (other?.faculty_name || other?.username || 'Private conversation')
    : `${room.member_count ?? '?'} members`;

  return (
    <button
      onClick={onClick}
      style={{
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '12px 16px',
        background: active ? 'var(--sidebar-active)' : 'transparent',
        border: 'none',
        borderRadius: 10,
        cursor: 'pointer',
        textAlign: 'left',
        transition: 'background 0.1s ease',
        marginBottom: 2,
      }}
      onMouseEnter={(e) => {
        if (!active) e.currentTarget.style.background = 'var(--sidebar-hover)';
      }}
      onMouseLeave={(e) => {
        if (!active) e.currentTarget.style.background = 'transparent';
      }}
    >
      <div style={{ flexShrink: 0 }}>
        {!isPrivate ? (
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: '50%',
              background: 'var(--accent)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 16,
              color: '#fff',
              fontWeight: 600,
            }}
          >
            {(room.name ?? 'G').charAt(0).toUpperCase()}
          </div>
        ) : (
          <Avatar username={other?.username ?? 'DM'} size={40} />
        )}
      </div>

      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 2,
          }}
        >
          <span
            style={{
              fontWeight: 500,
              fontSize: 14,
              color: 'var(--text)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              maxWidth: 140,
            }}
          >
            {displayName}
          </span>
          <span style={{ fontSize: 11, color: 'var(--text-muted)', flexShrink: 0 }}>
            {formatTime(room.updated_at)}
          </span>
        </div>
        <div
          style={{
            fontSize: 12,
            color: 'var(--text-muted)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {subtitle}
        </div>
      </div>
    </button>
  );
}
