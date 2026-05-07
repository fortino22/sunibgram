import { Avatar } from '@/components/shared/Avatar';
import type { ThreadItem } from '@/lib/services/threads';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadHeader } from './ThreadHeader';
import { getAuthorAvatarUrl, getThreadAuthor } from './threadUtils';

interface ThreadReplyItemProps {
  thread: ThreadItem;
  compact?: boolean;
}

export function ThreadReplyItem({ thread, compact = false }: ThreadReplyItemProps) {
  const author = getThreadAuthor(thread);

  return (
    <div className="flex gap-3">
      <Avatar
        username={author.username}
        avatarUrl={getAuthorAvatarUrl(author)}
        size={compact ? 'sm' : 'md'}
      />
      <div className="min-w-0 flex-1">
        <ThreadHeader author={author} timeAgo={thread.time_ago} compact={compact} showFaculty />
        <ThreadBody content={thread.content} media={thread.media} compact={compact} />
        <ThreadActions
          likes={thread.stats.likes}
          reposts={thread.stats.reposts}
          bookmarks={thread.stats.bookmarks}
          isLiked={thread.is_liked === true}
          isReposted={thread.is_reposted === true}
          isBookmarked={thread.is_bookmarked === true}
          compact={compact}
        />
      </div>
    </div>
  );
}
