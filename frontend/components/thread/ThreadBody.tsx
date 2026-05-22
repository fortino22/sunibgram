'use client';

import { cn } from '@/lib/utils';
import type { ThreadMediaResponse } from '@/lib/services/threads';
import { ThreadMediaGallery } from './ThreadMediaGallery';

interface ThreadBodyProps {
  content: string;
  media: ThreadMediaResponse[];
  compact?: boolean;
  /** Called with the clicked image index. Parent is responsible for opening the modal. */
  onImageClick?: (index: number) => void;
}

export function ThreadBody({ content, media, compact = false, onImageClick }: ThreadBodyProps) {
  const hasContent = content.trim().length > 0;

  return (
    <div className="min-w-0">
      {hasContent && (
        <div
          className={cn('whitespace-pre-wrap break-words', compact ? 'text-sm' : 'text-[15px]')}
          style={{ color: 'var(--text)', lineHeight: 1.6 }}
        >
          {content}
        </div>
      )}

      <ThreadMediaGallery
        media={media}
        compact={compact}
        onImageClick={onImageClick}
        topMargin={hasContent ? (compact ? 12 : 16) : 8}
      />
    </div>
  );
}
