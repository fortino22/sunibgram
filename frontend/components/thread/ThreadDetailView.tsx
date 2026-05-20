'use client';

import { useState } from 'react';
import { ArrowLeft, Repeat2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { useThreadActions } from '@/hooks/useThreadActions';
import { ThreadInlineEditor } from '@/components/thread/ThreadInlineEditor';
import { ThreadOptionsMenu } from '@/components/thread/ThreadOptionsMenu';
import { useThreadManagement } from '@/hooks/useThreadManagement';
import type { ThreadItem, ThreadMutationParams } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadComposer } from './ThreadComposer';
import { ReplyTree } from './ReplyTree';
import { ImageDetailModal } from './ImageDetailModal';
import { getAuthorAvatarUrl, getRepostLabel, getThreadAuthor, isOwnThread } from './threadUtils';

interface ThreadDetailViewProps {
  root: ThreadItem;
  replies: ThreadItem[];
  viewFullThreadId?: number | null;
}

export function ThreadDetailView({ root, replies, viewFullThreadId = null }: ThreadDetailViewProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(root.id, {
    onDeleteSuccess: () => router.replace('/'),
  });
  const { like, repost, bookmark } = useThreadActions(root);
  const [isEditing, setIsEditing] = useState(false);
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);
  const author = getThreadAuthor(root);
  const canManage = isOwnThread(root, currentUserId);
  const repostLabel = getRepostLabel(root, currentUserId);

  async function handleSave(params: ThreadMutationParams) {
    await updateMutation.mutateAsync(params);
    setIsEditing(false);
  }

  async function handleDelete() {
    if (!window.confirm('Delete this post?')) return;
    try {
      await deleteMutation.mutateAsync();
    } catch {
      // Mutation state already exposes the error.
    }
  }

  return (
    <main
      className="mx-auto min-h-screen max-w-[600px]"
      style={{
        background: 'var(--bg-card)',
        borderLeft: '1px solid var(--border)',
        borderRight: '1px solid var(--border)',
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
          className="inline-flex items-center justify-center rounded-full"
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

      <article
        className="border-b"
        style={{
          borderColor: 'var(--border)',
          background: 'var(--bg-card)',
          padding: '16px 20px',
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
          <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
          <div className="min-w-0 flex-1 flex items-center" style={{ gap: 8 }}>
            <span style={{ fontWeight: 700, fontSize: 15, color: 'var(--text)', letterSpacing: '-0.01em' }}>
              {author.username}
            </span>
            {author.is_premium && (
              <span style={{ fontSize: 10, background: 'var(--accent-soft)', color: 'var(--accent)', padding: '1px 5px', borderRadius: 4, fontWeight: 600, letterSpacing: '0.02em' }}>
                PRO
              </span>
            )}
            {root.is_pin && (
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)', marginLeft: 'auto' }}>Pinned</span>
            )}
          </div>
          {canManage && (
            <ThreadOptionsMenu
              onEdit={() => setIsEditing(true)}
              onDelete={() => void handleDelete()}
              deleting={deleteMutation.isPending}
              disabled={updateMutation.isPending}
              compact
            />
          )}
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
              errorMessage={updateMutation.error?.message}
            />
          ) : (
            <ThreadBody
              content={root.content}
              media={root.media}
              onImageClick={(i) => setLightboxIndex(i)}
            />
          )}
        </div>

        {root.time_ago && (
          <div style={{ marginTop: 16, fontSize: 13, color: 'var(--text-tertiary)' }}>
            {root.time_ago}
          </div>
        )}

        {deleteMutation.isError && (
          <div
            className="rounded-xl text-[13px]"
            style={{
              marginTop: 12,
              padding: '8px 12px',
              background: 'var(--coral-soft)',
              color: 'var(--coral)',
            }}
          >
            {deleteMutation.error.message}
          </div>
        )}

        {/* Stats row — border top and bottom */}
        <div
          className="border-t border-b flex flex-wrap text-sm"
          style={{
            marginTop: 12,
            padding: '12px 0',
            rowGap: 8,
            columnGap: 20,
            color: 'var(--text-tertiary)',
            borderColor: 'var(--border)',
          }}
        >
          <span>
            <strong style={{ color: 'var(--text)' }}>{root.stats.likes}</strong> Likes
          </span>
          <span>
            <strong style={{ color: 'var(--text)' }}>{root.stats.reposts}</strong> Reposts
          </span>
          <span>
            <strong style={{ color: 'var(--text)' }}>{root.stats.bookmarks}</strong> Bookmarks
          </span>
        </div>

        {/* Actions — 4px below stats border */}
        <div style={{ marginTop: 4 }}>
          <ThreadActions
            comments={root.stats.comments}
            likes={root.stats.likes}
            reposts={root.stats.reposts}
            bookmarks={root.stats.bookmarks}
            isLiked={root.is_liked === true}
            isReposted={root.is_reposted === true}
            isBookmarked={root.is_bookmarked === true}
            onLike={() => { if (!like.isPending) like.mutate(root.is_liked === true); }}
            onRepost={() => { if (!repost.isPending) repost.mutate(root.is_reposted === true); }}
            onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(root.is_bookmarked === true); }}
          />
        </div>
      </article>

      <ThreadComposer
        placeholder={`Post your reply to @${root.author.username}`}
        parentId={root.id}
      />

      {viewFullThreadId !== null && (
        <section
          className="border-b"
          style={{
            borderColor: 'var(--border)',
            background: 'var(--bg-card)',
            padding: '14px 20px',
          }}
        >
          <button
            type="button"
            onClick={() => router.push(`/thread/${viewFullThreadId}`)}
            className="rounded-full text-sm font-semibold"
            style={{
              padding: '8px 16px',
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
            }}
          >
            View full thread
          </button>
        </section>
      )}

      <ReplyTree threads={replies} />

      {lightboxIndex !== null && (
        <ImageDetailModal
          media={root.media}
          currentIndex={lightboxIndex}
          thread={root}
          onClose={() => setLightboxIndex(null)}
          onNavigate={setLightboxIndex}
        />
      )}
    </main>
  );
}
