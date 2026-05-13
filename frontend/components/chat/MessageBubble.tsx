'use client';

import Image from 'next/image';
import { Check, CheckCheck } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import type { ChatMessage } from '@/lib/services/chat';

interface MessageBubbleProps {
  message: ChatMessage;
  isOwn: boolean;
  showSender: boolean;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function isImage(mediaType: string) {
  return mediaType.startsWith('image/');
}

export function MessageBubble({ message, isOwn, showSender }: MessageBubbleProps) {
  const isDeleted = !!message.deleted_at;

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: isOwn ? 'row-reverse' : 'row',
        alignItems: 'flex-end',
        gap: 8,
        marginBottom: showSender ? 12 : 4,
      }}
    >
      {/* Avatar — only for others */}
      {!isOwn && (
        <div style={{ flexShrink: 0, marginBottom: 2 }}>
          {showSender ? (
            <Avatar username={message.sender.username} size={28} />
          ) : (
            <div style={{ width: 28 }} />
          )}
        </div>
      )}

      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: isOwn ? 'flex-end' : 'flex-start',
          maxWidth: '68%',
        }}
      >
        {/* Sender name — shown once per group */}
        {!isOwn && showSender && (
          <span
            style={{
              fontSize: 11,
              fontWeight: 600,
              color: 'var(--accent)',
              marginBottom: 3,
              paddingLeft: 2,
            }}
          >
            {message.sender.username}
          </span>
        )}

        {/* Media */}
        {message.media.length > 0 && !isDeleted && (
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              gap: 4,
              marginBottom: message.message ? 4 : 0,
              justifyContent: isOwn ? 'flex-end' : 'flex-start',
            }}
          >
            {message.media.map((m) => (
              <div
                key={m.media_id}
                style={{
                  borderRadius: 10,
                  overflow: 'hidden',
                  maxWidth: 220,
                  background: 'var(--bg)',
                }}
              >
                {isImage(m.media_type) ? (
                  <Image
                    src={m.url}
                    alt="media"
                    width={220}
                    height={160}
                    style={{ objectFit: 'cover', display: 'block' }}
                    unoptimized
                  />
                ) : (
                  <video
                    src={m.url}
                    controls
                    style={{ maxWidth: 220, maxHeight: 160, display: 'block' }}
                  />
                )}
              </div>
            ))}
          </div>
        )}

        {/* Bubble */}
        {(message.message || isDeleted) && (
          <div
            style={{
              padding: '8px 12px',
              borderRadius: isOwn ? '14px 14px 4px 14px' : '14px 14px 14px 4px',
              background: isDeleted
                ? 'transparent'
                : isOwn
                ? 'var(--accent)'
                : 'var(--bg)',
              border: isDeleted
                ? '1px dashed var(--border)'
                : isOwn
                ? 'none'
                : '1px solid var(--border)',
              fontSize: 14,
              lineHeight: 1.45,
              color: isDeleted
                ? 'var(--text-muted)'
                : isOwn
                ? '#fff'
                : 'var(--text)',
              fontStyle: isDeleted ? 'italic' : 'normal',
              wordBreak: 'break-word',
              whiteSpace: 'pre-wrap',
            }}
          >
            {isDeleted ? 'This message was deleted' : message.message}
          </div>
        )}

        {/* Timestamp + read receipt */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 3,
            marginTop: 3,
            paddingLeft: isOwn ? 0 : 2,
            paddingRight: isOwn ? 2 : 0,
            flexDirection: isOwn ? 'row-reverse' : 'row',
          }}
        >
          <span style={{ fontSize: 10, color: 'var(--text-muted)' }}>
            {formatTime(message.created_at)}
          </span>
          {isOwn && !isDeleted && (
            message.is_read
              ? <CheckCheck size={12} color="var(--accent)" />
              : <Check size={12} color="var(--text-muted)" />
          )}
        </div>
      </div>
    </div>
  );
}
