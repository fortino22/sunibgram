'use client';

import Link from 'next/link';
import { Trash2 } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import type { NotificationItem } from '@/lib/services/notifications';

interface NotificationCardProps {
  notification: NotificationItem;
  onDelete: (id: number) => void;
  isDeleting: boolean;
}

function resolveHref(notification: NotificationItem): string {
  switch (notification.type) {
    case 'follow':
      return `/profile/${notification.actor.user_id}`;
    case 'chat_message':
      return notification.reference_id
        ? `/messages?room=${notification.reference_id}`
        : '/messages';
    default:
      return notification.reference_id
        ? `/thread/${notification.reference_id}`
        : '#';
  }
}

function formatRelativeTime(isoString: string): string {
  const diffMs = Date.now() - new Date(isoString).getTime();
  const diffSec = Math.floor(diffMs / 1000);
  if (diffSec < 60) return `${diffSec}s`;
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin}m`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h`;
  const diffDay = Math.floor(diffHr / 24);
  if (diffDay < 7) return `${diffDay}d`;
  return new Date(isoString).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

export function NotificationCard({ notification, onDelete, isDeleting }: NotificationCardProps) {
  const href = resolveHref(notification);

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
        transition: 'background 0.12s',
        position: 'relative',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLDivElement).style.background = 'var(--bg-subtle)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLDivElement).style.background = 'transparent'; }}
    >
      <Link
        href={href}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          flex: 1,
          minWidth: 0,
          textDecoration: 'none',
        }}
      >
        <Avatar
          username={notification.actor.username}
          avatarUrl={notification.actor.avatar_url || null}
          size="md"
        />
        <div style={{ flex: 1, minWidth: 0 }}>
          <p style={{ margin: 0, fontSize: 14, color: 'var(--text)', lineHeight: 1.4 }}>
            <span style={{ fontWeight: 600 }}>{notification.actor.name}</span>
            {' '}
            <span style={{ color: 'var(--text-secondary)' }}>{notification.message}</span>
          </p>
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 3, display: 'block' }}>
            {formatRelativeTime(notification.created_at)}
          </span>
        </div>
      </Link>

      <button
        onClick={() => onDelete(notification.id)}
        disabled={isDeleting}
        aria-label="Delete notification"
        style={{
          flexShrink: 0,
          background: 'none',
          border: 'none',
          cursor: isDeleting ? 'default' : 'pointer',
          opacity: isDeleting ? 0.4 : 0.6,
          padding: 6,
          borderRadius: 6,
          color: 'var(--text-tertiary)',
          display: 'flex',
          alignItems: 'center',
          transition: 'opacity 0.12s, color 0.12s',
        }}
        onMouseEnter={(e) => { if (!isDeleting) (e.currentTarget as HTMLButtonElement).style.color = 'var(--error, #ef4444)'; (e.currentTarget as HTMLButtonElement).style.opacity = '1'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.color = 'var(--text-tertiary)'; (e.currentTarget as HTMLButtonElement).style.opacity = isDeleting ? '0.4' : '0.6'; }}
      >
        <Trash2 size={15} />
      </button>
    </div>
  );
}
