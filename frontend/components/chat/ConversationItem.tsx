'use client';

import { useState, useRef, useEffect } from 'react';
import { BellOff, MoreHorizontal, Bell, BellOff as BellOffSm } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import type { ChatRoom } from '@/lib/services/chat';

interface ConversationItemProps {
  room: ChatRoom;
  active: boolean;
  isTyping?: boolean;
  onClick: () => void;
  onMute?: () => void;
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

function getLastMessageText(room: ChatRoom): string | null {
  const lm = room.last_message;
  if (!lm) return null;
  if (lm.message) return lm.message;
  if (lm.gif_url) return 'GIF';
  if (lm.has_media) return 'Media';
  return null;
}

export function ConversationItem({ room, active, isTyping = false, onClick, onMute }: ConversationItemProps) {
  const isPrivate = room.type === 'private';
  const other = room.other_user;
  const [hovered, setHovered] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const hasUnread = (room.unread_count ?? 0) > 0;

  const displayName = isPrivate
    ? (other?.name || other?.username || 'Direct Message')
    : (room.name ?? 'Group Chat');

  const defaultSubtitle = isPrivate
    ? (other?.faculty_name || other?.username || 'Private conversation')
    : `${room.member_count ?? '?'} members`;

  const lastMessageText = getLastMessageText(room);
  const subtitle = isTyping ? 'typing...' : (lastMessageText ?? defaultSubtitle);

  useEffect(() => {
    if (!menuOpen) return;
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [menuOpen]);

  return (
    <div
      style={{ position: 'relative' }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => { setHovered(false); }}
    >
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
            room.avatar_url ? (
              <img
                src={room.avatar_url}
                alt={room.name ?? 'Group'}
                style={{ width: 40, height: 40, borderRadius: '50%', objectFit: 'cover' }}
              />
            ) : (
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
            )
          ) : (
            <Avatar
            username={other?.username ?? 'DM'}
            avatarUrl={other?.avatar_url ?? null}
            size={40}
          />
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
                fontWeight: hasUnread ? 700 : 500,
                fontSize: 14,
                color: 'var(--text)',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
                maxWidth: 120,
              }}
            >
              {displayName}
            </span>
            <div style={{ display: 'flex', alignItems: 'center', gap: 4, flexShrink: 0 }}>
              {room.is_muted && !hovered && (
                <BellOff size={12} style={{ color: 'var(--text-muted)' }} />
              )}
              <span style={{ fontSize: 11, color: hasUnread ? 'var(--accent)' : 'var(--text-muted)' }}>
                {formatTime(room.updated_at)}
              </span>
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 6 }}>
            <div
              style={{
                flex: 1,
                minWidth: 0,
                fontSize: 12,
                color: isTyping ? 'var(--accent)' : hasUnread ? 'var(--text)' : 'var(--text-muted)',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
                fontStyle: isTyping ? 'italic' : 'normal',
                fontWeight: hasUnread && !isTyping ? 500 : 400,
              }}
            >
              {subtitle}
            </div>
            {hasUnread && (
              <span
                style={{
                  minWidth: 18,
                  height: 18,
                  padding: '0 5px',
                  background: 'var(--accent)',
                  borderRadius: 9999,
                  fontSize: 11,
                  fontWeight: 600,
                  color: '#fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                }}
              >
                {room.unread_count}
              </span>
            )}
          </div>
        </div>
      </button>

      {/* "···" button appears on hover */}
      {hovered && onMute && (
        <div
          ref={menuRef}
          style={{ position: 'absolute', top: 10, right: 10, zIndex: 10 }}
        >
          <button
            onClick={(e) => { e.stopPropagation(); setMenuOpen((v) => !v); }}
            style={{
              width: 26,
              height: 26,
              borderRadius: 6,
              background: active ? 'var(--sidebar-active)' : 'var(--sidebar-hover)',
              border: '1px solid var(--border)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <MoreHorizontal size={14} style={{ color: 'var(--text-muted)' }} />
          </button>

          {menuOpen && (
            <div
              style={{
                position: 'absolute',
                top: 30,
                right: 0,
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 10,
                padding: 4,
                minWidth: 140,
                boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
              }}
            >
              <button
                onClick={(e) => { e.stopPropagation(); setMenuOpen(false); onMute(); }}
                style={{
                  width: '100%',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '8px 12px',
                  background: 'none',
                  border: 'none',
                  borderRadius: 7,
                  cursor: 'pointer',
                  color: 'var(--text)',
                  fontSize: 13,
                  textAlign: 'left',
                  fontFamily: 'inherit',
                }}
                onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--sidebar-hover)'; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; }}
              >
                {room.is_muted
                  ? <><Bell size={14} style={{ color: 'var(--text-muted)', flexShrink: 0 }} /> Unmute</>
                  : <><BellOffSm size={14} style={{ color: 'var(--text-muted)', flexShrink: 0 }} /> Mute</>
                }
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
