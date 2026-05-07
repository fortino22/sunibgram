import type { FeedParentContext } from '@/lib/services/threads';
import { cn } from '@/lib/utils';
import type { ThreadAuthorWithAvatar } from './threadUtils';
import { truncateContent } from './threadUtils';

interface ThreadHeaderProps {
  author: ThreadAuthorWithAvatar;
  timeAgo?: string;
  parentContext?: FeedParentContext;
  repostedBy?: string | null;
  compact?: boolean;
  showFaculty?: boolean;
  isPinned?: boolean;
}

export function ThreadHeader({
  author,
  timeAgo,
  parentContext,
  repostedBy,
  compact = false,
  showFaculty = false,
  isPinned = false,
}: ThreadHeaderProps) {
  return (
    <div className="min-w-0">
      {repostedBy && (
        <div className="mb-2 text-[13px] font-medium" style={{ color: 'var(--text-tertiary)' }}>
          ↺ @{repostedBy} reposted
        </div>
      )}

      <div className={cn('mb-1.5 flex min-w-0 items-center gap-1.5', compact && 'mb-1 flex-wrap')}>
        <span className="truncate text-sm font-bold" style={{ color: 'var(--text)' }}>
          @{author.username}
        </span>
        {author.is_premium && (
          <span aria-label="Premium" title="Premium" className="shrink-0 text-sm" style={{ color: 'var(--amber)' }}>
            ★
          </span>
        )}
        {showFaculty && author.faculty && (
          <span
            className="rounded-full border px-2 py-0.5 text-xs font-bold"
            style={{
              color: 'var(--accent)',
              background: 'var(--bg-subtle)',
              borderColor: 'var(--border)',
            }}
          >
            {author.faculty.faculty_name}
          </span>
        )}
        {timeAgo && (
          <>
            <span className="shrink-0 text-sm" style={{ color: 'var(--text-tertiary)' }}>
              ·
            </span>
            <span className="shrink-0 text-sm" style={{ color: 'var(--text-tertiary)' }}>
              {timeAgo}
            </span>
          </>
        )}
        {isPinned && (
          <span className="ml-auto text-xs" style={{ color: 'var(--text-tertiary)' }}>
            Pinned
          </span>
        )}
      </div>

      {parentContext && (
        <div
          className="mb-2.5 rounded-lg border px-2.5 py-2 text-[13px]"
          style={{
            borderColor: 'var(--border)',
            color: 'var(--text-tertiary)',
          }}
        >
          <span className="font-bold" style={{ color: 'var(--text)' }}>
            @{parentContext.author.username}
          </span>{' '}
          <span>{truncateContent(parentContext.content)}</span>
        </div>
      )}
    </div>
  );
}
