'use client';

import { useMemo, useState } from 'react';
import type { ThreadItem } from '@/lib/services/threads';
import { ThreadReplyItem } from './ThreadReplyItem';

interface ReplyTreeProps {
  threads: ThreadItem[];
  rootId: number;
}

export function ReplyTree({ threads, rootId }: ReplyTreeProps) {
  const tree = useMemo(
    () =>
      threads
        .filter((t) => t.parent_id === rootId)
        .map((t) => ({
          thread: t,
          children: threads.filter((c) => c.parent_id === t.id),
        })),
    [rootId, threads]
  );

  const [collapsed, setCollapsed] = useState<Record<number, boolean>>(() =>
    Object.fromEntries(tree.map(({ thread, children }) => [thread.id, children.length > 3]))
  );

  if (tree.length === 0) {
    return (
      <div className="px-6 py-14 text-center text-[15px]" style={{ color: 'var(--text-tertiary)' }}>
        No replies yet.
      </div>
    );
  }

  return (
    <section>
      {tree.map(({ thread, children }) => {
        const isCollapsed = collapsed[thread.id] === true;
        const visibleChildren = isCollapsed ? [] : children;

        return (
          <article key={thread.id} className="border-b p-4" style={{ borderColor: 'var(--border)' }}>
            <ThreadReplyItem thread={thread} />
            {children.length > 0 && (
              <div className="relative ml-[19px] mt-3.5 pl-7">
                <div
                  aria-hidden="true"
                  className="absolute bottom-0 left-[19px] top-0 w-0.5 rounded-full"
                  style={{
                    background: 'var(--border)',
                  }}
                />
                <button
                  type="button"
                  onClick={() => setCollapsed((prev) => ({ ...prev, [thread.id]: !isCollapsed }))}
                  className="pb-3 text-sm font-bold"
                  style={{
                    color: 'var(--accent)',
                  }}
                >
                  {isCollapsed ? `Show ${children.length} ${children.length === 1 ? 'reply' : 'replies'}` : 'Hide replies'}
                </button>
                <div className="flex flex-col gap-4">
                  {visibleChildren.map((child) => (
                    <ThreadReplyItem key={child.id} thread={child} compact />
                  ))}
                </div>
              </div>
            )}
          </article>
        );
      })}
    </section>
  );
}
