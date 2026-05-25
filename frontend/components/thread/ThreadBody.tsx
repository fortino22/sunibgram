'use client';

import Link from 'next/link';
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

function renderContent(content: string) {
  const parts = content.split(/(#[a-zA-Z0-9_]+)/g);
  return parts.map((part, i) => {
    if (/^#[a-zA-Z0-9_]+$/.test(part)) {
      return (
        <Link
          key={i}
          href={`/explore?q=${encodeURIComponent(part)}`}
          onClick={(e) => e.stopPropagation()}
          style={{ color: 'var(--accent)' }}
          className="hover:underline"
        >
          {part}
        </Link>
      );
    }
    return <span key={i}>{part}</span>;
  });
}

export function ThreadBody({ content, media, compact = false, onImageClick }: ThreadBodyProps) {
  const hasContent = content.trim().length > 0;

  return (
    <div className="min-w-0">
      {hasContent && (
        <div
          className={cn('whitespace-pre-wrap wrap-break-word', compact ? 'text-sm' : 'text-[15px]')}
          style={{ color: 'var(--text)', lineHeight: 1.6 }}
        >
          {renderContent(content)}
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
