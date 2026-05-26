'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';
import type { ThreadItem, ThreadMediaResponse } from '@/lib/services/threads';
import { Avatar } from '@/components/shared/Avatar';
import { ActionBar } from '@/components/shared/ActionBar';
import { PostComposer } from '@/components/shared/PostComposer';
import { useThreadChain } from '@/hooks/useThread';
import { useThreadActions } from '@/hooks/useThreadActions';
import { getAuthorAvatarUrl, getThreadAuthor } from './threadUtils';

/* ------------------------------------------------------------------ */
/* Minimal reply card used only inside this modal (avoids circular dep) */
/* ------------------------------------------------------------------ */
function ModalReply({ thread }: { thread: ThreadItem }) {
  const router = useRouter();
  const { like, repost, bookmark } = useThreadActions(thread);
  const author = getThreadAuthor(thread);

  return (
    <div
      style={{ borderBottom: '1px solid var(--border)', padding: '14px 20px', cursor: 'pointer' }}
      onClick={() => router.push(`/thread/${thread.id}`)}
    >
      <div style={{ display: 'flex', gap: 10 }}>
        <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="sm" />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
            <span style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em' }}>
              {author.username}
            </span>
            {thread.time_ago && (
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
            )}
          </div>

          {thread.content.trim().length > 0 && (
            <div style={{ fontSize: 14, color: 'var(--text)', lineHeight: 1.6, whiteSpace: 'pre-wrap', overflowWrap: 'break-word', marginBottom: 8 }}>
              {thread.content}
            </div>
          )}

          {thread.media.length > 0 && (
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: thread.media.length === 1 ? '1fr' : 'repeat(2, 1fr)',
                gap: 6,
                marginBottom: 8,
              }}
            >
              {thread.media.map((m) => (
                <img
                  key={m.media_id}
                  src={m.url}
                  alt=""
                  style={{
                    width: '100%',
                    height: thread.media.length === 1 ? 'auto' : 110,
                    maxHeight: thread.media.length === 1 ? 180 : 110,
                    objectFit: 'cover',
                    borderRadius: 10,
                    display: 'block',
                  }}
                />
              ))}
            </div>
          )}

          <div onClick={(e) => e.stopPropagation()}>
            <ActionBar
              stats={thread.stats}
              isLiked={!!thread.is_liked}
              isReposted={!!thread.is_reposted}
              isBookmarked={!!thread.is_bookmarked}
              onLike={() => { if (!like.isPending) like.mutate(thread.is_liked === true); }}
              onRepost={() => { if (!repost.isPending) repost.mutate(thread.is_reposted === true); }}
              onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(thread.is_bookmarked === true); }}
              size="sm"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/* Main modal                                                          */
/* ------------------------------------------------------------------ */
export interface ImageDetailModalProps {
  media: ThreadMediaResponse[];
  currentIndex: number;
  thread: ThreadItem;
  onClose: () => void;
  onNavigate: (index: number) => void;
}

export function ImageDetailModal({ media, currentIndex, thread, onClose, onNavigate }: ImageDetailModalProps) {
  const [showPanel, setShowPanel] = useState(false);

  const { replies, isLoading: repliesLoading } = useThreadChain(thread.id);
  const { like, repost, bookmark } = useThreadActions(thread);
  const author = getThreadAuthor(thread);

  const current = media[currentIndex];
  const hasPrev = currentIndex > 0;
  const hasNext = currentIndex < media.length - 1;

  useEffect(() => {
    setShowPanel(window.innerWidth >= 768);
    function handleResize() { setShowPanel(window.innerWidth >= 768); }
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = prev; };
  }, []);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowLeft' && hasPrev) onNavigate(currentIndex - 1);
      if (e.key === 'ArrowRight' && hasNext) onNavigate(currentIndex + 1);
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [currentIndex, hasPrev, hasNext, onClose, onNavigate]);

  if (!current) return null;

  const iconBtn = (extra?: React.CSSProperties): React.CSSProperties => ({
    width: 36,
    height: 36,
    borderRadius: '50%',
    background: 'rgba(255,255,255,0.14)',
    border: 'none',
    color: '#fff',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    ...extra,
  });

  return (
    <div style={{ position: 'fixed', inset: 0, zIndex: 1000, display: 'flex', background: '#000' }}>

      {/* ── Left: image panel ─────────────────────────────────────── */}
      <div style={{ flex: 1, position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden' }}>

        {/* Close */}
        <button type="button" aria-label="Close" onClick={onClose} style={{ ...iconBtn(), position: 'absolute', top: 16, left: 16, zIndex: 2 }}>
          <X size={18} />
        </button>

        {/* Prev */}
        {hasPrev && (
          <button type="button" aria-label="Previous" onClick={() => onNavigate(currentIndex - 1)}
            style={{ ...iconBtn({ width: 44, height: 44 }), position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', zIndex: 2 }}>
            <ChevronLeft size={24} />
          </button>
        )}

        <img
          src={current.url}
          alt=""
          style={{ maxWidth: '100%', maxHeight: '100vh', objectFit: 'contain', display: 'block', userSelect: 'none' }}
        />

        {/* Next */}
        {hasNext && (
          <button type="button" aria-label="Next" onClick={() => onNavigate(currentIndex + 1)}
            style={{ ...iconBtn({ width: 44, height: 44 }), position: 'absolute', right: 16, top: '50%', transform: 'translateY(-50%)', zIndex: 2 }}>
            <ChevronRight size={24} />
          </button>
        )}

        {/* Dot indicators */}
        {media.length > 1 && (
          <div style={{ position: 'absolute', bottom: 24, left: '50%', transform: 'translateX(-50%)', display: 'flex', gap: 6 }}>
            {media.map((_, i) => (
              <button
                key={i}
                type="button"
                aria-label={`Image ${i + 1}`}
                onClick={() => onNavigate(i)}
                style={{
                  width: i === currentIndex ? 20 : 8,
                  height: 8,
                  borderRadius: 4,
                  background: i === currentIndex ? '#fff' : 'rgba(255,255,255,0.35)',
                  border: 'none',
                  cursor: 'pointer',
                  padding: 0,
                  transition: 'width 0.15s, background 0.15s',
                }}
              />
            ))}
          </div>
        )}
      </div>

      {/* ── Right: thread panel ───────────────────────────────────── */}
      {showPanel && (
        <div style={{
          width: 420,
          borderLeft: '1px solid var(--border)',
          background: 'var(--bg-card)',
          display: 'flex',
          flexDirection: 'column',
          maxHeight: '100vh',
        }}>
          <div style={{ flex: 1, overflowY: 'auto' }}>

            {/* Author */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '20px 20px 14px' }}>
              <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
              <div>
                <div style={{ fontWeight: 700, fontSize: 15, color: 'var(--text)', letterSpacing: '-0.01em' }}>
                  {author.username}
                </div>
                {thread.time_ago && (
                  <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>{thread.time_ago}</div>
                )}
              </div>
            </div>

            {/* Caption */}
            {thread.content.trim().length > 0 && (
              <div style={{ padding: '0 20px 16px', fontSize: 15, color: 'var(--text)', lineHeight: 1.65, whiteSpace: 'pre-wrap', overflowWrap: 'break-word' }}>
                {thread.content}
              </div>
            )}

            {/* Stats */}
            <div style={{
              display: 'flex',
              flexWrap: 'wrap',
              columnGap: 16,
              rowGap: 6,
              fontSize: 14,
              color: 'var(--text-tertiary)',
              borderTop: '1px solid var(--border)',
              borderBottom: '1px solid var(--border)',
              padding: '12px 20px',
            }}>
              <span><strong style={{ color: 'var(--text)' }}>{thread.stats.likes}</strong> Likes</span>
              <span><strong style={{ color: 'var(--text)' }}>{thread.stats.reposts}</strong> Reposts</span>
              <span><strong style={{ color: 'var(--text)' }}>{thread.stats.bookmarks}</strong> Bookmarks</span>
            </div>

            {/* Actions */}
            <div style={{ padding: '4px 20px', borderBottom: '1px solid var(--border)' }}>
              <ActionBar
                stats={thread.stats}
                isLiked={!!thread.is_liked}
                isReposted={!!thread.is_reposted}
                isBookmarked={!!thread.is_bookmarked}
                onLike={() => { if (!like.isPending) like.mutate(thread.is_liked === true); }}
                onRepost={() => { if (!repost.isPending) repost.mutate(thread.is_reposted === true); }}
                onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(thread.is_bookmarked === true); }}
              />
            </div>

            {/* Reply composer */}
            <PostComposer placeholder={`Reply to @${author.username}`} parentId={thread.id} />

            {/* Replies */}
            {repliesLoading ? (
              <div style={{ padding: '32px 20px', textAlign: 'center', fontSize: 14, color: 'var(--text-tertiary)' }}>
                Loading replies…
              </div>
            ) : replies.length === 0 ? (
              <div style={{ padding: '32px 20px', textAlign: 'center', fontSize: 14, color: 'var(--text-tertiary)' }}>
                No replies yet.
              </div>
            ) : (
              replies.map((reply) => <ModalReply key={reply.id} thread={reply} />)
            )}

          </div>
        </div>
      )}
    </div>
  );
}
