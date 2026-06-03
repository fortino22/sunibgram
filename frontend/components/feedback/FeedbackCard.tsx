'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ThumbsUp, ThumbsDown, Trash2 } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { DeletePostModal } from '@/components/shared/DeletePostModal';
import { ImageZoomLightbox } from '@/components/shared/ImageZoomLightbox';
import { FeedbackTypeBadge } from './FeedbackTypeBadge';
import { FeedbackStatusBadge } from './FeedbackStatusBadge';
import { useAuthStore } from '@/stores/authStore';
import { useDeleteFeedback, useVoteFeedback } from '@/hooks/useFeedback';
import {
  type FeedbackResponse,
  type FeedbackVote,
} from '@/lib/services/feedback';

interface FeedbackCardProps {
  feedback: FeedbackResponse;
  view?: 'grid' | 'list';
}

function formatRelativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const diffSec = Math.floor(diffMs / 1000);
  if (diffSec < 60) return `${diffSec}s`;
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin}m`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h`;
  const diffDay = Math.floor(diffHr / 24);
  if (diffDay < 7) return `${diffDay}d`;
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

export function FeedbackCard({ feedback, view = 'grid' }: FeedbackCardProps) {
  const currentUser = useAuthStore((s) => s.user);
  const isAuthor = currentUser?.id === feedback.author.user_id;
  const isList = view === 'list';

  const [confirmDelete, setConfirmDelete] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [lightbox, setLightbox] = useState<number | null>(null);

  const voteMutation = useVoteFeedback();
  const deleteMutation = useDeleteFeedback();

  function handleVote(intent: FeedbackVote) {
    if (voteMutation.isPending) return;
    if (feedback.is_vote === intent) {
      // backend rejects re-vote; surface tiny inline hint without blocking
      setErrorMsg(`You already ${intent === 'like' ? 'liked' : 'disliked'} this.`);
      window.setTimeout(() => setErrorMsg(null), 2200);
      return;
    }
    setErrorMsg(null);
    voteMutation.mutate(
      { id: feedback.id, current: feedback.is_vote, next: intent },
      {
        onError: (err) => setErrorMsg(err instanceof Error ? err.message : 'Vote failed.'),
      },
    );
  }

  function handleConfirmDelete() {
    deleteMutation.mutate(feedback.id, {
      onSuccess: () => setConfirmDelete(false),
      onError: () => setConfirmDelete(false),
    });
  }

  const liked = feedback.is_vote === 'like';
  const disliked = feedback.is_vote === 'dislike';

  return (
    <article
      className="feedback-card"
      data-view={view}
      style={{
        padding: isList ? '14px 16px' : '16px',
        border: '1px solid var(--border)',
        borderRadius: isList ? 12 : 16,
        background: 'var(--bg-card)',
        display: 'flex',
        flexDirection: 'column',
        gap: isList ? 8 : 12,
        height: isList ? 'auto' : '100%',
      }}
    >
      <header style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
        <Link
          href={`/profile/${feedback.author.user_id}`}
          aria-label={`View ${feedback.author.name}'s profile`}
        >
          <Avatar
            username={feedback.author.username}
            avatarUrl={feedback.author.user_avatar}
            size={36}
          />
        </Link>

        <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 4 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
            <Link
              href={`/profile/${feedback.author.user_id}`}
              className="hover-underline"
              style={{
                fontSize: 14,
                fontWeight: 700,
                color: 'var(--text)',
                letterSpacing: '-0.01em',
                textDecoration: 'none',
              }}
            >
              {feedback.author.name}
            </Link>
            {feedback.author.is_premium && <VerifiedBadge size={14} />}
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              @{feedback.author.username}
            </span>
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>·</span>
            <span
              title={new Date(feedback.created_at).toLocaleString()}
              style={{ fontSize: 13, color: 'var(--text-tertiary)' }}
            >
              {formatRelativeTime(feedback.created_at)}
            </span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
            <FeedbackTypeBadge type={feedback.type} />
            <FeedbackStatusBadge status={feedback.status} />
          </div>
        </div>

        {isAuthor && (
          <button
            type="button"
            className="btn-ghost"
            onClick={() => setConfirmDelete(true)}
            aria-label="Delete feedback"
            title="Delete"
            style={{
              width: 32,
              height: 32,
              borderRadius: '50%',
              border: 'none',
              background: 'transparent',
              color: 'var(--text-tertiary)',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Trash2 size={15} strokeWidth={1.8} />
          </button>
        )}
      </header>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, flex: 1 }}>
        <h3
          style={{
            fontSize: 16,
            fontWeight: 600,
            color: 'var(--text)',
            letterSpacing: '-0.015em',
            lineHeight: 1.35,
            margin: 0,
          }}
        >
          {feedback.title}
        </h3>
        <p
          style={{
            fontSize: 14,
            color: 'var(--text-secondary)',
            lineHeight: 1.55,
            margin: 0,
            display: '-webkit-box',
            WebkitLineClamp: isList ? 2 : 3,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          {feedback.body}
        </p>

        {feedback.images.length > 0 && isList && (
          <div
            style={{
              display: 'flex',
              gap: 8,
              flexWrap: 'wrap',
            }}
          >
            {feedback.images.map((src, idx) => (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                key={src}
                src={src}
                alt="Feedback attachment"
                loading="lazy"
                role="button"
                onClick={() => setLightbox(idx)}
                style={{
                  width: 96,
                  height: 96,
                  objectFit: 'cover',
                  borderRadius: 10,
                  border: '1px solid var(--border)',
                  display: 'block',
                  flexShrink: 0,
                  cursor: 'zoom-in',
                }}
              />
            ))}
          </div>
        )}

        {feedback.images.length > 0 && !isList && (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns:
                feedback.images.length === 1 ? '1fr' : 'repeat(2, 1fr)',
              gap: 6,
              borderRadius: 12,
              overflow: 'hidden',
            }}
          >
            {feedback.images.map((src, idx) => {
              const count = feedback.images.length;
              // 1 img: wide banner. 3 imgs: first spans full width. else square.
              const spanFull = count === 3 && idx === 0;
              const ratio = count === 1 ? '16 / 10' : spanFull ? '16 / 7' : '1 / 1';
              return (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  key={src}
                  src={src}
                  alt="Feedback attachment"
                  loading="lazy"
                  role="button"
                  onClick={() => setLightbox(idx)}
                  style={{
                    width: '100%',
                    aspectRatio: ratio,
                    gridColumn: spanFull ? '1 / -1' : 'auto',
                    objectFit: 'cover',
                    border: '1px solid var(--border)',
                    display: 'block',
                    cursor: 'zoom-in',
                  }}
                />
              );
            })}
          </div>
        )}

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            marginTop: 'auto',
            paddingTop: 4,
            flexWrap: 'wrap',
          }}
        >
          <VoteButton
            kind="like"
            active={liked}
            count={feedback.likes}
            onClick={() => handleVote('like')}
          />
          <VoteButton
            kind="dislike"
            active={disliked}
            count={feedback.dislikes}
            onClick={() => handleVote('dislike')}
          />
        </div>

        {errorMsg && (
          <div style={{ fontSize: 12, color: 'var(--coral)', marginTop: 4 }}>{errorMsg}</div>
        )}
      </div>

      {lightbox !== null && (
        <ImageZoomLightbox
          images={feedback.images}
          index={lightbox}
          title={feedback.title}
          onClose={() => setLightbox(null)}
          onNavigate={setLightbox}
        />
      )}

      <DeletePostModal
        open={confirmDelete}
        isPending={deleteMutation.isPending}
        onConfirm={handleConfirmDelete}
        onCancel={() => { if (!deleteMutation.isPending) setConfirmDelete(false); }}
        title="Delete feedback?"
        description="This feedback and all its votes will be removed. This can't be undone."
      />
    </article>
  );
}

interface VoteButtonProps {
  kind: FeedbackVote;
  active: boolean;
  count: number;
  onClick: () => void;
}

function VoteButton({ kind, active, count, onClick }: VoteButtonProps) {
  const Glyph = kind === 'like' ? ThumbsUp : ThumbsDown;
  const accent = kind === 'like' ? 'var(--accent)' : 'var(--coral)';
  const idleColor = 'var(--text-tertiary)';
  const color = active ? accent : idleColor;
  const bg = active ? `color-mix(in oklab, ${accent} 14%, transparent)` : 'transparent';

  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      aria-label={kind === 'like' ? 'Like feedback' : 'Dislike feedback'}
      className="vote-btn"
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '6px 11px 6px 10px',
        borderRadius: 9999,
        border: '1px solid',
        borderColor: active ? `color-mix(in oklab, ${accent} 50%, transparent)` : 'var(--border)',
        background: bg,
        color,
        fontFamily: 'inherit',
        fontSize: 12.5,
        fontWeight: 600,
        cursor: 'pointer',
      }}
    >
      <Glyph
        size={14}
        strokeWidth={2}
        fill={active ? accent : 'none'}
        color={color}
      />
      <span style={{ minWidth: 8, textAlign: 'left' }}>{count}</span>
    </button>
  );
}
