'use client';

import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import type { ThreadItem } from '@/lib/services/threads';
import type { CommunityRole } from '@/lib/services/communities';

interface ThreadCardProps {
  thread: ThreadItem;
  communityRole?: CommunityRole | null;
  onPinThread?: (threadId: number) => void;
  onUnpinThread?: (threadId: number) => void;
  pinningThread?: boolean;
  inCommunityContext?: boolean;
}

export function ThreadCard({ thread, communityRole, onPinThread, onUnpinThread, pinningThread, inCommunityContext }: ThreadCardProps) {
  return (
    <ProfileThreadCard
      thread={thread}
      communityRole={communityRole}
      onPinThread={onPinThread}
      onUnpinThread={onUnpinThread}
      pinningThread={pinningThread}
      inCommunityContext={inCommunityContext}
    />
  );
}
