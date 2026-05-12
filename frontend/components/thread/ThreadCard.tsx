'use client';

import { useState } from 'react';
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
import { ThreadHeader } from './ThreadHeader';
import { getAuthorAvatarUrl, getRepostLabel, getThreadAuthor, isOwnThread } from './threadUtils';

interface ThreadCardProps {
  thread: ThreadItem;
}

export function ThreadCard({ thread }: ThreadCardProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(thread.id);
  const { like, repost, bookmark } = useThreadActions(thread);
  const [isEditing, setIsEditing] = useState(false);
  const author = getThreadAuthor(thread);
  const canManage = isOwnThread(thread, currentUserId);
  const repostLabel = getRepostLabel(thread, currentUserId);

  async function handleSave(params: ThreadMutationParams) {
    await updateMutation.mutateAsync(params);
    setIsEditing(false);
  }

  async function handleDelete() {
    if (!window.confirm('Delete this post?')) {
      return;
    }

    try {
      await deleteMutation.mutateAsync();
    } catch {
      // Mutation state already exposes the error.
    }
  }

  return (
    <article
      className="cursor-pointer border-b"
      onClick={() => router.push(`/thread/${thread.id}`)}
      style={{
        borderColor: 'var(--border)',
        background: 'var(--bg-card)',
        padding: '14px 20px',
      }}
    >
      <div className="flex gap-3">
        <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0 flex-1">
              <ThreadHeader
                author={author}
                timeAgo={thread.time_ago}
                parentContext={thread.parent_context}
                repostLabel={repostLabel}
              />
            </div>
            {canManage && (
              <ThreadOptionsMenu
                onEdit={() => setIsEditing(true)}
                onDelete={() => void handleDelete()}
                deleting={deleteMutation.isPending}
                disabled={updateMutation.isPending}
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
            />
          ) : (
            <ThreadBody content={thread.content} media={thread.media} />
          )}

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

          <ThreadActions
            comments={thread.stats.comments}
            likes={thread.stats.likes}
            reposts={thread.stats.reposts}
            bookmarks={thread.stats.bookmarks}
            isLiked={thread.is_liked === true}
            isReposted={thread.is_reposted === true}
            isBookmarked={thread.is_bookmarked === true}
            onLike={() => {
              if (!like.isPending) {
                like.mutate(thread.is_liked === true);
              }
            }}
            onRepost={() => {
              if (!repost.isPending) {
                repost.mutate(thread.is_reposted === true);
              }
            }}
            onBookmark={() => {
              if (!bookmark.isPending) {
                bookmark.mutate(thread.is_bookmarked === true);
              }
            }}
            onActionClick={(event) => event.stopPropagation()}
          />
        </div>
      </div>
    </article>
  );
}
