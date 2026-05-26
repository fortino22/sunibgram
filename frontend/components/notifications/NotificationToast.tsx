'use client';

import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import { X } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useNotificationStore } from '@/stores/notificationStore';
import { getNotificationTypeConfig } from './notificationTypeUtils';
import type { ToastItem } from '@/stores/notificationStore';

const TOAST_DURATION_MS = 4500;

interface NotificationToastItemProps {
  item: ToastItem;
}

function resolveHref(type: string, referenceId: number | null): string {
  switch (type) {
    case 'follow':       return '#'; // resolved per-item below via actor
    case 'chat_message': return referenceId ? `/messages?room=${referenceId}` : '/messages';
    default:             return referenceId ? `/thread/${referenceId}` : '#';
  }
}

function NotificationToastItem({ item }: NotificationToastItemProps) {
  const removeToast = useNotificationStore((s) => s.removeToast);
  const [visible, setVisible] = useState(false);
  const [leaving, setLeaving] = useState(false);

  const dismiss = useCallback(() => {
    setLeaving(true);
    setTimeout(() => removeToast(item.id), 280);
  }, [item.id, removeToast]);

  useEffect(() => {
    const enterTimer = setTimeout(() => setVisible(true), 16);
    const exitTimer  = setTimeout(dismiss, TOAST_DURATION_MS);
    return () => {
      clearTimeout(enterTimer);
      clearTimeout(exitTimer);
    };
  }, [dismiss]);

  const { notification } = item;
  const { Icon, color } = getNotificationTypeConfig(notification.type);
  const href =
    notification.type === 'follow'
      ? `/profile/${notification.actor.user_id}`
      : resolveHref(notification.type, notification.reference_id);

  const show = visible && !leaving;

  return (
    <div
      role="alert"
      style={{
        transform: show ? 'translateX(0) scale(1)' : 'translateX(110%) scale(0.95)',
        opacity: show ? 1 : 0,
        transition: 'transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.25s ease',
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderLeft: `3px solid ${color}`,
        borderRadius: 12,
        padding: '12px 10px 12px 14px',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
        maxWidth: 340,
        width: '100%',
        pointerEvents: 'all',
      }}
    >
      <Link
        href={href}
        onClick={dismiss}
        style={{ display: 'flex', alignItems: 'center', gap: 10, flex: 1, minWidth: 0, textDecoration: 'none' }}
      >
        <div style={{ position: 'relative', flexShrink: 0 }}>
          <Avatar
            username={notification.actor.username}
            avatarUrl={notification.actor.avatar_url || null}
            size="sm"
          />
          <div
            style={{
              position: 'absolute',
              bottom: -2,
              right: -2,
              background: color,
              borderRadius: '50%',
              width: 17,
              height: 17,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: '2px solid var(--bg-card)',
            }}
          >
            <Icon size={9} color="#fff" strokeWidth={2.5} />
          </div>
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--text)', lineHeight: 1.4, overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' }}>
            <span style={{ fontWeight: 600 }}>{notification.actor.name}</span>
            {' '}
            <span style={{ color: 'var(--text-secondary)' }}>{notification.message}</span>
          </p>
        </div>
      </Link>

      <button
        onClick={dismiss}
        aria-label="Dismiss notification"
        style={{
          flexShrink: 0,
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          color: 'var(--text-tertiary)',
          padding: '4px',
          borderRadius: 6,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          transition: 'color 0.12s',
        }}
        onMouseEnter={(e) => { e.currentTarget.style.color = 'var(--text)'; }}
        onMouseLeave={(e) => { e.currentTarget.style.color = 'var(--text-tertiary)'; }}
      >
        <X size={14} />
      </button>
    </div>
  );
}

export function NotificationToastContainer() {
  const toasts = useNotificationStore((s) => s.toasts);

  if (toasts.length === 0) return null;

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 24,
        right: 24,
        zIndex: 9999,
        display: 'flex',
        flexDirection: 'column-reverse',
        gap: 8,
        pointerEvents: 'none',
        maxWidth: 340,
        width: '100%',
      }}
    >
      {toasts.map((item) => (
        <NotificationToastItem key={item.id} item={item} />
      ))}
    </div>
  );
}
