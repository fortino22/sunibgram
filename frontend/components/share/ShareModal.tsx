'use client';

import { useState } from 'react';
import { ArrowLeft, Link2, Search } from 'lucide-react';
import { toast } from 'sonner';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useSearchShareUsers } from '@/hooks/useSearchShareUsers';
import { useRecordShare, useShareViaChat } from '@/hooks/useShareThread';
import { buildThreadShareUrl } from '@/lib/shareUrl';
import type { ShareUser } from '@/lib/services/share';

interface ShareModalProps {
  threadId: number;
  onClose: () => void;
}

export function ShareModal({ threadId, onClose }: ShareModalProps) {
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<ShareUser[]>([]);
  const [message, setMessage] = useState('');

  const searchResult = useSearchShareUsers(query);
  const users = searchResult.data ?? [];

  const recordShare = useRecordShare(threadId);
  const sendChat = useShareViaChat();

  const isSelected = (id: number) => selected.some((u) => u.user_id === id);
  const toggle = (u: ShareUser) => {
    setSelected((prev) =>
      prev.some((s) => s.user_id === u.user_id)
        ? prev.filter((s) => s.user_id !== u.user_id)
        : [...prev, u],
    );
  };

  async function handleCopyLink() {
    const url = buildThreadShareUrl(threadId);
    try {
      await navigator.clipboard.writeText(url);
      toast.success('Link copied to clipboard');
      recordShare.mutate();
      onClose();
    } catch {
      toast.error('Failed to copy link');
    }
  }

  function handleSend() {
    if (selected.length === 0 || sendChat.isPending) return;
    sendChat.mutate(
      {
        threadId,
        userIds: selected.map((u) => u.user_id),
        message: message.trim() || undefined,
      },
      {
        onSuccess: (data) => {
          toast.success(`Sent to ${data.recipient_count} recipient${data.recipient_count === 1 ? '' : 's'}`);
          onClose();
        },
        onError: (err) => {
          toast.error(err.message || 'Failed to send');
        },
      },
    );
  }

  return (
    <div
      role="dialog"
      aria-label="Share"
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          width: 560,
          maxWidth: 'calc(100vw - 32px)',
          maxHeight: '85vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <button
              onClick={onClose}
              aria-label="Close"
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--text)',
                cursor: 'pointer',
                padding: 4,
                display: 'flex',
              }}
            >
              <ArrowLeft size={20} />
            </button>
            <h2 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>
              Share{selected.length > 0 ? <span style={{ marginLeft: 8, color: 'var(--text-muted)', fontWeight: 500 }}>{selected.length} selected</span> : null}
            </h2>
          </div>
          <button
            onClick={handleCopyLink}
            disabled={recordShare.isPending}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 12px',
              background: 'var(--bg)',
              border: '1px solid var(--border)',
              borderRadius: 999,
              color: 'var(--text)',
              fontSize: 13,
              fontWeight: 500,
              cursor: recordShare.isPending ? 'wait' : 'pointer',
            }}
          >
            <Link2 size={14} />
            Copy link
          </button>
        </div>

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: '12px 16px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <Search size={16} style={{ color: 'var(--text-muted)' }} />
          <input
            type="text"
            placeholder="Search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
            style={{
              flex: 1,
              background: 'none',
              border: 'none',
              outline: 'none',
              fontSize: 14,
              color: 'var(--text)',
              fontFamily: 'inherit',
            }}
          />
        </div>

        <div style={{ flex: 1, overflowY: 'auto', padding: '8px 8px 8px' }}>
          {searchResult.isLoading ? (
            Array.from({ length: 5 }, (_, i) => (
              <div
                key={i}
                style={{
                  height: 60,
                  borderRadius: 10,
                  background: 'var(--skeleton, oklch(16% 0.005 260))',
                  marginBottom: 4,
                  animation: 'pulse 1.5s ease-in-out infinite',
                }}
              />
            ))
          ) : users.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '32px 20px', color: 'var(--text-muted)', fontSize: 13 }}>
              {query.trim() ? 'No users found' : 'Search for someone to share with'}
            </div>
          ) : (
            users.map((u) => {
              const sel = isSelected(u.user_id);
              return (
                <button
                  key={u.user_id}
                  onClick={() => toggle(u)}
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 12,
                    padding: '10px 12px',
                    background: 'none',
                    border: 'none',
                    borderRadius: 10,
                    cursor: 'pointer',
                    textAlign: 'left',
                  }}
                  onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--sidebar-hover)'; }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; }}
                >
                  <Avatar username={u.username} avatarUrl={u.avatar_url ?? null} size={42} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text)', display: 'flex', alignItems: 'center', gap: 4 }}>
                      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {u.name || u.username}
                      </span>
                      {u.is_premium && <VerifiedBadge size={15} />}
                    </div>
                    <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>@{u.username}</div>
                  </div>
                  <div
                    style={{
                      width: 22,
                      height: 22,
                      borderRadius: '50%',
                      border: sel ? 'none' : '2px solid var(--border)',
                      background: sel ? 'var(--accent)' : 'transparent',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      flexShrink: 0,
                    }}
                  >
                    {sel && (
                      <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
                        <path d="M2.5 6.5l3 3 5-5" stroke="#fff" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                    )}
                  </div>
                </button>
              );
            })
          )}
        </div>

        {selected.length > 0 && (
          <div style={{ padding: '12px 16px', borderTop: '1px solid var(--border)', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <input
              type="text"
              placeholder="Write a message…"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              maxLength={5000}
              style={{
                width: '100%',
                background: 'var(--bg)',
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: '11px 14px',
                fontSize: 14,
                color: 'var(--text)',
                fontFamily: 'inherit',
                outline: 'none',
                boxSizing: 'border-box',
              }}
            />
            <button
              onClick={handleSend}
              disabled={sendChat.isPending}
              style={{
                width: '100%',
                padding: '13px',
                borderRadius: 999,
                background: sendChat.isPending ? 'var(--border)' : 'var(--text)',
                border: 'none',
                color: sendChat.isPending ? 'var(--text-muted)' : 'var(--bg-card)',
                fontWeight: 600,
                fontSize: 15,
                cursor: sendChat.isPending ? 'wait' : 'pointer',
              }}
            >
              {sendChat.isPending
                ? 'Sending…'
                : selected.length === 1
                ? 'Send'
                : 'Send separately'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
