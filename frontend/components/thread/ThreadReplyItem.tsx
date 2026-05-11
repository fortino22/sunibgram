import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import type { ThreadItem } from '@/lib/services/threads';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadHeader } from './ThreadHeader';
import { getAuthorAvatarUrl, getThreadAuthor, isRepostThread } from './threadUtils';

interface ThreadReplyItemProps {
  thread: ThreadItem;
  compact?: boolean;
}

export function ThreadReplyItem({ thread, compact = false }: ThreadReplyItemProps) {
  const router = useRouter();
  const author = getThreadAuthor(thread);

  return (
    <div
      className="flex cursor-pointer gap-3"
      onClick={() => router.push(`/thread/${thread.id}`)}
    >
      <Avatar
        username={author.username}
        avatarUrl={getAuthorAvatarUrl(author)}
        size={compact ? 'sm' : 'md'}
      />
      <div className="min-w-0 flex-1">
        <ThreadHeader
          author={author}
          timeAgo={thread.time_ago}
          compact={compact}
          parentContext={thread.parent_context}
          repostedBy={isRepostThread(thread) ? thread.author.username : null}
        />
        <ThreadBody content={thread.content} media={thread.media} compact={compact} />
        <div className="mt-3 flex items-center justify-between gap-3">
          <ThreadActions
            comments={thread.stats.comments}
            likes={thread.stats.likes}
            reposts={thread.stats.reposts}
            bookmarks={thread.stats.bookmarks}
            isLiked={thread.is_liked === true}
            isReposted={thread.is_reposted === true}
            isBookmarked={thread.is_bookmarked === true}
            compact={compact}
            className="mt-0 max-w-none flex-1 justify-start"
            onActionClick={(event) => event.stopPropagation()}
          />
        </div>
      </div>
    </div>
  );
}
