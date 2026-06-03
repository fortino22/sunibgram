'use client';

import { ThumbsUp, ThumbsDown, ImageIcon } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { AdminFeedbackTypeBadge, AdminFeedbackStatusBadge } from './FeedbackBadges';
import type { Feedback } from '@/lib/services/admin/types';

interface AdminFeedbackCardProps {
  feedback: Feedback;
  onOpen: () => void;
  onImageClick: (images: string[], index: number) => void;
}

export function formatRelativeTime(iso?: string): string {
  if (!iso) return '';
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

export function AdminFeedbackCard({ feedback, onOpen, onImageClick }: AdminFeedbackCardProps) {
  const author = feedback.author;
  const images = feedback.images ?? [];
  const shown = images.slice(0, 4);
  const extra = images.length - shown.length;

  return (
    <article
      role="button"
      tabIndex={0}
      onClick={onOpen}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onOpen(); } }}
      className="admin-row"
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
        padding: 16,
        border: '1px solid var(--border)',
        borderRadius: 14,
        background: 'var(--bg-card)',
        cursor: 'pointer',
      }}
    >
      <header style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
        <Avatar
          username={author?.name ?? author?.username ?? '?'}
          avatarUrl={author?.user_avatar}
          size={32}
        />
        <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
            <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.01em' }}>
              {author?.name ?? author?.username ?? 'Unknown'}
            </span>
            {author?.is_premium && <VerifiedBadge size={13} />}
            {author?.username && (
              <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{author.username}</span>
            )}
            {feedback.created_at && (
              <>
                <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>·</span>
                <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
                  {formatRelativeTime(feedback.created_at)}
                </span>
              </>
            )}
          </div>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
          <AdminFeedbackTypeBadge type={feedback.type} />
          <AdminFeedbackStatusBadge status={feedback.status} />
        </div>
      </header>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        <h3
          style={{
            fontSize: 15,
            fontWeight: 600,
            color: 'var(--text)',
            letterSpacing: '-0.015em',
            margin: 0,
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
          }}
        >
          {feedback.title}
        </h3>
        <p
          style={{
            fontSize: 13.5,
            color: 'var(--text-secondary)',
            lineHeight: 1.5,
            margin: 0,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          {feedback.body}
        </p>
      </div>

      {shown.length > 0 && (
        <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
          {shown.map((src, idx) => (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              key={src}
              src={src}
              alt="Feedback attachment"
              loading="lazy"
              onClick={(e) => { e.stopPropagation(); onImageClick(images, idx); }}
              style={{
                width: 56,
                height: 56,
                objectFit: 'cover',
                borderRadius: 8,
                border: '1px solid var(--border)',
                display: 'block',
                cursor: 'zoom-in',
              }}
            />
          ))}
          {extra > 0 && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 12,
                color: 'var(--text-tertiary)',
              }}
            >
              <ImageIcon size={13} />+{extra}
            </span>
          )}
        </div>
      )}

      <footer style={{ display: 'flex', alignItems: 'center', gap: 14, color: 'var(--text-tertiary)' }}>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, fontSize: 12.5 }}>
          <ThumbsUp size={13} strokeWidth={2} /> {feedback.likes ?? 0}
        </span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, fontSize: 12.5 }}>
          <ThumbsDown size={13} strokeWidth={2} /> {feedback.dislikes ?? 0}
        </span>
      </footer>
    </article>
  );
}
