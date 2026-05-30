'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { ActionBar } from '@/components/shared/ActionBar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { UserHoverCard } from '@/components/shared/UserHoverCard';
import { BookmarkCollectionPopover } from '@/components/bookmarks/BookmarkCollectionPopover';
import { useThreadActions } from '@/hooks/useThreadActions';
import { useComposeStore } from '@/stores/composeStore';
import { getAuthorAvatarUrl, getThreadAuthor } from './threadUtils';
import { ThreadBody } from './ThreadBody';
import { ThreadPoll } from './ThreadPoll';
import { ImageDetailModal } from './ImageDetailModal';
import type { ThreadItem, ThreadMediaResponse } from '@/lib/services/threads';

interface ChainThreadCardProps {
  thread: ThreadItem;
}

// Layout constants — the absolute-positioned connector line lands on the
// next avatar by bleeding past the article's bottom edge. Padding-Y matches
// the anchor article (16) so the bleed lines up with both chain and anchor
// avatars without crossing them.
const ARTICLE_PADDING_X = 20;
const ARTICLE_PADDING_Y = 16;
const AVATAR_SIZE = 40;
const AVATAR_GAP_BELOW = 4;
const LINE_WIDTH = 2;
const LINE_BLEED = 16;

export function ChainThreadCard({ thread }: ChainThreadCardProps) {
  const router = useRouter();
  const { like, repost, bookmark } = useThreadActions(thread);
  const openReply = useComposeStore((s) => s.openReply);
  const author = getThreadAuthor(thread);
  const [hovered, setHovered] = useState(false);
  const [lightboxState, setLightboxState] = useState<{ media: ThreadMediaResponse[]; index: number } | null>(null);
  const [bookmarkAnchor, setBookmarkAnchor] = useState<DOMRect | null>(null);

  const bookmarkTargetId = thread.type?.toLowerCase() === 'repost' && typeof thread.original_thread_id === 'number'
    ? thread.original_thread_id
    : thread.id;

  function handleBookmark(e: React.MouseEvent<HTMLButtonElement>) {
    if (bookmark.isPending) return;
    const wasBookmarked = thread.is_bookmarked === true;
    bookmark.mutate();
    if (!wasBookmarked) setBookmarkAnchor(e.currentTarget.getBoundingClientRect());
  }

  return (
    <>
      <article
        onClick={() => router.push(`/thread/${thread.id}`)}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        style={{
          position: 'relative',
          background: hovered ? 'var(--bg-hover)' : 'var(--bg-card)',
          cursor: 'pointer',
          transition: 'background 0.12s ease',
        }}
      >
        {/* Connector line from below this avatar to the next item's avatar.
            zIndex:1 needed so the bleed past the bottom edge paints over the
            next card's background (which would otherwise hide the line). */}
        <div
          aria-hidden
          style={{
            position: 'absolute',
            top: ARTICLE_PADDING_Y + AVATAR_SIZE + AVATAR_GAP_BELOW,
            bottom: -LINE_BLEED,
            left: ARTICLE_PADDING_X + AVATAR_SIZE / 2 - LINE_WIDTH / 2,
            width: LINE_WIDTH,
            background: 'var(--border)',
            pointerEvents: 'none',
            zIndex: 1,
          }}
        />

        <div style={{ display: 'flex', gap: 12, padding: `${ARTICLE_PADDING_Y}px ${ARTICLE_PADDING_X}px` }}>
          <div style={{ flexShrink: 0 }}>
            <UserHoverCard userId={thread.author.id}>
              <Link
                href={`/profile/${thread.author.id}`}
                onClick={(e) => e.stopPropagation()}
                style={{ display: 'inline-flex' }}
              >
                <Avatar username={thread.author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
              </Link>
            </UserHoverCard>
          </div>

          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
              <UserHoverCard userId={thread.author.id}>
                <Link
                  href={`/profile/${thread.author.id}`}
                  onClick={(e) => e.stopPropagation()}
                  style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em', textDecoration: 'none' }}
                >
                  {thread.author.name || thread.author.username}
                </Link>
              </UserHoverCard>
              {thread.author.is_premium && <VerifiedBadge size={15} />}
              <Link
                href={`/profile/${thread.author.id}`}
                onClick={(e) => e.stopPropagation()}
                style={{ fontSize: 13, color: 'var(--text-tertiary)', textDecoration: 'none' }}
              >
                @{thread.author.username}
              </Link>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
            </div>

            <ThreadBody
              content={thread.content}
              media={thread.media}
              onImageClick={(index) => setLightboxState({ media: thread.media, index })}
            />

            <ThreadPoll
              thread={thread}
              compact
              onImageClick={(media, index) => setLightboxState({ media, index })}
            />

            <div onClick={(e) => e.stopPropagation()}>
              <ActionBar
                stats={thread.stats}
                isLiked={!!thread.is_liked}
                isReposted={!!thread.is_reposted}
                isBookmarked={!!thread.is_bookmarked}
                onLike={() => { if (!like.isPending) like.mutate(); }}
                onRepost={() => { if (!repost.isPending) repost.mutate(); }}
                onBookmark={handleBookmark}
                onReply={() => openReply(thread)}
                size="sm"
              />
            </div>
          </div>
        </div>
      </article>

      {lightboxState !== null && (
        <div onClick={(e) => e.stopPropagation()}>
          <ImageDetailModal
            media={lightboxState.media}
            currentIndex={lightboxState.index}
            thread={thread}
            onClose={() => setLightboxState(null)}
            onNavigate={(index) => setLightboxState((current) => (
              current ? { ...current, index } : current
            ))}
          />
        </div>
      )}

      {bookmarkAnchor && (
        <div onClick={(e) => e.stopPropagation()}>
          <BookmarkCollectionPopover
            threadId={bookmarkTargetId}
            anchorRect={bookmarkAnchor}
            onClose={() => setBookmarkAnchor(null)}
          />
        </div>
      )}
    </>
  );
}
