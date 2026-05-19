'use client';

import { ThreadCard } from '@/components/thread/ThreadCard';
import type { ThreadItem } from '@/lib/services/threads';

interface CommunityThreadCardProps {
  thread: ThreadItem;
}

export function CommunityThreadCard({ thread }: CommunityThreadCardProps) {
  return (
    <div style={{ background: 'var(--bg-card)' }}>
      <ThreadCard thread={thread} />
    </div>
  );
}
