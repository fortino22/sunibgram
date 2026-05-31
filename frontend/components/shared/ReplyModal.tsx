'use client';

import { X } from 'lucide-react';
import { useComposeStore } from '@/stores/composeStore';
import { PostComposer } from './PostComposer';
import { Avatar } from './Avatar';

export function ReplyModal() {
  const replyThread = useComposeStore((s) => s.replyThread);
  const closeReply = useComposeStore((s) => s.closeReply);

  if (!replyThread) return null;

  const author = replyThread.author;

  return (
    <div
      onClick={closeReply}
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.55)',
        zIndex: 200,
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'center',
        padding: '80px 0',
        overflowY: 'auto',
        animation: 'modal-backdrop-in 0.18s ease',
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: '100%',
          maxWidth: 600,
          height: 'fit-content',
          background: 'var(--bg-card)',
          borderRadius: 16,
          boxShadow: '0 8px 40px rgba(0,0,0,0.4)',
          margin: '0 16px',
          animation: 'modal-panel-in 0.22s cubic-bezier(0.16, 1, 0.3, 1)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', padding: '12px 16px 0' }}>
          <button
            onClick={closeReply}
            className="hover-dim"
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 6,
              borderRadius: 8,
            }}
          >
            <X size={20} />
          </button>
        </div>

        <div style={{ padding: '8px 20px 0', display: 'flex', gap: 12 }}>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
            <Avatar username={author.username} avatarUrl={author.avatar_url ?? null} size="md" />
            <div
              style={{
                width: 2,
                flex: 1,
                minHeight: 24,
                background: 'var(--border)',
                marginTop: 6,
                borderRadius: 1,
              }}
            />
          </div>
          <div style={{ flex: 1, minWidth: 0, paddingBottom: 12 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
              <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)' }}>
                {author.name || author.username}
              </span>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
                @{author.username}
              </span>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
                · {replyThread.time_ago}
              </span>
            </div>
            <p
              style={{
                fontSize: 14,
                color: 'var(--text)',
                lineHeight: 1.55,
                margin: 0,
                wordBreak: 'break-word',
              }}
            >
              {replyThread.content}
            </p>
            <div style={{ marginTop: 10, fontSize: 13, color: 'var(--text-tertiary)' }}>
              Replying to{' '}
              <span style={{ color: 'var(--accent)' }}>@{author.username}</span>
            </div>
          </div>
        </div>

        <PostComposer
          placeholder="Post your reply"
          parentId={replyThread.id}
          onSuccess={closeReply}
        />
      </div>
    </div>
  );
}
