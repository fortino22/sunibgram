'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Users, Repeat2, Pin } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { ActionBar } from '@/components/shared/ActionBar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { ThreadInlineEditor } from '@/components/thread/ThreadInlineEditor';
import { ThreadOptionsMenu } from '@/components/thread/ThreadOptionsMenu';
import { ReportPostModal } from '@/components/thread/ReportPostModal';
import { DeletePostModal } from '@/components/shared/DeletePostModal';
import { useThreadManagement } from '@/hooks/useThreadManagement';
import type { ThreadItem, ThreadMediaResponse, ThreadMutationParams } from '@/lib/services/threads';
import type { CommunityRole } from '@/lib/services/communities';
import { useAuthStore } from '@/stores/authStore';
import { getAuthorAvatarUrl, getThreadAuthor, isOwnThread } from '@/components/thread/threadUtils';
import { ThreadBody } from '@/components/thread/ThreadBody';
import { ImageDetailModal } from '@/components/thread/ImageDetailModal';
import { QuotedPostBox } from '@/components/thread/QuotedPostBox';
import { ThreadPoll } from '@/components/thread/ThreadPoll';

interface ProfileThreadCardProps {
  thread: ThreadItem;
  replyingTo?: string;
  replyingToId?: number;
  hideParentContext?: boolean;
  onLike?: () => void;
  onRepost?: () => void;
  onBookmark?: () => void;
  onReply?: () => void;
  communityRole?: CommunityRole | null;
  onPinThread?: (threadId: number) => void;
  onUnpinThread?: (threadId: number) => void;
  pinningThread?: boolean;
  inCommunityContext?: boolean;
}

export function ProfileThreadCard({ thread, replyingTo, replyingToId, hideParentContext = false, onLike, onRepost, onBookmark, onReply, communityRole, onPinThread, onUnpinThread, pinningThread, inCommunityContext = false }: ProfileThreadCardProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(thread.id);
  const [isEditing, setIsEditing] = useState(false);
  const [hovered, setHovered] = useState(false);
  const [lightboxState, setLightboxState] = useState<{ media: ThreadMediaResponse[]; index: number } | null>(null);
  const [reportModalOpen, setReportModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const author = getThreadAuthor(thread);
  const canManage = isOwnThread(thread, currentUserId);
  const canReport = currentUserId != null && currentUserId !== thread.author.id;
  const replyUsername = replyingTo ?? thread.parent_context?.author.username;
  const replyUserId = replyingToId ?? thread.parent_context?.author.id;

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
    <>
    <article
      onClick={() => router.push(`/thread/${thread.id}`)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        borderBottom: '1px solid var(--border)',
        background: hovered ? 'var(--bg-subtle)' : 'var(--bg-card)',
        cursor: 'pointer',
        transition: 'background 0.12s ease',
      }}
    >
      {thread.reposted_by_username && (
        <div
          style={{
            padding: '8px 20px 0 52px',
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          <Repeat2 size={13} strokeWidth={2.5} color="var(--text-tertiary)" />
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)', fontWeight: 600 }}>
            {thread.reposted_by_name || thread.reposted_by_username} reposting
          </span>
        </div>
      )}

      {inCommunityContext && thread.pinned_by_admin && (
        <div
          style={{
            padding: '8px 20px 0 52px',
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          <Pin size={13} strokeWidth={2.5} color="var(--text-tertiary)" />
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)', fontWeight: 600 }}>
            Pinned by Community admin
          </span>
        </div>
      )}

      {thread.community && !thread.reposted_by_username && !(inCommunityContext && thread.pinned_by_admin) && (
        <div
          style={{
            padding: `${thread.reposted_by_username ? '4px' : '10px'} 20px 0 55px`,
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
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

      <div style={{ padding: (thread.reposted_by_username || thread.community) ? '2px 20px 8px' : '14px 20px 8px', display: 'flex', gap: 12 }}>
        <Link href={`/profile/${thread.author.id}`} onClick={(e) => e.stopPropagation()} style={{ flexShrink: 0, alignSelf: 'flex-start' }}>
          <Avatar username={thread.author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
        </Link>

        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ marginBottom: replyUsername ? 2 : 4 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Link
                href={`/profile/${thread.author.id}`}
                onClick={(e) => e.stopPropagation()}
                style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em', textDecoration: 'none' }}
              >
                {thread.author.name || thread.author.username}
              </Link>
              {thread.author.is_premium && <VerifiedBadge size={15} />}
              <Link
                href={`/profile/${thread.author.id}`}
                onClick={(e) => e.stopPropagation()}
                style={{ fontSize: 13, color: 'var(--text-tertiary)', textDecoration: 'none' }}
              >
                @{thread.author.username}
              </Link>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
              {(canManage || canReport) && (
                <div style={{ marginLeft: 'auto' }} onClick={(e) => e.stopPropagation()}>
                  <ThreadOptionsMenu
                    onEdit={canManage ? () => setIsEditing(true) : undefined}
                    onDelete={canManage ? () => setDeleteModalOpen(true) : undefined}
                    deleting={deleteMutation.isPending}
                    disabled={updateMutation.isPending}
                    compact
                    communityRole={communityRole}
                    isPinned={thread.pinned_by_admin === true}
                    pinning={pinningThread}
                    onPin={onPinThread ? () => onPinThread(thread.id) : undefined}
                    onUnpin={onUnpinThread ? () => onUnpinThread(thread.id) : undefined}
                    onReport={canReport ? () => setReportModalOpen(true) : undefined}
                  />
                </div>
              )}
            </div>
            {thread.author.faculty && (
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
                  {thread.author.faculty.faculty_name}
                </span>
              </div>
            )}
          </div>
          
          {replyUsername && (
            <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 4, letterSpacing: '-0.01em' }}>
              Replying to{' '}
              {replyUserId ? (
                <Link
                  href={`/profile/${replyUserId}`}
                  onClick={(e) => e.stopPropagation()}
                  style={{ color: 'var(--accent)', textDecoration: 'none' }}
                >
                  @{replyUsername}
                </Link>
              ) : (
                <span style={{ color: 'var(--accent)' }}>@{replyUsername}</span>
              )}
            </div>
          )}

          {isEditing ? (
            <div onClick={(e) => e.stopPropagation()}>
              <ThreadInlineEditor
                initialContent={thread.content}
                initialMedia={thread.media}
                onCancel={() => setIsEditing(false)}
                onSave={handleSave}
                pending={updateMutation.isPending}
                errorMessage={updateMutation.error?.message}
                compact
              />
            </div>
          ) : (
            <>
              <ThreadBody
                content={thread.content}
                media={thread.media}
                onImageClick={(index) => setLightboxState({ media: thread.media, index })}
              />
              {!hideParentContext && thread.parent_context && (
                <div onClick={(e) => e.stopPropagation()}>
                  <QuotedPostBox context={thread.parent_context} />
                </div>
              )}
            </>
          )}

          <ThreadPoll
            thread={thread}
            compact
            onImageClick={(media, index) => setLightboxState({ media, index })}
          />

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

          {deleteMutation.isError && (
            <div
              style={{
                marginTop: 12,
                padding: '10px 12px',
                borderRadius: 10,
                background: 'var(--coral-soft)',
                color: 'var(--coral)',
                fontSize: 13,
                lineHeight: 1.5,
              }}
            >
              {deleteMutation.error.message}
            </div>
          )}

          <div onClick={(e) => e.stopPropagation()}>
            <ActionBar
              stats={thread.stats}
              isLiked={!!thread.is_liked}
              isReposted={!!thread.is_reposted}
              isBookmarked={!!thread.is_bookmarked}
              onLike={onLike}
              onRepost={onRepost}
              onBookmark={onBookmark}
              onReply={onReply}
              size="sm"
            />
          </div>
        </div>
      </div>
    </article>

    {reportModalOpen && (
      <div onClick={(e) => e.stopPropagation()}>
        <ReportPostModal
          postId={thread.id}
          onClose={() => setReportModalOpen(false)}
        />
      </div>
    )}
    <DeletePostModal
      open={deleteModalOpen}
      isPending={deleteMutation.isPending}
      onConfirm={() => void handleDeleteConfirm()}
      onCancel={() => setDeleteModalOpen(false)}
    />
    </>
  );
}
