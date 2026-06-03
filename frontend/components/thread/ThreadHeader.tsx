import type { FeedParentContext } from '@/lib/services/threads';
import { cn, shortTimeAgo } from '@/lib/utils';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import type { ThreadAuthorWithAvatar } from './threadUtils';
import { truncateContent } from './threadUtils';

interface ThreadHeaderProps {
  author: ThreadAuthorWithAvatar;
  timeAgo?: string;
  parentContext?: FeedParentContext;
  repostLabel?: string | null;
  compact?: boolean;
  showFaculty?: boolean;
  isPinned?: boolean;
}

export function ThreadHeader({
  author,
  timeAgo,
  parentContext,
  repostLabel,
  compact = false,
  showFaculty = false,
  isPinned = false,
}: ThreadHeaderProps) {
  return (
    <div className="min-w-0">
      {repostLabel && (
        <div className="mb-2 text-[13px] font-medium" style={{ color: 'var(--text-tertiary)' }}>
          ↺ {repostLabel}
        </div>
      )}

      <div className={cn('mb-1.5 flex min-w-0 items-center gap-1.5', compact && 'mb-1 flex-wrap')}>
        {/* Name + username share remaining width and BOTH truncate with
            ellipsis (X-style). Badges/time stay pinned to the right and
            never shrink. */}
        <div className="flex min-w-0 flex-1 items-center gap-1.5">
          <span className="min-w-0 truncate text-sm font-bold" style={{ color: 'var(--text)' }}>
            {author.name || author.username}
          </span>
          {author.is_premium && <VerifiedBadge size={15} />}
          <span className="min-w-0 truncate text-sm" style={{ color: 'var(--text-tertiary)' }}>
            @{author.username}
          </span>
        </div>
        {showFaculty && author.faculty && (
          <span
            className="faculty-pill shrink-0 rounded-full border px-2 py-0.5 text-xs font-bold"
            title={author.faculty.faculty_name}
            style={{
              color: 'var(--accent)',
              background: 'var(--bg-subtle)',
              borderColor: 'var(--border)',
              maxWidth: '40%',
              whiteSpace: 'nowrap',
              overflowX: 'auto',
              overflowY: 'hidden',
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
            <span
              className="shrink-0 text-sm"
              style={{ color: 'var(--text-tertiary)' }}
              title={timeAgo}
            >
              {shortTimeAgo(timeAgo)}
            </span>
          </>
        )}
        {isPinned && (
          <span className="ml-auto shrink-0 text-xs" style={{ color: 'var(--text-tertiary)' }}>
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
