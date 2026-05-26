'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Users } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useThreadActions } from '@/hooks/useThreadActions';
import { ThreadInlineEditor } from '@/components/thread/ThreadInlineEditor';
import { ThreadOptionsMenu } from '@/components/thread/ThreadOptionsMenu';
import { useThreadManagement } from '@/hooks/useThreadManagement';
import type { ThreadItem, ThreadMediaResponse, ThreadMutationParams } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadHeader } from './ThreadHeader';
import { ThreadPoll } from './ThreadPoll';
import { getAuthorAvatarUrl, getRepostLabel, getThreadAuthor, isOwnThread } from './threadUtils';
import { ImageDetailModal } from './ImageDetailModal';
import { DeletePostModal } from '@/components/shared/DeletePostModal';

interface ThreadReplyItemProps {
  thread: ThreadItem;
  compact?: boolean;
}

export function ThreadReplyItem({ thread, compact = false }: ThreadReplyItemProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(thread.id);
  const { like, repost, bookmark } = useThreadActions(thread);
  const [isEditing, setIsEditing] = useState(false);
  const [lightboxState, setLightboxState] = useState<{ media: ThreadMediaResponse[]; index: number } | null>(null);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const author = getThreadAuthor(thread);
  const canManage = isOwnThread(thread, currentUserId);
  const repostLabel = getRepostLabel(thread, currentUserId);

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
      className="cursor-pointer"
      onClick={() => router.push(`/thread/${thread.id}`)}
    >
      {thread.community && (
        <div
          className="flex items-center gap-1"
          style={{ paddingLeft: compact ? 27 : 35 }}
        >
          <Users size={13} strokeWidth={2} color="var(--text-tertiary)" />
          <Link
            href={`/communities/${thread.community.id}`}
            onClick={(e) => e.stopPropagation()}
            style={{ color: 'var(--text-tertiary)', textDecoration: 'none', fontWeight: 700, fontSize: 14, letterSpacing: '-0.01em' }}
          >
            {thread.community.name}
          </Link>
        </div>
      )}
      <div className="flex gap-3">
      <Avatar
        username={author.username}
        avatarUrl={getAuthorAvatarUrl(author)}
        size={compact ? 'sm' : 'md'}
      />
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <ThreadHeader
              author={author}
              timeAgo={thread.time_ago}
              compact={compact}
              parentContext={thread.parent_context}
              repostLabel={repostLabel}
            />
          </div>
          {canManage && (
            <ThreadOptionsMenu
              onEdit={() => setIsEditing(true)}
              onDelete={() => setDeleteModalOpen(true)}
              deleting={deleteMutation.isPending}
              disabled={updateMutation.isPending}
              compact={compact}
            />
          )}
        </div>

        {isEditing ? (
          <ThreadInlineEditor
            initialContent={thread.content}
            initialMedia={thread.media}
            onCancel={() => setIsEditing(false)}
            onSave={handleSave}
            pending={updateMutation.isPending}
            errorMessage={updateMutation.error?.message}
            compact={compact}
          />
        ) : (
          <ThreadBody
            content={thread.content}
            media={thread.media}
            compact={compact}
            onImageClick={(index) => setLightboxState({ media: thread.media, index })}
          />
        )}

        <ThreadPoll
          thread={thread}
          compact={compact}
          onImageClick={(media, index) => setLightboxState({ media, index })}
        />

        {deleteMutation.isError && (
          <div
            className="mt-3 rounded-xl px-3 py-2 text-[13px]"
            style={{
              background: 'var(--coral-soft)',
              color: 'var(--coral)',
            }}
          >
            {deleteMutation.error.message}
          </div>
        )}

        <div className="flex items-center justify-between" style={{ marginTop: 12, gap: 12 }}>
          <ThreadActions
            comments={thread.stats.comments}
            likes={thread.stats.likes}
            reposts={thread.stats.reposts}
            bookmarks={thread.stats.bookmarks}
            isLiked={thread.is_liked === true}
            isReposted={thread.is_reposted === true}
            isBookmarked={thread.is_bookmarked === true}
            compact={compact}
            topMargin={0}
            onLike={() => {
              if (!like.isPending) {
                like.mutate();
              }
            }}
            onRepost={() => {
              if (!repost.isPending) {
                repost.mutate();
              }
            }}
            onBookmark={() => {
              if (!bookmark.isPending) {
                bookmark.mutate();
              }
            }}
            onActionClick={(event) => event.stopPropagation()}
          />
        </div>
      </div>

      {lightboxState !== null && (
        <ImageDetailModal
          media={lightboxState.media}
          currentIndex={lightboxState.index}
          thread={thread}
          onClose={() => setLightboxState(null)}
          onNavigate={(index) => setLightboxState((current) => (
            current ? { ...current, index } : current
          ))}
        />
      )}
      </div>
      <DeletePostModal
        open={deleteModalOpen}
        isPending={deleteMutation.isPending}
        onConfirm={() => void handleDeleteConfirm()}
        onCancel={() => setDeleteModalOpen(false)}
      />
    </div>
  );
}
