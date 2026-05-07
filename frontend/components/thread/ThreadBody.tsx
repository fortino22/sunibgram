import { cn } from '@/lib/utils';
import type { ThreadMediaResponse } from '@/lib/services/threads';

interface ThreadBodyProps {
  content: string;
  media: ThreadMediaResponse[];
  compact?: boolean;
}

function mediaHeightClassName(count: number): string {
  if (count === 1) {
    return 'max-h-[400px]';
  }

  if (count === 2) {
    return 'max-h-[250px]';
  }

  return 'max-h-[200px]';
}

export function ThreadBody({ content, media, compact = false }: ThreadBodyProps) {
  return (
    <div className="min-w-0">
      <div
        className={cn('whitespace-pre-wrap break-words', compact ? 'text-sm' : 'text-[15px]')}
        style={{ color: 'var(--text)', lineHeight: 1.6 }}
      >
        {content}
      </div>

      {media.length > 0 && (
        <div className={cn('mt-3 grid gap-2', media.length === 1 ? 'grid-cols-1' : 'grid-cols-2')}>
          {media.map((item) => (
            <img
              key={item.media_id}
              src={item.url}
              alt=""
              className={cn('block w-full rounded-lg object-cover', mediaHeightClassName(media.length))}
              style={{ background: 'var(--bg-subtle)' }}
            />
          ))}
        </div>
      )}
    </div>
  );
}
