'use client';

import type { ThreadItem } from '@/lib/services/threads';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';

interface ReplyTreeProps {
  threads: ThreadItem[];
}

export function ReplyTree({ threads }: ReplyTreeProps) {
  if (threads.length === 0) {
    return (
      <div className="px-6 text-center text-[15px]" style={{ color: 'var(--text-tertiary)', paddingTop: 80, paddingBottom: 64 }}>
        No replies yet.
      </div>
    );
  }

  return (
    <section>
      {threads.map((thread) => (
        <ProfileThreadCard key={thread.id} thread={thread} hideParentContext />
      ))}
    </section>
  );
}
