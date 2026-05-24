'use client';

import { useThreadActions } from '@/hooks/useThreadActions';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import type { ThreadItem } from '@/lib/services/threads';
import type { CommunityRole } from '@/lib/services/communities';
import { useComposeStore } from '@/stores/composeStore';

interface ThreadCardProps {
  thread: ThreadItem;
  communityRole?: CommunityRole | null;
  onPinThread?: (threadId: number) => void;
  onUnpinThread?: (threadId: number) => void;
  pinningThread?: boolean;
  inCommunityContext?: boolean;
}

export function ThreadCard({ thread, communityRole, onPinThread, onUnpinThread, pinningThread, inCommunityContext }: ThreadCardProps) {
  const { like, repost, bookmark } = useThreadActions(thread);
  const openReply = useComposeStore((s) => s.openReply);

  return (
    <ProfileThreadCard
      thread={thread}
      onLike={() => { if (!like.isPending) like.mutate(thread.is_liked === true); }}
      onRepost={() => { if (!repost.isPending) repost.mutate(thread.is_reposted === true); }}
      onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(thread.is_bookmarked === true); }}
      onReply={() => openReply(thread)}
      communityRole={communityRole}
      onPinThread={onPinThread}
      onUnpinThread={onUnpinThread}
      pinningThread={pinningThread}
      inCommunityContext={inCommunityContext}
    />
  );
}
