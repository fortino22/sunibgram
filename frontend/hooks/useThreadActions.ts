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

function effectiveThreadID(t: ThreadItem): number {
  if (t.type?.toLowerCase() === 'repost' && typeof t.original_thread_id === 'number') {
    return t.original_thread_id;
  }
  return t.id;
}

function updateThreadList(
  current: ThreadListResponse,
  threadId: number,
  updater: (thread: ThreadItem) => ThreadItem
): ThreadListResponse {
  return {
    ...current,
    data: current.data.map((thread) => (effectiveThreadID(thread) === threadId ? updater(thread) : thread)),
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
    (current: unknown) => (isThreadItem(current) && effectiveThreadID(current) === threadId ? updater(current) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['thread-replies'] },
    (current: unknown) => (isInfiniteThreadListData(current) ? updateInfiniteThreadList(current, threadId, updater) : current)
  );
  queryClient.setQueriesData(
    { queryKey: ['thread'] },
    (current: unknown) => (isThreadItem(current) && effectiveThreadID(current) === threadId ? updater(current) : current)
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

function findCanonicalThread(queryClient: QueryClient, effectiveId: number): ThreadItem | null {
  let found: ThreadItem | null = null;
  const scan = (current: unknown) => {
    if (found) return;
    if (isInfiniteThreadListData(current)) {
      for (const page of current.pages) {
        const hit = page.data.find((t) => effectiveThreadID(t) === effectiveId);
        if (hit) { found = hit; return; }
      }
    } else if (isThreadListResponse(current)) {
      const hit = current.data.find((t) => effectiveThreadID(t) === effectiveId);
      if (hit) { found = hit; return; }
    } else if (isThreadItem(current) && effectiveThreadID(current) === effectiveId) {
      found = current;
    }
  };
  for (const key of [['threads'], ['likes'], ['bookmarks'], ['explore'], ['thread-root'], ['thread'], ['community'], ['communities', 'threads']]) {
    if (found) break;
    const entries = queryClient.getQueriesData({ queryKey: key });
    for (const [, data] of entries) {
      scan(data);
      if (found) break;
    }
  }
  return found;
}

export function useThreadActions(thread: ThreadItem) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  const effectiveId = thread.type?.toLowerCase() === 'repost' && typeof thread.original_thread_id === 'number'
    ? thread.original_thread_id
    : thread.id;

  const readCanonicalFlag = (key: 'is_liked' | 'is_bookmarked' | 'is_reposted'): boolean => {
    const canonical = findCanonicalThread(queryClient, effectiveId);
    if (canonical) return canonical[key] === true;
    return thread[key] === true;
  };

  const like = useMutation({
    mutationFn: async () => {
      if (!token) {
        throw new Error('Authentication token is required to like this thread.');
      }
      const isLiked = readCanonicalFlag('is_liked');
      return isLiked ? unlikeThread(effectiveId, token) : likeThread(effectiveId, token);
    },
    onMutate: async () => {
      const isLiked = readCanonicalFlag('is_liked');
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

      patchThreadCaches(queryClient, effectiveId, (current) => updateLikesState(current, isLiked));

      patchInteractionList(
        queryClient,
        ['likes'],
        updateLikesState(thread, isLiked),
        isLiked ? 'remove' : 'add'
      );
      return { isLiked };
    },
    onError: (_err, _vars, context) => {
      const isLiked = context?.isLiked ?? false;
      patchThreadCaches(queryClient, effectiveId, (current) => revertLikesState(current, isLiked));

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
    mutationFn: async () => {
      if (!token) {
        throw new Error('Authentication token is required to bookmark this thread.');
      }
      const isBookmarked = readCanonicalFlag('is_bookmarked');
      return isBookmarked ? unbookmarkThread(effectiveId, token) : bookmarkThread(effectiveId, token);
    },
    onMutate: async () => {
      const isBookmarked = readCanonicalFlag('is_bookmarked');
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

      patchThreadCaches(queryClient, effectiveId, (current) => updateBookmarksState(current, isBookmarked));

      patchInteractionList(
        queryClient,
        ['bookmarks'],
        updateBookmarksState(thread, isBookmarked),
        isBookmarked ? 'remove' : 'add'
      );
      return { isBookmarked };
    },
    onError: (_err, _vars, context) => {
      const isBookmarked = context?.isBookmarked ?? false;
      patchThreadCaches(queryClient, effectiveId, (current) => revertBookmarksState(current, isBookmarked));

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
    mutationFn: async () => {
      if (!token) {
        throw new Error('Authentication token is required to repost this thread.');
      }
      const isReposted = readCanonicalFlag('is_reposted');
      return isReposted ? unrepostThread(effectiveId, token) : repostThread(effectiveId, token);
    },
    onMutate: async () => {
      const isReposted = readCanonicalFlag('is_reposted');
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['threads'] }),
        queryClient.cancelQueries({ queryKey: ['thread'] }),
        queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
        queryClient.cancelQueries({ queryKey: ['explore'] }),
        queryClient.cancelQueries({ queryKey: ['community'] }),
        queryClient.cancelQueries({ queryKey: ['communities', 'threads'] }),
      ]);

      patchThreadCaches(queryClient, effectiveId, (current) => updateRepostsState(current, isReposted));
      return { isReposted };
    },
    onError: (_err, _vars, context) => {
      const isReposted = context?.isReposted ?? false;
      patchThreadCaches(queryClient, effectiveId, (current) => revertRepostsState(current, isReposted));
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
