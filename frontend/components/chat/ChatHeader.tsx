'use client';

import { Avatar } from '@/components/shared/Avatar';
import type { ChatRoom } from '@/lib/services/chat';

interface ChatHeaderProps {
  room: ChatRoom;
  otherUsername?: string;
  typingUsers: string[];
}

export function ChatHeader({ room, otherUsername, typingUsers }: ChatHeaderProps) {
  const displayName =
    room.type === 'group'
      ? (room.name ?? 'Group Chat')
      : (otherUsername ?? 'Direct Message');

  const subtitle =
    typingUsers.length > 0
      ? `${typingUsers.join(', ')} ${typingUsers.length === 1 ? 'is' : 'are'} typing...`
      : room.type === 'group'
      ? `${room.member_count ?? '?'} members`
      : null;

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
        flexShrink: 0,
        background: 'var(--bg-card)',
      }}
    >
      {room.type === 'group' ? (
        <div
          style={{
            width: 36,
            height: 36,
            borderRadius: '50%',
            background: 'var(--accent)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 15,
            color: '#fff',
            fontWeight: 600,
            flexShrink: 0,
          }}
        >
          {(room.name ?? 'G').charAt(0).toUpperCase()}
        </div>
      ) : (
        <Avatar username={otherUsername ?? 'DM'} size={36} />
      )}

      <div>
        <div
          style={{
            fontWeight: 600,
            fontSize: 15,
            color: 'var(--text)',
            letterSpacing: '-0.02em',
          }}
        >
          {displayName}
        </div>
        {subtitle && (
          <div
            style={{
              fontSize: 12,
              color: typingUsers.length > 0 ? 'var(--accent)' : 'var(--text-muted)',
              marginTop: 1,
            }}
          >
            {subtitle}
          </div>
        )}
      </div>
    </div>
  );
}
