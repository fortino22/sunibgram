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
import type { ThreadItem, ThreadMutationParams } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { getAuthorAvatarUrl, getThreadAuthor, isOwnThread } from '@/components/thread/threadUtils';
import { ThreadBody } from '@/components/thread/ThreadBody';
import { ImageDetailModal } from '@/components/thread/ImageDetailModal';

interface ProfileThreadCardProps {
  thread: ThreadItem;
  replyingTo?: string;
  replyingToId?: number;
  onLike?: () => void;
  onRepost?: () => void;
  onBookmark?: () => void;
}

export function ProfileThreadCard({ thread, replyingTo, replyingToId, onLike, onRepost, onBookmark }: ProfileThreadCardProps) {
  const router = useRouter();
  const currentUserId = useAuthStore((state) => state.user?.id);
  const { updateMutation, deleteMutation } = useThreadManagement(thread.id);
  const [isEditing, setIsEditing] = useState(false);
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);
  const author = getThreadAuthor(thread);
  const canManage = isOwnThread(thread, currentUserId);

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

      {thread.community && (
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

      {replyingTo && (
        <div
          style={{
            padding: '10px 20px 0 72px',
            fontSize: 13,
            color: 'var(--text-tertiary)',
            letterSpacing: '-0.01em',
          }}
        >
          Replying to{' '}
          {replyingToId ? (
            <Link
              href={`/profile/${replyingToId}`}
              onClick={(e) => e.stopPropagation()}
              style={{ color: 'var(--accent)', textDecoration: 'none' }}
            >
              @{replyingTo}
            </Link>
          ) : (
            <span style={{ color: 'var(--accent)' }}>@{replyingTo}</span>
          )}
        </div>
      )}

      <div style={{ padding: (thread.reposted_by_username || thread.community) ? '2px 20px 14px' : '14px 20px', display: 'flex', gap: 12 }}>
        <Avatar username={thread.author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />

        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
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
            {thread.is_pin && (
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)', marginLeft: 'auto' }}>
                Pinned
              </span>
            )}
            {canManage && (
              <div
                style={{ marginLeft: thread.is_pin ? 8 : 'auto' }}
                onClick={(e) => e.stopPropagation()}
              >
                <ThreadOptionsMenu
                  onEdit={() => setIsEditing(true)}
                  onDelete={() => void handleDelete()}
                  deleting={deleteMutation.isPending}
                  disabled={updateMutation.isPending}
                  compact
                />
              </div>
            )}
          </div>

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
            <ThreadBody
              content={thread.content}
              media={thread.media}
              onImageClick={(i) => setLightboxIndex(i)}
            />
          )}

          {lightboxIndex !== null && (
            <div onClick={(e) => e.stopPropagation()}>
              <ImageDetailModal
                media={thread.media}
                currentIndex={lightboxIndex}
                thread={thread}
                onClose={() => setLightboxIndex(null)}
                onNavigate={setLightboxIndex}
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
              size="sm"
            />
          </div>
        </div>
      </div>
    </article>
  );
}
