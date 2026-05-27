import type { MouseEvent } from 'react';
import { Bookmark, Heart, MessageCircle, Repeat2 } from 'lucide-react';
import { ActionButton } from './ActionButton';

interface ThreadActionsProps {
  comments: number;
  likes: number;
  reposts: number;
  bookmarks: number;
  isLiked: boolean;
  isReposted: boolean;
  isBookmarked: boolean;
  compact?: boolean;
  onComment?: () => void;
  onLike?: () => void;
  onRepost?: () => void;
  onBookmark?: () => void;
  onActionClick?: (event: MouseEvent<HTMLButtonElement>) => void;
  /** Inline top margin override in px. Defaults to 12. */
  topMargin?: number;
}

export function ThreadActions({
  comments,
  likes,
  reposts,
  bookmarks,
  isLiked,
  isReposted,
  isBookmarked,
  compact = false,
  onComment,
  onLike,
  onRepost,
  onBookmark,
  onActionClick,
  topMargin = 12,
}: ThreadActionsProps) {
  const iconSize = compact ? 16 : 18;
  const size = compact ? 'sm' : 'md';

  function handleClick(handler: (() => void) | undefined, event: MouseEvent<HTMLButtonElement>) {
    onActionClick?.(event);
    handler?.();
  }

  return (
    <div
      className="flex w-full justify-between text-[13px]"
      style={{ marginTop: topMargin, color: 'var(--text-tertiary)' }}
    >
      <ActionButton aria-label="Comment" onClick={(event) => handleClick(onComment, event)} tint="var(--accent)" softTint="var(--accent-soft)" size={size}>
        <MessageCircle size={iconSize} strokeWidth={2} />
        <span>{comments}</span>
      </ActionButton>
      <ActionButton
        aria-label="Repost"
        onClick={(event) => handleClick(onRepost, event)}
        active={isReposted}
        tint="var(--green)"
        softTint="var(--green-soft)"
        size={size}
      >
        <Repeat2 size={iconSize} strokeWidth={2} />
        <span>{reposts}</span>
      </ActionButton>
      <ActionButton
        aria-label="Like"
        onClick={(event) => handleClick(onLike, event)}
        active={isLiked}
        tint="var(--coral)"
        softTint="var(--coral-soft)"
        size={size}
      >
        <Heart size={iconSize} strokeWidth={2} fill={isLiked ? 'var(--coral)' : 'none'} />
        <span>{likes}</span>
      </ActionButton>
      <ActionButton
        aria-label="Bookmark"
        onClick={(event) => handleClick(onBookmark, event)}
        active={isBookmarked}
        tint="var(--accent)"
        softTint="var(--accent-soft)"
        size={size}
      >
        <Bookmark size={iconSize} strokeWidth={2} fill={isBookmarked ? 'var(--accent)' : 'none'} />
      </ActionButton>
    </div>
  );
}
