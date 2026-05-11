'use client';

import { useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import type { ThreadDetailData } from '@/hooks/useThread';
import {
  likeThread, unlikeThread,
  bookmarkThread, unbookmarkThread,
  repostThread, unrepostThread,
  type ThreadListResponse, type ThreadItem,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isThreadItem(value: unknown): value is ThreadItem {
  return isRecord(value) && typeof value.id === 'number' && isRecord(value.stats);
}

function isThreadListResponse(value: unknown): value is ThreadListResponse {
  return (
    isRecord(value) &&
    Array.isArray(value.data) &&
    isRecord(value.meta) &&
    typeof value.meta.total === 'number'
  );
}

function isThreadDetailData(value: unknown): value is ThreadDetailData {
  return (
    isRecord(value) &&
    isThreadItem(value.root) &&
    Array.isArray(value.replies)
  );
}

function updateThreadList(
  current: ThreadListResponse,
  threadId: number,
  updater: (thread: ThreadItem) => ThreadItem
): ThreadListResponse {
  return {
    ...current,
    data: current.data.map((thread) => (thread.id === threadId ? updater(thread) : thread)),
  };
}

function updateThreadDetail(
  current: ThreadDetailData,
  threadId: number,
  updater: (thread: ThreadItem) => ThreadItem
): ThreadDetailData {
  return {
    ...current,
    root: current.root.id === threadId ? updater(current.root) : current.root,
    replies: current.replies.map((thread) => (thread.id === threadId ? updater(thread) : thread)),
  };
}

function upsertThreadList(current: ThreadListResponse, thread: ThreadItem): ThreadListResponse {
  const existingIndex = current.data.findIndex((item) => item.id === thread.id);

  if (existingIndex >= 0) {
    return {
      ...current,
      data: current.data.map((item) => (item.id === thread.id ? thread : item)),
    };
  }

  return {
    ...current,
    data: [thread, ...current.data],
    meta: {
      ...current.meta,
      total: current.meta.total + 1,
    },
  };
}

function removeFromThreadList(current: ThreadListResponse, threadId: number): ThreadListResponse {
  const nextData = current.data.filter((item) => item.id !== threadId);

  if (nextData.length === current.data.length) {
    return current;
  }

  return {
    ...current,
    data: nextData,
    meta: {
      ...current.meta,
      total: Math.max(0, current.meta.total - 1),
    },
  };
}

function updateLikesState(thread: ThreadItem, isLiked: boolean): ThreadItem {
  return {
    ...thread,
    is_liked: !isLiked,
    stats: {
      ...thread.stats,
      likes: Math.max(0, thread.stats.likes + (isLiked ? -1 : 1)),
    },
  };
}

function revertLikesState(thread: ThreadItem, isLiked: boolean): ThreadItem {
  return {
    ...thread,
    is_liked: isLiked,
    stats: {
      ...thread.stats,
      likes: Math.max(0, thread.stats.likes + (isLiked ? 1 : -1)),
    },
  };
}

function updateBookmarksState(thread: ThreadItem, isBookmarked: boolean): ThreadItem {
  return {
    ...thread,
    is_bookmarked: !isBookmarked,
    stats: {
      ...thread.stats,
      bookmarks: Math.max(0, thread.stats.bookmarks + (isBookmarked ? -1 : 1)),
    },
  };
}

function revertBookmarksState(thread: ThreadItem, isBookmarked: boolean): ThreadItem {
  return {
    ...thread,
    is_bookmarked: isBookmarked,
    stats: {
      ...thread.stats,
      bookmarks: Math.max(0, thread.stats.bookmarks + (isBookmarked ? 1 : -1)),
    },
  };
}

function patchThreadCaches(
  queryClient: QueryClient,
  threadId: number,
  updater: (thread: ThreadItem) => ThreadItem
) {
  queryClient.setQueriesData(
    { queryKey: ['threads'] },
    (current: unknown) => (isThreadListResponse(current) ? updateThreadList(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['likes'] },
    (current: unknown) => (isThreadListResponse(current) ? updateThreadList(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['bookmarks'] },
    (current: unknown) => (isThreadListResponse(current) ? updateThreadList(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['explore'] },
    (current: unknown) => (isThreadListResponse(current) ? updateThreadList(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['thread-chain'] },
    (current: unknown) => (isThreadDetailData(current) ? updateThreadDetail(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['thread'] },
    (current: unknown) => (isThreadItem(current) && current.id === threadId ? updater(current) : current)
  );
}

function patchInteractionList(
  queryClient: QueryClient,
  queryKey: readonly unknown[],
  thread: ThreadItem,
  action: 'add' | 'remove'
) {
  queryClient.setQueriesData(
    { queryKey },
    (current: unknown) => {
      if (!isThreadListResponse(current)) {
        return current;
      }

      return action === 'add'
        ? upsertThreadList(current, thread)
        : removeFromThreadList(current, thread.id);
    }
  );
}

export function useThreadActions(thread: ThreadItem) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  const like = useMutation({
    mutationFn: (isLiked: boolean) => {
      if (!token) {
        throw new Error('Authentication token is required to like this thread.');
      }

      return isLiked ? unlikeThread(thread.id, token) : likeThread(thread.id, token);
    },
    onMutate: async (isLiked) => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['threads'] }),
        queryClient.cancelQueries({ queryKey: ['likes'] }),
        queryClient.cancelQueries({ queryKey: ['bookmarks'] }),
        queryClient.cancelQueries({ queryKey: ['thread'] }),
        queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
        queryClient.cancelQueries({ queryKey: ['explore'] }),
      ]);

      patchThreadCaches(queryClient, thread.id, (current) => updateLikesState(current, isLiked));

      patchInteractionList(
        queryClient,
        ['likes'],
        updateLikesState(thread, isLiked),
        isLiked ? 'remove' : 'add'
      );
    },
    onError: (_, isLiked) => {
      patchThreadCaches(queryClient, thread.id, (current) => revertLikesState(current, isLiked));

      patchInteractionList(
        queryClient,
        ['likes'],
        revertLikesState(thread, isLiked),
        isLiked ? 'add' : 'remove'
      );
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ['likes'] });
    },
  });

  const bookmark = useMutation({
    mutationFn: (isBookmarked: boolean) => {
      if (!token) {
        throw new Error('Authentication token is required to bookmark this thread.');
      }

      return isBookmarked ? unbookmarkThread(thread.id, token) : bookmarkThread(thread.id, token);
    },
    onMutate: async (isBookmarked) => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['threads'] }),
        queryClient.cancelQueries({ queryKey: ['likes'] }),
        queryClient.cancelQueries({ queryKey: ['bookmarks'] }),
        queryClient.cancelQueries({ queryKey: ['thread'] }),
        queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
        queryClient.cancelQueries({ queryKey: ['explore'] }),
      ]);

      patchThreadCaches(queryClient, thread.id, (current) => updateBookmarksState(current, isBookmarked));

      patchInteractionList(
        queryClient,
        ['bookmarks'],
        updateBookmarksState(thread, isBookmarked),
        isBookmarked ? 'remove' : 'add'
      );
    },
    onError: (_, isBookmarked) => {
      patchThreadCaches(queryClient, thread.id, (current) => revertBookmarksState(current, isBookmarked));

      patchInteractionList(
        queryClient,
        ['bookmarks'],
        revertBookmarksState(thread, isBookmarked),
        isBookmarked ? 'add' : 'remove'
      );
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ['bookmarks'] });
    },
  });

  const repost = useMutation({
    mutationFn: (isReposted: boolean) =>
      isReposted
        ? unrepostThread(thread.id, token!)
        : repostThread(thread.id, token!),
    onMutate: (isReposted) => {
      patchThreadCaches(queryClient, thread.id, (current) => ({
        ...current,
        is_reposted: !isReposted,
        stats: {
          ...current.stats,
          reposts: Math.max(0, current.stats.reposts + (isReposted ? -1 : 1)),
        },
      }));
    },
    onError: (_, isReposted) => {
      patchThreadCaches(queryClient, thread.id, (current) => ({
        ...current,
        is_reposted: isReposted,
        stats: {
          ...current.stats,
          reposts: Math.max(0, current.stats.reposts + (isReposted ? 1 : -1)),
        },
      }));
    },
  });

  return { like, bookmark, repost };
}
