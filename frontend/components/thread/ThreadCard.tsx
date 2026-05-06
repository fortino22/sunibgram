import type { ThreadItem } from '@/lib/services/threads';

interface ThreadCardProps {
  thread: ThreadItem;
}

export function ThreadCard({ thread }: ThreadCardProps) {
  const authorInitial = thread.author.username.charAt(0).toUpperCase();

  return (
    <article
      style={{
        borderBottom: '1px solid var(--border)',
        padding: 16,
        display: 'flex',
        gap: 12,
        background: 'var(--bg-card)',
      }}
    >
      <div
        aria-hidden="true"
        style={{
          width: 40,
          height: 40,
          borderRadius: '50%',
          background: 'oklch(58% 0.23 270)',
          color: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
          fontWeight: 600,
          fontSize: 15,
        }}
      >
        {authorInitial}
      </div>

      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            color: 'var(--text-tertiary)',
            fontSize: 14,
            marginBottom: 6,
          }}
        >
          <span style={{ color: 'var(--text)', fontWeight: 600 }}>@{thread.author.username}</span>
          <span>{thread.time_ago}</span>
        </div>

        <div
          style={{
            color: 'var(--text)',
            fontSize: 15,
            lineHeight: 1.6,
            whiteSpace: 'pre-wrap',
            overflowWrap: 'break-word',
          }}
        >
          {thread.content}
        </div>

        {thread.media.length > 0 && (
          <div style={{ display: 'grid', gap: 8, marginTop: 12 }}>
            {thread.media.map((media) => (
              <img
                key={media.media_id}
                src={media.url}
                alt=""
                style={{
                  maxWidth: '100%',
                  borderRadius: 8,
                  display: 'block',
                  background: 'var(--bg-subtle)',
                }}
              />
            ))}
          </div>
        )}

        <div
          style={{
            display: 'flex',
            gap: 20,
            marginTop: 12,
            color: 'var(--text-tertiary)',
            fontSize: 13,
          }}
        >
          <span style={{ color: thread.is_liked ? 'var(--accent)' : 'var(--text-tertiary)' }}>
            {thread.stats.likes} likes
          </span>
          <span style={{ color: thread.is_reposted ? 'var(--accent)' : 'var(--text-tertiary)' }}>
            {thread.stats.reposts} reposts
          </span>
          <span style={{ color: thread.is_bookmarked ? 'var(--accent)' : 'var(--text-tertiary)' }}>
            {thread.stats.bookmarks} bookmarks
          </span>
        </div>
      </div>
    </article>
  );
}
