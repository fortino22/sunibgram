'use client';

import { useState } from 'react';
import { ArrowLeft, Search } from 'lucide-react';
import { toast } from 'sonner';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useSearchShareUsers } from '@/hooks/useSearchShareUsers';
import { useShareViaChat } from '@/hooks/useShareThread';
import type { ShareUser } from '@/lib/services/share';

interface ShareModalProps {
  threadId: number;
  onClose: () => void;
}

export function ShareModal({ threadId, onClose }: ShareModalProps) {
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<ShareUser[]>([]);
  const [message, setMessage] = useState('');
  const [searchFocused, setSearchFocused] = useState(false);

  const searchResult = useSearchShareUsers(query);
  const sendChat = useShareViaChat();

  const selectedIds = new Set(selected.map((u) => u.user_id));
  const isSelected = (id: number) => selectedIds.has(id);
  const toggle = (u: ShareUser) => {
    setSelected((prev) =>
      prev.some((s) => s.user_id === u.user_id)
        ? prev.filter((s) => s.user_id !== u.user_id)
        : [...prev, u],
    );
  };

  const hasQuery = query.trim().length > 0;
  const results = hasQuery ? (searchResult.data ?? []) : [];
  const visibleResults = results.filter((u) => !selectedIds.has(u.user_id));
  const showFirstLoadSkeleton = hasQuery && searchResult.isLoading && !searchResult.data;
  const hasAnythingToShow = selected.length > 0 || visibleResults.length > 0;

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
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '14px 20px' }}>
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
            Share
            {selected.length > 0 && (
              <span style={{ marginLeft: 10, color: 'var(--text-muted)', fontWeight: 500 }}>
                {selected.length} selected
              </span>
            )}
          </h2>
        </div>

        <div style={{ padding: '0 16px 8px' }}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              padding: '10px 14px',
              borderRadius: 999,
              background: 'var(--bg)',
              border: `1px solid ${searchFocused ? 'var(--accent)' : 'transparent'}`,
              transition: 'border-color 0.15s ease',
            }}
          >
            <Search size={16} style={{ color: searchFocused ? 'var(--accent)' : 'var(--text-muted)' }} />
            <input
              type="text"
              placeholder="Search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onFocus={() => setSearchFocused(true)}
              onBlur={() => setSearchFocused(false)}
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
        </div>

        <div style={{ flex: 1, overflowY: 'auto', padding: '4px 8px 8px' }}>
          {selected.length > 0 && (
            <div style={{ marginBottom: 4 }}>
              {selected.map((u) => (
                <UserRow key={u.user_id} user={u} selected onToggle={() => toggle(u)} />
              ))}
            </div>
          )}

          {showFirstLoadSkeleton ? (
            Array.from({ length: 4 }, (_, i) => (
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
          ) : (
            visibleResults.map((u) => (
              <UserRow key={u.user_id} user={u} selected={isSelected(u.user_id)} onToggle={() => toggle(u)} />
            ))
          )}

          {!showFirstLoadSkeleton && !hasAnythingToShow && (
            <div style={{ textAlign: 'center', padding: '32px 20px', color: 'var(--text-muted)', fontSize: 13 }}>
              {query.trim() ? 'No users found' : 'Search for someone to share with'}
            </div>
          )}
        </div>

        {selected.length > 0 && (
          <div style={{ padding: '10px 16px 14px', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <input
              type="text"
              placeholder="Write a message…"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              maxLength={5000}
              style={{
                width: '100%',
                background: 'var(--bg)',
                border: '1px solid transparent',
                borderRadius: 12,
                padding: '11px 14px',
                fontSize: 14,
                color: 'var(--text)',
                fontFamily: 'inherit',
                outline: 'none',
                boxSizing: 'border-box',
                transition: 'border-color 0.15s ease',
              }}
              onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--accent)'; }}
              onBlur={(e) => { e.currentTarget.style.borderColor = 'transparent'; }}
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

interface UserRowProps {
  user: ShareUser;
  selected: boolean;
  onToggle: () => void;
}

function UserRow({ user, selected, onToggle }: UserRowProps) {
  return (
    <button
      onClick={onToggle}
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
      <Avatar username={user.username} avatarUrl={user.avatar_url ?? null} size={42} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text)', display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {user.name || user.username}
          </span>
          {user.is_premium && <VerifiedBadge size={15} />}
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>@{user.username}</div>
      </div>
      <div
        style={{
          width: 22,
          height: 22,
          borderRadius: '50%',
          border: selected ? 'none' : '2px solid var(--border)',
          background: selected ? 'var(--accent)' : 'transparent',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        {selected && (
          <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
            <path d="M2.5 6.5l3 3 5-5" stroke="#fff" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        )}
      </div>
    </button>
  );
}
