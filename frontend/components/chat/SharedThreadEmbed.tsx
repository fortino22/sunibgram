'use client';

import { useRouter } from 'next/navigation';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import type { SharedThreadEmbed as SharedThreadEmbedType } from '@/lib/services/chat';

interface SharedThreadEmbedProps {
  embed: SharedThreadEmbedType;
  isOwn: boolean;
}

export function SharedThreadEmbed({ embed, isOwn }: SharedThreadEmbedProps) {
  const router = useRouter();

  if (embed.deleted || !embed.author) {
    return (
      <div
        style={{
          maxWidth: 260,
          padding: '12px 14px',
          borderRadius: 12,
          border: '1px dashed var(--border)',
          background: 'transparent',
          color: 'var(--text-muted)',
          fontSize: 13,
          fontStyle: 'italic',
        }}
      >
        This post is unavailable
      </div>
    );
  }

  const handleClick = () => {
    router.push(`/thread/${embed.thread_id}`);
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      style={{
        maxWidth: 280,
        textAlign: 'left',
        padding: 12,
        borderRadius: 12,
        background: isOwn ? 'rgba(255,255,255,0.10)' : 'var(--bg)',
        border: '1px solid var(--border)',
        color: 'var(--text)',
        cursor: 'pointer',
        display: 'flex',
        flexDirection: 'column',
        gap: 6,
        font: 'inherit',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, fontWeight: 600 }}>
        <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {embed.author.name || embed.author.username}
        </span>
        {embed.author.is_premium && <VerifiedBadge size={13} />}
        <span style={{ color: 'var(--text-muted)', fontWeight: 400 }}>
          @{embed.author.username}
        </span>
      </div>
      <div
        style={{
          fontSize: 13,
          lineHeight: 1.4,
          color: 'var(--text)',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          display: '-webkit-box',
          WebkitLineClamp: 4,
          WebkitBoxOrient: 'vertical',
          overflow: 'hidden',
        }}
      >
        {embed.content}
      </div>
      {embed.community_name ? (
        <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>
          in {embed.community_name}
        </div>
      ) : null}
    </button>
  );
}
