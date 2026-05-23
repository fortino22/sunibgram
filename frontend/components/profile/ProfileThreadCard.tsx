'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Users, Repeat2 } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { ActionBar } from '@/components/shared/ActionBar';
import { ThreadInlineEditor } from '@/components/thread/ThreadInlineEditor';
import { ThreadOptionsMenu } from '@/components/thread/ThreadOptionsMenu';
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
}

export function ProfileThreadCard({ thread, replyingTo, replyingToId, hideParentContext = false, onLike, onRepost, onBookmark, onReply, communityRole, onPinThread, onUnpinThread, pinningThread }: ProfileThreadCardProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(thread.id);
  const [isEditing, setIsEditing] = useState(false);
  const [lightboxState, setLightboxState] = useState<{ media: ThreadMediaResponse[]; index: number } | null>(null);
  const author = getThreadAuthor(thread);
  const canManage = isOwnThread(thread, currentUserId);
  const replyUsername = replyingTo ?? thread.parent_context?.author.username;
  const replyUserId = replyingToId ?? thread.parent_context?.author.id;

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
    <article
      onClick={() => router.push(`/thread/${thread.id}`)}
      style={{
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
        cursor: 'pointer',
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

      {thread.is_pin && (
        <div
          style={{
            padding: '8px 20px 0 55px',
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>📌 Pinned</span>
        </div>
      )}

      {thread.community && !thread.reposted_by_username && (
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

      <div style={{ padding: (thread.reposted_by_username || thread.community) ? '2px 20px 14px' : '14px 20px', display: 'flex', gap: 12 }}>
        <Avatar username={thread.author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />

        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: replyUsername ? 2 : 4 }}>
            <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em' }}>
              {thread.author.name || thread.author.username}
            </span>
            {thread.author.is_premium && (
              <span
                style={{
                  fontSize: 10,
                  background: 'var(--accent-soft)',
                  color: 'var(--accent)',
                  padding: '1px 5px',
                  borderRadius: 4,
                  fontWeight: 600,
                  letterSpacing: '0.02em',
                }}
              >
                PRO
              </span>
            )}
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              @{thread.author.username}
            </span>
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
            {(canManage || communityRole === 'ADMIN') && (
              <div
                style={{ marginLeft: 'auto' }}
                onClick={(e) => e.stopPropagation()}
              >
                <ThreadOptionsMenu
                  onEdit={canManage ? () => setIsEditing(true) : undefined}
                  onDelete={canManage ? () => void handleDelete() : undefined}
                  deleting={deleteMutation.isPending}
                  disabled={updateMutation.isPending}
                  compact
                  communityRole={communityRole}
                  isPinned={thread.is_pin === true}
                  pinning={pinningThread}
                  onPin={onPinThread ? () => onPinThread(thread.id) : undefined}
                  onUnpin={onUnpinThread ? () => onUnpinThread(thread.id) : undefined}
                />
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
  );
}
