'use client';

import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import type { FeedParentContext } from '@/lib/services/threads';
import { renderRichContent } from './MentionText';

interface QuotedPostBoxProps {
  context: FeedParentContext;
}

export function QuotedPostBox({ context }: QuotedPostBoxProps) {
  const router = useRouter();
  const avatarUrl = context.author.avatar_url ?? null;
  const images = context.media.filter((m) => /\.(jpe?g|png|webp|gif)$/i.test(m.url));

  return (
    <div
      role="link"
      tabIndex={0}
      onClick={(e) => { e.stopPropagation(); router.push(`/thread/${context.id}`); }}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.stopPropagation(); router.push(`/thread/${context.id}`); } }}
      style={{
        border: '1px solid var(--border)',
        borderRadius: 12,
        padding: '10px 12px',
        background: 'var(--bg-subtle)',
        cursor: 'pointer',
        marginTop: 10,
      }}
    >
      {/* Author row */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
        <Avatar username={context.author.username} avatarUrl={avatarUrl} size={18} />
        <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--text)', letterSpacing: '-0.01em' }}>
          {context.author.name || context.author.username}
        </span>
        <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
          @{context.author.username}
        </span>
      </div>

      {/* Content preview */}
      {context.content && (
        <p
          style={{
            fontSize: 14,
            color: 'var(--text)',
            lineHeight: 1.5,
            margin: 0,
            display: '-webkit-box',
            WebkitLineClamp: 3,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
          }}
        >
          {renderRichContent(context.content)}
        </p>
      )}

      {/* Media thumbnails — show all images in a compact grid */}
      {images.length > 0 && (
        <div
          style={{
            marginTop: 8,
            display: 'grid',
            gap: 4,
            gridTemplateColumns: images.length === 1 ? '1fr' : 'repeat(2, 1fr)',
          }}
        >
          {images.map((img) =>
            images.length === 1 ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                key={img.media_id}
                src={img.url}
                alt=""
                style={{
                  width: '100%',
                  maxHeight: 160,
                  objectFit: 'cover',
                  borderRadius: 8,
                  display: 'block',
                }}
              />
            ) : (
              <div
                key={img.media_id}
                style={{ borderRadius: 8, overflow: 'hidden', height: 100, background: 'var(--bg-card)' }}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={img.url}
                  alt=""
                  style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
                />
              </div>
            )
          )}
        </div>
      )}
    </div>
  );
}
