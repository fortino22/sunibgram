'use client';

import type { ThreadItem } from '@/lib/services/threads';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { useThreadActions } from '@/hooks/useThreadActions';

function ConnectedReply({ thread }: { thread: ThreadItem }) {
  const { like, repost, bookmark } = useThreadActions(thread);

  return (
    <ProfileThreadCard
      thread={thread}
      hideParentContext
      onLike={() => { if (!like.isPending) like.mutate(thread.is_liked === true); }}
      onRepost={() => { if (!repost.isPending) repost.mutate(thread.is_reposted === true); }}
      onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(thread.is_bookmarked === true); }}
    />
  );
}

interface ReplyTreeProps {
  threads: ThreadItem[];
}

export function ReplyTree({ threads }: ReplyTreeProps) {
  if (threads.length === 0) {
    return (
      <div className="px-6 py-14 text-center text-[15px]" style={{ color: 'var(--text-tertiary)' }}>
        No replies yet.
      </div>
    );
  }

  return (
    <section>
      {threads.map((thread) => (
        <ConnectedReply key={thread.id} thread={thread} />
      ))}
    </section>
  );
}
