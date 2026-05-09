import Link from 'next/link';
import { MoreHorizontal } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { ActionBar } from '@/components/shared/ActionBar';
import type { ThreadItem } from '@/lib/services/threads';
import { getAuthorAvatarUrl, getThreadAuthor } from '@/components/thread/threadUtils';

interface ProfileThreadCardProps {
  thread: ThreadItem;
  replyingTo?: string;
  replyingToId?: number;
  onLike?: () => void;
  onRepost?: () => void;
  onBookmark?: () => void;
}

export function ProfileThreadCard({ thread, replyingTo, replyingToId, onLike, onRepost, onBookmark }: ProfileThreadCardProps) {
  const author = getThreadAuthor(thread);

  return (
    <article
      style={{
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
        cursor: 'pointer',
      }}
    >
      {replyingTo && (
        <div
          style={{
            padding: '10px 20px 0 72px',
            fontSize: 13,
            color: 'var(--text-tertiary)',
            letterSpacing: '-0.01em',
          }}
        >
          Replying to{' '}
          {replyingToId ? (
            <Link href={`/profile/${replyingToId}`} style={{ color: 'var(--accent)', textDecoration: 'none' }}>
              @{replyingTo}
            </Link>
          ) : (
            <span style={{ color: 'var(--accent)' }}>@{replyingTo}</span>
          )}
        </div>
      )}

      <div style={{ padding: '14px 20px', display: 'flex', gap: 12 }}>
        <Avatar username={thread.author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />

        <div style={{ flex: 1, minWidth: 0 }}>
          {/* Author row */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
            <span style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em' }}>
              {thread.author.username}
            </span>
            {thread.author.is_premium && (
              <span
                style={{
                  fontSize: 10,
                  background: 'var(--accent-soft)',
                  color: 'var(--accent)',
                  padding: '1px 5px',
                  borderRadius: 4,
                  fontWeight: 600,
                  letterSpacing: '0.02em',
                }}
              >
                PRO
              </span>
            )}
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
            {thread.is_pin && (
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)', marginLeft: 'auto' }}>
                Pinned
              </span>
            )}
            <button
              type="button"
              aria-label="More options"
              style={{
                marginLeft: thread.is_pin ? 8 : 'auto',
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                color: 'var(--text-tertiary)',
                padding: 2,
                borderRadius: 4,
                display: 'flex',
                alignItems: 'center',
              }}
            >
              <MoreHorizontal size={15} strokeWidth={1.8} />
            </button>
          </div>

          {/* Content */}
          <div
            style={{
              fontSize: 15,
              color: 'var(--text)',
              lineHeight: 1.6,
              whiteSpace: 'pre-wrap',
              overflowWrap: 'break-word',
            }}
          >
            {thread.content}
          </div>

          {/* Media */}
          {thread.media.length > 0 && (
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: thread.media.length === 1 ? '1fr' : 'repeat(2, 1fr)',
                gap: 8,
                marginTop: 12,
              }}
            >
              {thread.media.map((m) => (
                <img
                  key={m.media_id}
                  src={m.url}
                  alt=""
                  style={{ width: '100%', borderRadius: 10, display: 'block', background: 'var(--bg-subtle)' }}
                />
              ))}
            </div>
          )}

          <ActionBar
            stats={thread.stats}
            isLiked={!!thread.is_liked}
            isReposted={!!thread.is_reposted}
            isBookmarked={!!thread.is_bookmarked}
            onLike={onLike}
            onRepost={onRepost}
            onBookmark={onBookmark}
            size="sm"
          />
        </div>
      </div>
    </article>
  );
}
