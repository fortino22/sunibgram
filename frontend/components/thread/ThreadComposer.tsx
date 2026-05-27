'use client';

import { useState } from 'react';
import { Avatar } from '@/components/shared/Avatar';
import { PostComposer } from '@/components/shared/PostComposer';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useAuthStore } from '@/stores/authStore';

interface ThreadComposerProps {
  placeholder: string;
  parentId: number;
}

export function ThreadComposer({ placeholder, parentId }: ThreadComposerProps) {
  const [expanded, setExpanded] = useState(false);
  const user = useAuthStore((state) => state.user);
  const { data: profile } = useMyProfile();
  const avatarSrc = profile?.avatar ?? null;

  if (expanded) {
    return (
      <PostComposer
        placeholder={placeholder}
        parentId={parentId}
        autoFocus
        onSuccess={() => setExpanded(false)}
      />
    );
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => setExpanded(true)}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') setExpanded(true); }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '12px 20px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
        cursor: 'text',
      }}
    >
      <Avatar username={user?.username ?? ''} avatarUrl={avatarSrc} size="md" />

      <span style={{ flex: 1, fontSize: 15, color: 'var(--text-tertiary)', userSelect: 'none' }}>
        {placeholder}
      </span>

      <button
        type="button"
        onClick={(e) => { e.stopPropagation(); setExpanded(true); }}
        style={{
          background: 'var(--bg-subtle)',
          border: '1px solid var(--border)',
          borderRadius: 20,
          color: 'var(--text)',
          fontWeight: 600,
          fontSize: 14,
          padding: '7px 18px',
          cursor: 'pointer',
          fontFamily: 'inherit',
          flexShrink: 0,
        }}
      >
        Reply
      </button>
    </div>
  );
}
