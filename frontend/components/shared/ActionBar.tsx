'use client';

import type { MouseEvent } from 'react';
import { Bookmark, Heart, MessageCircle, Repeat2 } from 'lucide-react';
import { ActionButton } from '@/components/thread/ActionButton';

interface ActionBarStats {
  likes: number;
  reposts: number;
  bookmarks: number;
  comments: number;
}

interface ActionBarProps {
  stats: ActionBarStats;
  isLiked: boolean;
  isReposted: boolean;
  isBookmarked: boolean;
  onLike?: () => void;
  onRepost?: () => void;
  onBookmark?: (e: MouseEvent<HTMLButtonElement>) => void;
  onReply?: () => void;
  size?: 'sm' | 'md';
}

export function ActionBar({
  stats,
  isLiked,
  isReposted,
  isBookmarked,
  onLike,
  onRepost,
  onBookmark,
  onReply,
  size = 'md',
}: ActionBarProps) {
  const iconSize = size === 'sm' ? 18 : 20;
  const sw = 2;

  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 10, color: 'var(--text-tertiary)' }}>
      <ActionButton aria-label="Reply" onClick={onReply} tint="var(--accent)" softTint="var(--accent-soft)" size={size}>
        <MessageCircle size={iconSize} strokeWidth={sw} />
        <span style={{ fontSize: 14 }}>{stats.comments}</span>
      </ActionButton>

      <ActionButton
        aria-label="Repost"
        onClick={onRepost}
        active={isReposted}
        tint="var(--green)"
        softTint="var(--green-soft)"
        size={size}
      >
        <Repeat2 size={iconSize} strokeWidth={sw} />
        <span style={{ fontSize: 14 }}>{stats.reposts}</span>
      </ActionButton>

      <ActionButton
        aria-label="Like"
        onClick={onLike}
        active={isLiked}
        tint="var(--coral)"
        softTint="var(--coral-soft)"
        size={size}
      >
        <Heart size={iconSize} strokeWidth={sw} fill={isLiked ? 'var(--coral)' : 'none'} />
        <span style={{ fontSize: 14 }}>{stats.likes}</span>
      </ActionButton>

      <ActionButton
        aria-label="Bookmark"
        onClick={onBookmark}
        active={isBookmarked}
        tint="var(--accent)"
        softTint="var(--accent-soft)"
        size={size}
      >
        <Bookmark
          size={iconSize}
          strokeWidth={sw}
          fill={isBookmarked ? 'var(--accent)' : 'none'}
        />
      </ActionButton>
    </div>
  );
}
