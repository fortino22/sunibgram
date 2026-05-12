import type { FeedAuthor, ThreadItem } from '@/lib/services/threads';

export type ThreadAuthorWithAvatar = FeedAuthor;

export function getThreadAuthor(thread: ThreadItem): ThreadAuthorWithAvatar {
  return thread.author;
}

export function getAuthorAvatarUrl(author: ThreadAuthorWithAvatar): string | null {
  return author.avatar_url ?? null;
}

export function isRepostThread(thread: ThreadItem): boolean {
  return thread.type.toUpperCase() === 'REPOST' || thread.reposted_by_id !== undefined || Boolean(thread.reposted_at);
}

export function getRepostLabel(thread: ThreadItem, currentUserId: number | null | undefined): string | null {
  if (!isRepostThread(thread)) {
    return null;
  }

  if (currentUserId !== null && currentUserId !== undefined && thread.reposted_by_id === currentUserId) {
    return 'You reposted';
  }

  const displayName = thread.reposted_by_name?.trim();
  if (displayName) {
    return `${displayName} reposted`;
  }

  const username = thread.reposted_by_username?.trim();
  if (username) {
    return `@${username} reposted`;
  }

  return null;
}

export function isOwnThread(thread: ThreadItem, userId: number | null | undefined): boolean {
  return userId !== null && userId !== undefined && thread.author.id === userId;
}

export function truncateContent(value: string, maxLength = 140): string {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}...` : value;
}
