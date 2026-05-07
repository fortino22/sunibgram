import type { FeedAuthor, ThreadItem } from '@/lib/services/threads';

export interface ThreadAuthorWithAvatar extends FeedAuthor {
  avatar_url?: string | null;
  image_url?: string | null;
  profile_image_url?: string | null;
  profileImageUrl?: string | null;
}

export function getThreadAuthor(thread: ThreadItem): ThreadAuthorWithAvatar {
  return thread.author as ThreadAuthorWithAvatar;
}

export function getAuthorAvatarUrl(author: ThreadAuthorWithAvatar): string | null {
  return author.profile_image_url ?? author.profileImageUrl ?? author.avatar_url ?? author.image_url ?? null;
}

export function isRepostThread(thread: ThreadItem): boolean {
  return thread.type.toUpperCase() === 'REPOST' || Boolean(thread.reposted_at);
}

export function truncateContent(value: string, maxLength = 140): string {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}...` : value;
}
