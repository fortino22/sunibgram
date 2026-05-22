import type { FeedAuthor, ThreadItem, ThreadMediaResponse } from '@/lib/services/threads';

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

export function getPollOptionMedia(thread: Pick<ThreadItem, 'poll' | 'media'>): Array<ThreadMediaResponse | null> {
  if (!thread.poll) {
    return [];
  }

  const fallbackMedia =
    thread.media.length === thread.poll.options.length
      ? thread.media.map((item) => {
          const url = item.url.trim();
          return url ? { ...item, url } : null;
        })
      : [];

  return thread.poll.options.map((option, index) => {
    const optionUrl = option.image_url?.trim();

    if (optionUrl) {
      return { media_id: option.option_id, url: optionUrl };
    }

    return fallbackMedia[index] ?? null;
  });
}

export function getPollMedia(thread: Pick<ThreadItem, 'poll' | 'media'>): ThreadMediaResponse[] {
  return getPollOptionMedia(thread).filter((item): item is ThreadMediaResponse => item !== null);
}

export function truncateContent(value: string, maxLength = 140): string {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}...` : value;
}
