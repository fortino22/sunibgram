'use client';

import { useThreadActions } from '@/hooks/useThreadActions';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import type { ThreadItem } from '@/lib/services/threads';

interface ThreadCardProps {
  thread: ThreadItem;
}

export function ThreadCard({ thread }: ThreadCardProps) {
  const { like, repost, bookmark } = useThreadActions(thread);

  return (
    <ProfileThreadCard
      thread={thread}
      onLike={() => { if (!like.isPending) like.mutate(thread.is_liked === true); }}
      onRepost={() => { if (!repost.isPending) repost.mutate(thread.is_reposted === true); }}
      onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(thread.is_bookmarked === true); }}
    />
  );
}
