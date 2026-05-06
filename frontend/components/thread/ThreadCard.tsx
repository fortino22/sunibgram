'use client';

import type { CSSProperties, MouseEvent } from 'react';
import { Bookmark, Heart, MessageCircle, Repeat2 } from 'lucide-react';
import type { ThreadItem } from '@/lib/services/threads';

interface ThreadCardProps {
  thread: ThreadItem;
}

const actionButtonStyle: CSSProperties = {
  border: 0,
  borderRadius: 999,
  background: 'transparent',
  color: 'inherit',
  cursor: 'pointer',
  display: 'inline-flex',
  alignItems: 'center',
  gap: 6,
  minWidth: 44,
  height: 32,
  padding: '0 8px',
  font: 'inherit',
};

function setHoverBackground(event: MouseEvent<HTMLButtonElement>, hovered: boolean) {
  event.currentTarget.style.background = hovered ? 'var(--bg-subtle)' : 'transparent';
}

function truncate(value: string, maxLength = 140) {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}...` : value;
}

function mediaGridStyle(count: number): CSSProperties {
  return {
    display: 'grid',
    gridTemplateColumns: count === 1 ? '1fr' : 'repeat(2, minmax(0, 1fr))',
    gap: 8,
    marginTop: 12,
  };
}

function mediaImageStyle(count: number): CSSProperties {
  const maxHeight = count === 1 ? 400 : count === 2 ? 250 : 200;

  return {
    width: '100%',
    maxHeight,
    borderRadius: 8,
    display: 'block',
    background: 'var(--bg-subtle)',
    objectFit: 'cover',
  };
}

export function ThreadCard({ thread }: ThreadCardProps) {
  const isLiked = thread.is_liked === true;
  const isReposted = thread.is_reposted === true;
  const isBookmarked = thread.is_bookmarked === true;
  const isRepost = thread.type.toUpperCase() === 'REPOST' || Boolean(thread.reposted_at);
  const authorInitial = thread.author.username.charAt(0).toUpperCase();

  return (
    <article
      style={{
        borderBottom: '1px solid var(--border)',
        padding: 16,
        background: 'var(--bg-card)',
      }}
    >
      {isRepost && (
        <div
          style={{
            color: 'var(--text-tertiary)',
            fontSize: 13,
            fontWeight: 500,
            marginBottom: 8,
            paddingLeft: 52,
          }}
        >
          ↺ {thread.author.username} reposted
        </div>
      )}

      <div style={{ display: 'flex', gap: 12 }}>
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
            fontWeight: 700,
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
              gap: 6,
              color: 'var(--text-tertiary)',
              fontSize: 14,
              marginBottom: 6,
              minWidth: 0,
            }}
          >
            <span
              style={{
                color: 'var(--text)',
                fontWeight: 700,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              @{thread.author.username}
            </span>
            {thread.author.is_premium && (
              <span aria-label="Premium" title="Premium" style={{ color: '#f5c542', flexShrink: 0 }}>
                ★
              </span>
            )}
            <span style={{ flexShrink: 0 }}>·</span>
            <span style={{ flexShrink: 0 }}>{thread.time_ago}</span>
          </div>

          {thread.parent_context && (
            <div
              style={{
                border: '1px solid var(--border)',
                borderRadius: 8,
                padding: '8px 10px',
                marginBottom: 10,
                color: 'var(--text-tertiary)',
                fontSize: 13,
                minWidth: 0,
              }}
            >
              <span style={{ color: 'var(--text)', fontWeight: 700 }}>
                @{thread.parent_context.author.username}
              </span>{' '}
              <span
                style={{
                  display: 'inline',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {truncate(thread.parent_context.content)}
              </span>
            </div>
          )}

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
            <div style={mediaGridStyle(thread.media.length)}>
              {thread.media.map((media) => (
                <img key={media.media_id} src={media.url} alt="" style={mediaImageStyle(thread.media.length)} />
              ))}
            </div>
          )}

          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              gap: 12,
              marginTop: 12,
              color: 'var(--text-tertiary)',
              fontSize: 13,
              maxWidth: 360,
            }}
          >
            <button
              type="button"
              aria-label="Like"
              onClick={() => {}}
              onMouseEnter={(event) => setHoverBackground(event, true)}
              onMouseLeave={(event) => setHoverBackground(event, false)}
              style={{ ...actionButtonStyle, color: isLiked ? '#f04452' : 'var(--text-tertiary)' }}
            >
              <Heart size={18} strokeWidth={2} fill={isLiked ? '#f04452' : 'none'} />
              <span>{thread.stats.likes}</span>
            </button>

            <button
              type="button"
              aria-label="Comment"
              onClick={() => {}}
              onMouseEnter={(event) => setHoverBackground(event, true)}
              onMouseLeave={(event) => setHoverBackground(event, false)}
              style={actionButtonStyle}
            >
              <MessageCircle size={18} strokeWidth={2} />
            </button>

            <button
              type="button"
              aria-label="Repost"
              onClick={() => {}}
              onMouseEnter={(event) => setHoverBackground(event, true)}
              onMouseLeave={(event) => setHoverBackground(event, false)}
              style={{ ...actionButtonStyle, color: isReposted ? 'var(--accent)' : 'var(--text-tertiary)' }}
            >
              <Repeat2 size={18} strokeWidth={2} />
              <span>{thread.stats.reposts}</span>
            </button>

            <button
              type="button"
              aria-label="Bookmark"
              onClick={() => {}}
              onMouseEnter={(event) => setHoverBackground(event, true)}
              onMouseLeave={(event) => setHoverBackground(event, false)}
              style={{ ...actionButtonStyle, color: isBookmarked ? 'var(--accent)' : 'var(--text-tertiary)' }}
            >
              <Bookmark size={18} strokeWidth={2} fill={isBookmarked ? 'var(--accent)' : 'none'} />
              <span>{thread.stats.bookmarks}</span>
            </button>
          </div>
        </div>
      </div>
    </article>
  );
}
