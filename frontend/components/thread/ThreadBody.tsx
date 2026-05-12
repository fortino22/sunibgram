'use client';

import { cn } from '@/lib/utils';
import type { ThreadMediaResponse } from '@/lib/services/threads';

interface ThreadBodyProps {
  content: string;
  media: ThreadMediaResponse[];
  compact?: boolean;
  /** Called with the clicked image index. Parent is responsible for opening the modal. */
  onImageClick?: (index: number) => void;
}

export function ThreadBody({ content, media, compact = false, onImageClick }: ThreadBodyProps) {
  return (
    <div className="min-w-0">
      {content.trim().length > 0 && (
        <div
          className={cn('whitespace-pre-wrap break-words', compact ? 'text-sm' : 'text-[15px]')}
          style={{ color: 'var(--text)', lineHeight: 1.6 }}
        >
          {content}
        </div>
      )}

      {media.length > 0 && (
        <div
          style={{
            display: 'grid',
            gap: 8,
            marginTop: content.trim().length > 0 ? (compact ? 12 : 16) : 8,
            gridTemplateColumns: media.length === 1 ? '1fr' : 'repeat(2, 1fr)',
          }}
        >
          {media.map((item, i) =>
            media.length === 1 ? (
              <img
                key={item.media_id}
                src={item.url}
                alt=""
                onClick={onImageClick ? (e) => { e.stopPropagation(); onImageClick(i); } : undefined}
                style={{
                  width: 'auto',
                  maxWidth: '100%',
                  height: 'auto',
                  maxHeight: 480,
                  display: 'block',
                  borderRadius: 16,
                  background: 'var(--bg-subtle)',
                  cursor: onImageClick ? 'zoom-in' : 'default',
                }}
              />
            ) : (
              <div
                key={item.media_id}
                onClick={onImageClick ? (e) => { e.stopPropagation(); onImageClick(i); } : undefined}
                style={{
                  borderRadius: 16,
                  overflow: 'hidden',
                  background: 'var(--bg-subtle)',
                  height: 220,
                  cursor: onImageClick ? 'zoom-in' : 'default',
                }}
              >
                <img
                  src={item.url}
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
