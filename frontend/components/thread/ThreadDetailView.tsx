'use client';

import { useState, useRef, useEffect } from 'react';
import { ArrowLeft, Repeat2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Avatar } from '@/components/shared/Avatar';
import { ActionBar } from '@/components/shared/ActionBar';
import { UserHoverCard } from '@/components/shared/UserHoverCard';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useThreadActions } from '@/hooks/useThreadActions';
import { ThreadInlineEditor } from '@/components/thread/ThreadInlineEditor';
import { ThreadOptionsMenu } from '@/components/thread/ThreadOptionsMenu';
import { ThreadFoxButton } from '@/components/thread/ThreadFoxButton';
import { ReportPostModal } from '@/components/thread/ReportPostModal';
import { DeletePostModal } from '@/components/shared/DeletePostModal';
import { BookmarkCollectionPopover } from '@/components/bookmarks/BookmarkCollectionPopover';
import { ShareModal } from '@/components/share/ShareModal';
import { SharePopover } from '@/components/share/SharePopover';
import { useThreadManagement } from '@/hooks/useThreadManagement';
import type { ThreadItem, ThreadMediaResponse, ThreadMutationParams } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { useComposeStore } from '@/stores/composeStore';
import { ThreadBody } from './ThreadBody';
import { ThreadComposer } from './ThreadComposer';
import { ReplyTree } from './ReplyTree';
import { ChainThreadCard } from './ChainThreadCard';
import { ThreadCardSkeleton } from './skeletons/ThreadCardSkeleton';
import { ImageDetailModal } from './ImageDetailModal';
import { ThreadPoll } from './ThreadPoll';
import { getAuthorAvatarUrl, getRepostLabel, getThreadAuthor, isOwnThread } from './threadUtils';
import { QuotedPostBox } from './QuotedPostBox';

function formatThreadDate(dateStr: string): string {
  const date = new Date(dateStr);
  const time = date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });
  const day = date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  return `${time} · ${day}`;
}

interface ThreadDetailViewProps {
  root: ThreadItem;
  ancestors?: ThreadItem[];
  belowChain?: ThreadItem[];
  replies: ThreadItem[];
  hasNextPage?: boolean;
  fetchNextPage?: () => void;
  isFetchingNextPage?: boolean;
}

export function ThreadDetailView({ root, ancestors = [], belowChain = [], replies, hasNextPage, fetchNextPage, isFetchingNextPage }: ThreadDetailViewProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(root.id, {
    onDeleteSuccess: () => router.replace('/'),
  });
  const { like, repost, bookmark } = useThreadActions(root);
  const openReply = useComposeStore((s) => s.openReply);
  const [isEditing, setIsEditing] = useState(false);
  const [lightboxState, setLightboxState] = useState<{ media: ThreadMediaResponse[]; index: number } | null>(null);
  const [reportModalOpen, setReportModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [bookmarkAnchor, setBookmarkAnchor] = useState<DOMRect | null>(null);
  const [shareAnchor, setShareAnchor] = useState<DOMRect | null>(null);
  const [shareOpen, setShareOpen] = useState(false);
  const author = getThreadAuthor(root);
  const bookmarkTargetId = root.type?.toLowerCase() === 'repost' && typeof root.original_thread_id === 'number'
    ? root.original_thread_id
    : root.id;

  function handleBookmark(e: React.MouseEvent<HTMLButtonElement>) {
    if (bookmark.isPending) return;
    const wasBookmarked = root.is_bookmarked === true;
    bookmark.mutate();
    if (!wasBookmarked) setBookmarkAnchor(e.currentTarget.getBoundingClientRect());
  }

  const canManage = isOwnThread(root, currentUserId);
  const canReport = currentUserId != null && currentUserId !== root.author.id;
  const repostLabel = getRepostLabel(root, currentUserId);
  const repliesSentinelRef = useRef<HTMLDivElement>(null);
  const anchorRef = useRef<HTMLElement>(null);

  // When ancestors are present, land the user on the focused thread so the
  // parent chain sits above (scrollable up). Deferred via rAF so we run
  // *after* Next.js's layout-router scroll-to-top effect (parent effects
  // fire after child effects), which would otherwise undo our scroll.
  useEffect(() => {
    if (ancestors.length === 0) return;
    const rafId = requestAnimationFrame(() => {
      anchorRef.current?.scrollIntoView({ block: 'start' });
    });
    return () => cancelAnimationFrame(rafId);
  }, [root.id, ancestors.length]);


  useEffect(() => {
    const sentinel = repliesSentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage?.();
        }
      },
      { rootMargin: '300px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  async function handleSave(params: ThreadMutationParams) {
    await updateMutation.mutateAsync(params);
    setIsEditing(false);
  }

  async function handleDeleteConfirm() {
    try {
      await deleteMutation.mutateAsync();
      setDeleteModalOpen(false);
    } catch {
      // Mutation state already exposes the error.
    }
  }

  return (
    <div
      style={{
        background: 'var(--bg-card)',
        minHeight: '100vh',
      }}
    >
      <header
        className="sticky top-0 z-20 flex items-center border-b"
        style={{
          background: 'var(--bg-card)',
          borderColor: 'var(--border)',
          padding: '12px 16px',
          gap: 18,
        }}
      >
        <button
          type="button"
          aria-label="Back"
          onClick={() => router.back()}
          className="inline-flex items-center justify-center rounded-full transition-colors cursor-pointer"
          onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
          style={{
            width: 36,
            height: 36,
            background: 'transparent',
            color: 'var(--text)',
          }}
        >
          <ArrowLeft size={20} strokeWidth={2.2} />
        </button>
        <div className="text-xl font-bold" style={{ color: 'var(--text)' }}>Post</div>
      </header>

      {ancestors.length > 0 && (
        <section>
          {ancestors.map((ancestor) => (
            <ChainThreadCard key={ancestor.id} thread={ancestor} />
          ))}
        </section>
      )}

      <article
        ref={anchorRef}
        className="border-b"
        style={{
          borderColor: 'var(--border)',
          background: 'var(--bg-card)',
          padding: '16px 20px',
          scrollMarginTop: 60,
        }}
      >
        {repostLabel && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 5, marginBottom: 12 }}>
            <Repeat2 size={13} strokeWidth={2.5} color="var(--text-tertiary)" />
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)', fontWeight: 600 }}>
              {repostLabel}
            </span>
          </div>
        )}

        {/* Author row */}
        <div className="flex items-center" style={{ gap: 12 }}>
          <UserHoverCard userId={author.id} style={{ flexShrink: 0 }}>
            <Link
              href={`/profile/${author.id}`}
              className="transition-opacity hover:opacity-90"
              style={{ display: 'inline-flex', cursor: 'pointer' }}
            >
              <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
            </Link>
          </UserHoverCard>
          <div className="min-w-0 flex-1" style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <UserHoverCard userId={author.id}>
                <Link
                  href={`/profile/${author.id}`}
                  className="hover:underline"
                  style={{ fontWeight: 700, fontSize: 15, color: 'var(--text)', letterSpacing: '-0.01em', textDecoration: 'none' }}
                >
                  {author.name}
                </Link>
              </UserHoverCard>
              {author.is_premium && <VerifiedBadge size={15} />}
              {root.is_pin && (
                <span style={{ fontSize: 11, color: 'var(--text-tertiary)', marginLeft: 'auto' }}>Pinned</span>
              )}
            </div>
            <Link
              href={`/profile/${author.id}`}
              className="hover:underline"
              style={{ fontSize: 13, color: 'var(--text-tertiary)', textDecoration: 'none', width: 'fit-content' }}
            >
              @{author.username}
            </Link>
            {author.faculty && (
              <div style={{ marginTop: 2 }}>
                <span
                  style={{
                    display: 'inline-block',
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '2px 8px',
                    borderRadius: 999,
                    background: 'var(--accent-soft)',
                    color: 'var(--accent)',
                    letterSpacing: '0.01em',
                  }}
                >
                  {author.faculty.faculty_name}
                </span>
              </div>
            )}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 2, flexShrink: 0 }}>
            <ThreadFoxButton thread={root} />
            {(canManage || canReport) && (
              <ThreadOptionsMenu
                onEdit={canManage ? () => setIsEditing(true) : undefined}
                onDelete={canManage ? () => setDeleteModalOpen(true) : undefined}
                onReport={canReport ? () => setReportModalOpen(true) : undefined}
                deleting={deleteMutation.isPending}
                disabled={updateMutation.isPending}
                compact
              />
            )}
          </div>
        </div>

        {/* Content — full width, 12px below author row */}
        <div style={{ marginTop: 12 }}>
          {isEditing ? (
            <ThreadInlineEditor
              initialContent={root.content}
              initialMedia={root.media}
              onCancel={() => setIsEditing(false)}
              onSave={handleSave}
              pending={updateMutation.isPending}
            />
          ) : (
            <>
              <ThreadBody
                content={root.content}
                media={root.media}
                onImageClick={(index) => setLightboxState({ media: root.media, index })}
              />
              {root.parent_context && (
                <QuotedPostBox context={root.parent_context} />
              )}
            </>
          )}
        </div>

        <ThreadPoll
          thread={root}
          onImageClick={(media, index) => setLightboxState({ media, index })}
        />

        {root.created_at && (
          <div style={{ marginTop: 16, fontSize: 15, color: 'var(--text-tertiary)' }}>
            {formatThreadDate(root.created_at)}
          </div>
        )}

        <hr style={{ marginTop: 12, borderColor: 'var(--border)', borderTopWidth: 1 }} />

        {/* Actions */}
        <div style={{ marginTop: 12 }}>
          <ActionBar
            stats={root.stats}
            isLiked={root.is_liked === true}
            isReposted={root.is_reposted === true}
            isBookmarked={root.is_bookmarked === true}
            onLike={() => { if (!like.isPending) like.mutate(); }}
            onRepost={() => { if (!repost.isPending) repost.mutate(); }}
            onBookmark={handleBookmark}
            onReply={() => openReply(root)}
            onShare={(e) => setShareAnchor(e.currentTarget.getBoundingClientRect())}
          />
        </div>
      </article>

      <ThreadComposer
        placeholder={`Post your reply to @${root.author.username}`}
        parentId={root.id}
      />

      {belowChain.length > 0 && (
        <section>
          {belowChain.map((item, i) => (
            <ChainThreadCard
              key={item.id}
              thread={item}
              hideConnector={i === belowChain.length - 1}
              withDivider
            />
          ))}
        </section>
      )}

      <ReplyTree threads={replies} />
      <div ref={repliesSentinelRef} style={{ height: 1 }} />
      {isFetchingNextPage && (
        <div>{Array.from({ length: 2 }, (_, i) => <ThreadCardSkeleton key={i} />)}</div>
      )}

      {/* Reserve viewport-tall space below so the anchor can actually scroll
          to the top of the viewport when ancestors are present. Without this
          the document isn't tall enough and scrollIntoView is a no-op. */}
      {ancestors.length > 0 && <div aria-hidden style={{ minHeight: '100vh' }} />}


      {lightboxState !== null && (
        <ImageDetailModal
          media={lightboxState.media}
          currentIndex={lightboxState.index}
          thread={root}
          onClose={() => setLightboxState(null)}
          onNavigate={(index) => setLightboxState((current) => (
            current ? { ...current, index } : current
          ))}
        />
      )}

      {bookmarkAnchor && (
        <BookmarkCollectionPopover
          threadId={bookmarkTargetId}
          anchorRect={bookmarkAnchor}
          onClose={() => setBookmarkAnchor(null)}
        />
      )}

      {shareAnchor && (
        <SharePopover
          threadId={bookmarkTargetId}
          anchorRect={shareAnchor}
          onClose={() => setShareAnchor(null)}
          onOpenChatModal={() => { setShareAnchor(null); setShareOpen(true); }}
        />
      )}
      {shareOpen && (
        <ShareModal threadId={bookmarkTargetId} onClose={() => setShareOpen(false)} />
      )}

      {reportModalOpen && (
        <ReportPostModal
          postId={root.id}
          onClose={() => setReportModalOpen(false)}
        />
      )}
      <DeletePostModal
        open={deleteModalOpen}
        isPending={deleteMutation.isPending}
        onConfirm={() => void handleDeleteConfirm()}
        onCancel={() => setDeleteModalOpen(false)}
      />
    </div>
  );
}
