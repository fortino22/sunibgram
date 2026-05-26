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
      onLike={() => { if (!like.isPending) like.mutate(); }}
      onRepost={() => { if (!repost.isPending) repost.mutate(); }}
      onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(); }}
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
