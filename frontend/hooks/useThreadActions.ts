'use client';

import { useMutation, useQueryClient, type QueryClient, type InfiniteData } from '@tanstack/react-query';
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

function isInfiniteThreadListData(value: unknown): value is InfiniteData<ThreadListResponse> {
  return isRecord(value) && Array.isArray(value.pages) && (value.pages as unknown[]).every(isThreadListResponse);
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

function updateInfiniteThreadList(
  current: InfiniteData<ThreadListResponse>,
  threadId: number,
  updater: (thread: ThreadItem) => ThreadItem
): InfiniteData<ThreadListResponse> {
  return {
    ...current,
    pages: current.pages.map((page) => updateThreadList(page, threadId, updater)),
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

function updateRepostsState(thread: ThreadItem, isReposted: boolean): ThreadItem {
  return {
    ...thread,
    is_reposted: !isReposted,
    stats: {
      ...thread.stats,
      reposts: Math.max(0, thread.stats.reposts + (isReposted ? -1 : 1)),
    },
  };
}

function revertRepostsState(thread: ThreadItem, isReposted: boolean): ThreadItem {
  return {
    ...thread,
    is_reposted: isReposted,
    stats: {
      ...thread.stats,
      reposts: Math.max(0, thread.stats.reposts + (isReposted ? 1 : -1)),
    },
  };
}

function patchThreadCaches(
  queryClient: QueryClient,
  threadId: number,
  updater: (thread: ThreadItem) => ThreadItem
) {
  function patchAny(current: unknown): unknown {
    if (isInfiniteThreadListData(current)) return updateInfiniteThreadList(current, threadId, updater);
    if (isThreadListResponse(current)) return updateThreadList(current, threadId, updater);
    return current;
  }

  queryClient.setQueriesData({ queryKey: ['threads'] }, patchAny);
  queryClient.setQueriesData({ queryKey: ['likes'] }, patchAny);
  queryClient.setQueriesData({ queryKey: ['bookmarks'] }, patchAny);
  queryClient.setQueriesData({ queryKey: ['explore'] }, patchAny);
  queryClient.setQueriesData(
    { queryKey: ['thread-root'] },
    (current: unknown) => (isThreadItem(current) && current.id === threadId ? updater(current) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['thread-replies'] },
    (current: unknown) => (isInfiniteThreadListData(current) ? updateInfiniteThreadList(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['thread'] },
    (current: unknown) => (isThreadItem(current) && current.id === threadId ? updater(current) : current)
  );
  queryClient.setQueriesData({ queryKey: ['community'] }, patchAny);
  queryClient.setQueriesData({ queryKey: ['communities', 'threads'] }, patchAny);
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

  const isRepost = thread.type === 'repost' && typeof thread.original_thread_id === 'number';
  const effectiveId = isRepost ? (thread.original_thread_id as number) : thread.id;

  const like = useMutation({
    mutationFn: (isLiked: boolean) => {
      if (!token) {
        throw new Error('Authentication token is required to like this thread.');
      }

      return isLiked ? unlikeThread(effectiveId, token) : likeThread(effectiveId, token);
    },
    onMutate: async (isLiked) => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['threads'] }),
        queryClient.cancelQueries({ queryKey: ['likes'] }),
        queryClient.cancelQueries({ queryKey: ['bookmarks'] }),
        queryClient.cancelQueries({ queryKey: ['thread'] }),
        queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
        queryClient.cancelQueries({ queryKey: ['explore'] }),
        queryClient.cancelQueries({ queryKey: ['community'] }),
        queryClient.cancelQueries({ queryKey: ['communities', 'threads'] }),
      ]);

      patchThreadCaches(queryClient, thread.id, (current) => updateLikesState(current, isLiked));
      if (isRepost) {
        patchThreadCaches(queryClient, effectiveId, (current) => updateLikesState(current, isLiked));
      }

      patchInteractionList(
        queryClient,
        ['likes'],
        updateLikesState(thread, isLiked),
        isLiked ? 'remove' : 'add'
      );
    },
    onError: (_, isLiked) => {
      patchThreadCaches(queryClient, thread.id, (current) => revertLikesState(current, isLiked));
      if (isRepost) {
        patchThreadCaches(queryClient, effectiveId, (current) => revertLikesState(current, isLiked));
      }

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

      return isBookmarked ? unbookmarkThread(effectiveId, token) : bookmarkThread(effectiveId, token);
    },
    onMutate: async (isBookmarked) => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['threads'] }),
        queryClient.cancelQueries({ queryKey: ['likes'] }),
        queryClient.cancelQueries({ queryKey: ['bookmarks'] }),
        queryClient.cancelQueries({ queryKey: ['thread'] }),
        queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
        queryClient.cancelQueries({ queryKey: ['explore'] }),
        queryClient.cancelQueries({ queryKey: ['community'] }),
        queryClient.cancelQueries({ queryKey: ['communities', 'threads'] }),
      ]);

      patchThreadCaches(queryClient, thread.id, (current) => updateBookmarksState(current, isBookmarked));
      if (isRepost) {
        patchThreadCaches(queryClient, effectiveId, (current) => updateBookmarksState(current, isBookmarked));
      }

      patchInteractionList(
        queryClient,
        ['bookmarks'],
        updateBookmarksState(thread, isBookmarked),
        isBookmarked ? 'remove' : 'add'
      );
    },
    onError: (_, isBookmarked) => {
      patchThreadCaches(queryClient, thread.id, (current) => revertBookmarksState(current, isBookmarked));
      if (isRepost) {
        patchThreadCaches(queryClient, effectiveId, (current) => revertBookmarksState(current, isBookmarked));
      }

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
    mutationFn: (isReposted: boolean) => {
      if (!token) {
        throw new Error('Authentication token is required to repost this thread.');
      }

      return isReposted ? unrepostThread(effectiveId, token) : repostThread(effectiveId, token);
    },
    onMutate: async (isReposted) => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['threads'] }),
        queryClient.cancelQueries({ queryKey: ['thread'] }),
        queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
        queryClient.cancelQueries({ queryKey: ['explore'] }),
        queryClient.cancelQueries({ queryKey: ['community'] }),
        queryClient.cancelQueries({ queryKey: ['communities', 'threads'] }),
      ]);

      patchThreadCaches(queryClient, thread.id, (current) => updateRepostsState(current, isReposted));
      if (isRepost) {
        patchThreadCaches(queryClient, effectiveId, (current) => updateRepostsState(current, isReposted));
      }
    },
    onError: (_, isReposted) => {
      patchThreadCaches(queryClient, thread.id, (current) => revertRepostsState(current, isReposted));
      if (isRepost) {
        patchThreadCaches(queryClient, effectiveId, (current) => revertRepostsState(current, isReposted));
      }
    },
    onSettled: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['threads'] }),
        queryClient.invalidateQueries({ queryKey: ['thread'] }),
        queryClient.invalidateQueries({ queryKey: ['thread-chain'] }),
        queryClient.invalidateQueries({ queryKey: ['explore'] }),
        queryClient.invalidateQueries({ queryKey: ['community'] }),
        queryClient.invalidateQueries({ queryKey: ['communities', 'threads'] }),
      ]);
    },
  });

  return { like, bookmark, repost };
}
