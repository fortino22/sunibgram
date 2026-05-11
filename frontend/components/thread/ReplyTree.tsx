import type { ThreadItem } from '@/lib/services/threads';
import { ThreadReplyItem } from './ThreadReplyItem';

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
        <article
          key={thread.id}
          className="border-b"
          style={{
            borderColor: 'var(--border)',
            background: 'var(--bg-card)',
            padding: '14px 20px',
          }}
        >
          <ThreadReplyItem thread={thread} />
        </article>
      ))}
    </section>
  );
}
