import type { FeedAuthor, ThreadItem } from '@/lib/services/threads';

export type ThreadAuthorWithAvatar = FeedAuthor;

export function getThreadAuthor(thread: ThreadItem): ThreadAuthorWithAvatar {
  return thread.author;
}

export function getAuthorAvatarUrl(author: ThreadAuthorWithAvatar): string | null {
  return author.avatar_url ?? null;
}

export function isRepostThread(thread: ThreadItem): boolean {
  return thread.type.toUpperCase() === 'REPOST' || Boolean(thread.reposted_at);
}

export function truncateContent(value: string, maxLength = 140): string {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}...` : value;
}
