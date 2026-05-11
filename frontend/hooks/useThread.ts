'use client';

import { useQuery } from '@tanstack/react-query';
import {
  DEFAULT_THREAD_CHAIN_PARAMS,
  type ThreadItem,
  type ThreadResponse,
  getThread,
  getThreadChain,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export interface ThreadDetailData {
  root: ThreadItem;
  replies: ThreadItem[];
  rootThreadId: number | null;
}

function formatTimeAgo(value: string): string {
  const createdAt = new Date(value);

  if (Number.isNaN(createdAt.getTime())) {
    return value;
  }

  const diffMs = Date.now() - createdAt.getTime();
  const minute = 60_000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (diffMs < minute) {
    return 'now';
  }

  if (diffMs < hour) {
    return `${Math.max(1, Math.floor(diffMs / minute))}m`;
  }

  if (diffMs < day) {
    return `${Math.max(1, Math.floor(diffMs / hour))}h`;
  }

  if (diffMs < 7 * day) {
    return `${Math.max(1, Math.floor(diffMs / day))}d`;
  }

  const showYear = createdAt.getFullYear() !== new Date().getFullYear();
  const dateOptions: Intl.DateTimeFormatOptions = {
    day: 'numeric',
    month: 'short',
  };

  if (showYear) {
    dateOptions.year = 'numeric';
  }

  return createdAt.toLocaleDateString('en-US', dateOptions);
}

function mapThreadResponseToThreadItem(thread: ThreadResponse): ThreadItem {
  return {
    type: 'THREAD',
    id: thread.thread_id,
    parent_id: thread.parent_id,
    root_id: thread.root_id,
    community: undefined,
    content: thread.content,
    created_at: thread.created_at,
    updated_at: thread.updated_at,
    time_ago: formatTimeAgo(thread.created_at),
    action_at: undefined,
    reposted_at: thread.reposted_at,
    author: {
      id: thread.user_id,
      username: thread.username,
      is_premium: false,
      avatar_url: null,
      faculty: undefined,
    },
    stats: {
      likes: thread.like_count,
      reposts: thread.repost_count,
      bookmarks: thread.bookmark_count,
      comments: thread.comment_count,
    },
    media: thread.media,
    is_liked: false,
    is_bookmarked: false,
    is_reposted: false,
    is_pin: false,
  };
}

export function useThread(id: number | string) {
  const token = useAuthStore((state) => state.token);
  const hasValidId = id !== undefined && id !== null && String(id).length > 0;

  return useQuery<ThreadResponse>({
    queryKey: ['thread', id],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch this thread.');
      }

      return getThread(id, token);
    },
    enabled: token !== null && hasValidId,
  });
}

export function useThreadChain(id: number | string) {
  const token = useAuthStore((state) => state.token);
  const { page, limit } = DEFAULT_THREAD_CHAIN_PARAMS;
  const hasValidId = id !== undefined && id !== null && String(id).length > 0;
  const numericId = Number(id);

  return useQuery<ThreadDetailData>({
    queryKey: ['thread-chain', id, page, limit],
    queryFn: async () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch this thread chain.');
      }

      const [thread, chain] = await Promise.all([
        getThread(id, token),
        getThreadChain(id, { page, limit }, token),
      ]);
      const selectedThread = chain.data.find((item) => item.id === numericId);
      const root = selectedThread ?? mapThreadResponseToThreadItem(thread);
      const rootThreadId =
        thread.parent_id !== null && thread.parent_id !== undefined
          ? (thread.root_id ?? thread.parent_id)
          : null;

      return {
        root,
        replies: chain.data.filter((item) => item.parent_id === root.id),
        rootThreadId: rootThreadId === root.id ? null : rootThreadId,
      };
    },
    enabled: token !== null && hasValidId,
  });
}
