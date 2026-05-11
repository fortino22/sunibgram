'use client';

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
  onBookmark?: () => void;
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
  const iconSize = size === 'sm' ? 15 : 18;
  const sw = size === 'sm' ? 1.8 : 2;

  return (
    <div style={{ display: 'flex', gap: 4, marginTop: 12 }}>
      <ActionButton aria-label="Reply" onClick={onReply} size={size}>
        <MessageCircle size={iconSize} strokeWidth={sw} />
        <span style={{ fontSize: size === 'sm' ? 13 : 14 }}>{stats.comments}</span>
      </ActionButton>

      <ActionButton
        aria-label="Repost"
        onClick={onRepost}
        color={isReposted ? 'var(--green)' : undefined}
        size={size}
      >
        <Repeat2 size={iconSize} strokeWidth={sw} />
        <span style={{ fontSize: size === 'sm' ? 13 : 14 }}>{stats.reposts}</span>
      </ActionButton>

      <ActionButton
        aria-label="Like"
        onClick={onLike}
        color={isLiked ? 'var(--coral)' : undefined}
        size={size}
      >
        <Heart size={iconSize} strokeWidth={sw} fill={isLiked ? 'var(--coral)' : 'none'} />
        <span style={{ fontSize: size === 'sm' ? 13 : 14 }}>{stats.likes}</span>
      </ActionButton>

      <ActionButton
        aria-label="Bookmark"
        onClick={onBookmark}
        color={isBookmarked ? 'var(--accent)' : undefined}
        size={size}
      >
        <Bookmark
          size={iconSize}
          strokeWidth={sw}
          fill={isBookmarked ? 'var(--accent)' : 'none'}
        />
        <span style={{ fontSize: size === 'sm' ? 13 : 14 }}>{stats.bookmarks}</span>
      </ActionButton>
    </div>
  );
}
