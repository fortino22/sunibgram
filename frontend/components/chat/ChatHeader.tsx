'use client';

import { Avatar } from '@/components/shared/Avatar';
import type { ChatRoom } from '@/lib/services/chat';
import type { UserProfile } from '@/lib/services/profile';

interface ChatHeaderProps {
  room: ChatRoom;
  otherProfile?: UserProfile | null;
  isTyping: boolean;
  onGroupSettings?: () => void;
}

export function ChatHeader({ room, otherProfile, isTyping, onGroupSettings }: ChatHeaderProps) {
  const isPrivate = room.type === 'private';
  const other = room.other_user;

  const displayName = isPrivate
    ? (otherProfile?.name || other?.name || other?.username || 'Direct Message')
    : (room.name ?? 'Group Chat');

  const subtitle = isPrivate
    ? (otherProfile?.faculty?.faculty_name || other?.faculty_name || other?.username || null)
    : `${room.member_count ?? '?'} members`;

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
      {/* Avatar */}
      {!isPrivate ? (
        room.avatar_url ? (
          <img
            src={room.avatar_url}
            alt={room.name ?? 'Group'}
            style={{ width: 36, height: 36, borderRadius: '50%', objectFit: 'cover', flexShrink: 0 }}
          />
        ) : (
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
        )
      ) : (
        <div style={{ position: 'relative', flexShrink: 0 }}>
          <Avatar
            username={other?.username ?? 'DM'}
            avatarUrl={otherProfile?.avatar}
            size={36}
          />
          {/* Typing dots overlaid on avatar when typing */}
          {isTyping && (
            <div
              style={{
                position: 'absolute',
                bottom: -2,
                right: -2,
                background: 'var(--bg-card)',
                borderRadius: 10,
                padding: '2px 5px',
                border: '1px solid var(--border)',
                display: 'flex',
                gap: 3,
                alignItems: 'center',
              }}
            >
              {[0, 1, 2].map((i) => (
                <span
                  key={i}
                  style={{
                    width: 4,
                    height: 4,
                    borderRadius: '50%',
                    background: 'var(--accent)',
                    display: 'inline-block',
                    animation: `bounce 1.1s ease-in-out ${i * 0.18}s infinite`,
                  }}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Name + subtitle */}
      <div>
        <div
          onClick={!isPrivate ? onGroupSettings : undefined}
          style={{
            fontWeight: 600,
            fontSize: 15,
            color: 'var(--text)',
            letterSpacing: '-0.02em',
            cursor: !isPrivate ? 'pointer' : 'default',
            display: 'inline-block',
          }}
          title={!isPrivate ? 'Edit group settings' : undefined}
        >
          {displayName}
        </div>
        {subtitle && (
          <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 1 }}>
            {subtitle}
          </div>
        )}
      </div>
    </div>
  );
}
