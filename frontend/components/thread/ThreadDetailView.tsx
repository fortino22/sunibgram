'use client';

import { useState } from 'react';
import { ArrowLeft, Repeat2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useThreadActions } from '@/hooks/useThreadActions';
import { ThreadInlineEditor } from '@/components/thread/ThreadInlineEditor';
import { ThreadOptionsMenu } from '@/components/thread/ThreadOptionsMenu';
import { ReportPostModal } from '@/components/thread/ReportPostModal';
import { DeletePostModal } from '@/components/shared/DeletePostModal';
import { useThreadManagement } from '@/hooks/useThreadManagement';
import type { ThreadItem, ThreadMediaResponse, ThreadMutationParams } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadComposer } from './ThreadComposer';
import { ReplyTree } from './ReplyTree';
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
  replies: ThreadItem[];
}

export function ThreadDetailView({ root, replies }: ThreadDetailViewProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(root.id, {
    onDeleteSuccess: () => router.replace('/'),
  });
  const { like, repost, bookmark } = useThreadActions(root);
  const [isEditing, setIsEditing] = useState(false);
  const [lightboxState, setLightboxState] = useState<{ media: ThreadMediaResponse[]; index: number } | null>(null);
  const [reportModalOpen, setReportModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const author = getThreadAuthor(root);
  const canManage = isOwnThread(root, currentUserId);
  const canReport = currentUserId != null && currentUserId !== root.author.id;
  const repostLabel = getRepostLabel(root, currentUserId);

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
          <div className="min-w-0 flex-1" style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <span style={{ fontWeight: 700, fontSize: 15, color: 'var(--text)', letterSpacing: '-0.01em' }}>
                {author.name}
              </span>
              {author.is_premium && <VerifiedBadge size={15} />}
              {root.is_pin && (
                <span style={{ fontSize: 11, color: 'var(--text-tertiary)', marginLeft: 'auto' }}>Pinned</span>
              )}
            </div>
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              @{author.username}
            </span>
          </div>
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

        {/* Actions */}
        <div style={{ marginTop: 12 }}>
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

      <ReplyTree threads={replies} />

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
    </main>
  );
}
