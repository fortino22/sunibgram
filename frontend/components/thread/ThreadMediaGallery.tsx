'use client';

import type { ThreadMediaResponse } from '@/lib/services/threads';

interface ThreadMediaGalleryProps {
  media: ThreadMediaResponse[];
  compact?: boolean;
  onImageClick?: (index: number) => void;
  topMargin?: number;
}

export function ThreadMediaGallery({
  media,
  compact = false,
  onImageClick,
  topMargin = 8,
}: ThreadMediaGalleryProps) {
  if (media.length === 0) {
    return null;
  }

  return (
    <div
      style={{
        display: 'grid',
        gap: 8,
        marginTop: topMargin,
        gridTemplateColumns: media.length === 1 ? '1fr' : 'repeat(2, 1fr)',
      }}
    >
      {media.map((item, i) =>
        media.length === 1 ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            key={item.media_id}
            src={item.url}
            alt=""
            onClick={onImageClick ? (e) => {
              e.stopPropagation();
              onImageClick(i);
            } : undefined}
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
            onClick={onImageClick ? (e) => {
              e.stopPropagation();
              onImageClick(i);
            } : undefined}
            style={{
              borderRadius: 16,
              overflow: 'hidden',
              background: 'var(--bg-subtle)',
              height: compact ? 180 : 220,
              cursor: onImageClick ? 'zoom-in' : 'default',
            }}
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={item.url}
              alt=""
              style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
            />
          </div>
        )
      )}
    </div>
  );
}
