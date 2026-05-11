'use client';

import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import type { ThreadItem } from '@/lib/services/threads';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadHeader } from './ThreadHeader';
import { getAuthorAvatarUrl, getThreadAuthor, isRepostThread } from './threadUtils';

interface ThreadCardProps {
  thread: ThreadItem;
}

export function ThreadCard({ thread }: ThreadCardProps) {
  const router = useRouter();
  const author = getThreadAuthor(thread);

  return (
    <article
      className="cursor-pointer border-b"
      onClick={() => router.push(`/thread/${thread.id}`)}
      style={{
        borderColor: 'var(--border)',
        background: 'var(--bg-card)',
        padding: '14px 20px',
      }}
    >
      <div className="flex gap-3">
        <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
        <div className="min-w-0 flex-1">
          <ThreadHeader
            author={author}
            timeAgo={thread.time_ago}
            parentContext={thread.parent_context}
            repostedBy={isRepostThread(thread) ? thread.author.username : null}
          />
          <ThreadBody content={thread.content} media={thread.media} />
          <ThreadActions
            comments={thread.stats.comments}
            likes={thread.stats.likes}
            reposts={thread.stats.reposts}
            bookmarks={thread.stats.bookmarks}
            isLiked={thread.is_liked === true}
            isReposted={thread.is_reposted === true}
            isBookmarked={thread.is_bookmarked === true}
            onActionClick={(event) => event.stopPropagation()}
          />
        </div>
      </div>
    </article>
  );
}
