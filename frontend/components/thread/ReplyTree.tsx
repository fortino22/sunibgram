'use client';

import { useMemo, useState } from 'react';
import { Bookmark, Heart, MessageCircle, Repeat2 } from 'lucide-react';
import type { ThreadItem } from '@/lib/services/threads';
import { ActionButton } from './ActionButton';

interface ReplyTreeProps {
  threads: ThreadItem[];
  rootId: number;
}

interface ReplyItemProps {
  thread: ThreadItem;
  compact?: boolean;
}

function ReplyItem({ thread, compact = false }: ReplyItemProps) {
  const isLiked = thread.is_liked === true;
  const isReposted = thread.is_reposted === true;
  const isBookmarked = thread.is_bookmarked === true;
  const authorInitial = thread.author.username.charAt(0).toUpperCase();
  const size = compact ? 'sm' : 'md';

  return (
    <div style={{ display: 'flex', gap: 12, position: 'relative' }}>
      <div
        aria-hidden="true"
        style={{
          width: compact ? 34 : 40,
          height: compact ? 34 : 40,
          borderRadius: '50%',
          background: 'oklch(58% 0.23 270)',
          color: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
          fontWeight: 700,
          fontSize: compact ? 13 : 15,
        }}
      >
        {authorInitial}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: 6,
            color: 'var(--text-tertiary)',
            fontSize: 14,
            marginBottom: 5,
          }}
        >
          <span style={{ color: 'var(--text)', fontWeight: 700 }}>@{thread.author.username}</span>
          {thread.author.faculty && (
            <span
              style={{
                color: 'var(--accent)',
                background: 'var(--bg-subtle)',
                border: '1px solid var(--border)',
                borderRadius: 999,
                padding: '2px 7px',
                fontSize: 12,
                fontWeight: 700,
              }}
            >
              {thread.author.faculty.faculty_name}
            </span>
          )}
          <span>·</span>
          <span>{thread.time_ago}</span>
        </div>
        <div
          style={{
            color: 'var(--text)',
            fontSize: compact ? 14 : 15,
            lineHeight: 1.6,
            whiteSpace: 'pre-wrap',
            overflowWrap: 'break-word',
          }}
        >
          {thread.content}
        </div>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            gap: 8,
            marginTop: 10,
            color: 'var(--text-tertiary)',
            fontSize: 13,
            maxWidth: 320,
          }}
        >
          <ActionButton aria-label="Comment" size={size}>
            <MessageCircle size={compact ? 16 : 18} strokeWidth={2} />
          </ActionButton>
          <ActionButton aria-label="Like" size={size} color={isLiked ? '#f04452' : undefined}>
            <Heart size={compact ? 16 : 18} strokeWidth={2} fill={isLiked ? '#f04452' : 'none'} />
            <span>{thread.stats.likes}</span>
          </ActionButton>
          <ActionButton aria-label="Repost" size={size} color={isReposted ? 'var(--accent)' : undefined}>
            <Repeat2 size={compact ? 16 : 18} strokeWidth={2} />
            <span>{thread.stats.reposts}</span>
          </ActionButton>
          <ActionButton aria-label="Bookmark" size={size} color={isBookmarked ? 'var(--accent)' : undefined}>
            <Bookmark size={compact ? 16 : 18} strokeWidth={2} fill={isBookmarked ? 'var(--accent)' : 'none'} />
          </ActionButton>
        </div>
      </div>
    </div>
  );
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
      <div style={{ padding: '56px 24px', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 15 }}>
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
          <article key={thread.id} style={{ padding: 16, borderBottom: '1px solid var(--border)' }}>
            <ReplyItem thread={thread} />
            {children.length > 0 && (
              <div style={{ position: 'relative', marginTop: 14, marginLeft: 19, paddingLeft: 28 }}>
                <div
                  aria-hidden="true"
                  style={{
                    position: 'absolute',
                    left: 19,
                    top: 0,
                    bottom: 0,
                    width: 2,
                    background: 'var(--border)',
                    borderRadius: 999,
                  }}
                />
                <button
                  type="button"
                  onClick={() => setCollapsed((prev) => ({ ...prev, [thread.id]: !isCollapsed }))}
                  style={{
                    border: 0,
                    background: 'transparent',
                    color: 'var(--accent)',
                    cursor: 'pointer',
                    font: 'inherit',
                    fontSize: 14,
                    fontWeight: 700,
                    padding: '0 0 12px',
                  }}
                >
                  {isCollapsed ? `Show ${children.length} ${children.length === 1 ? 'reply' : 'replies'}` : 'Hide replies'}
                </button>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                  {visibleChildren.map((child) => (
                    <ReplyItem key={child.id} thread={child} compact />
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
