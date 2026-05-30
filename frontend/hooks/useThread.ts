'use client';

import { useQuery, useInfiniteQuery } from '@tanstack/react-query';
import {
  type ThreadItem,
  type ThreadResponse,
  type ThreadListResponse,
  getThread,
  getThreadFeed,
  getThreadChain,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const REPLIES_PAGE_SIZE = 20;

export interface ThreadDetailData {
  root: ThreadItem;
  replies: ThreadItem[];
  rootThreadId: number | null;
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

export interface ThreadChainResult {
  root: ThreadItem | undefined;
  ancestors: ThreadItem[];
  replies: ThreadItem[];
  rootThreadId: number | null | undefined;
  hasNextPage: boolean;
  fetchNextPage: () => void;
  isFetchingNextPage: boolean;
  isLoading: boolean;
  isError: boolean;
}

export function useThreadChain(id: number | string): ThreadChainResult {
  const token = useAuthStore((state) => state.token);
  const hasValidId = id !== undefined && id !== null && String(id).length > 0;
  const numericId = Number(id);

  const rootQuery = useQuery({
    queryKey: ['thread-root', id],
    queryFn: () => {
      if (!token) throw new Error('Authentication token is required.');
      return getThreadFeed(id, token);
    },
    enabled: !!token && hasValidId,
  });

  const thread = rootQuery.data;
  const isRepost = thread?.type === 'REPOST' && thread?.original_thread_id != null;
  const effectiveId = isRepost ? thread!.original_thread_id! : numericId;

  const repliesQuery = useInfiniteQuery<ThreadListResponse>({
    queryKey: ['thread-replies', id],
    queryFn: ({ pageParam }) => {
      if (!token) throw new Error('Authentication token is required.');
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getThreadChain(id, { page, limit: REPLIES_PAGE_SIZE }, token);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled: !!token && hasValidId,
  });

  let root: ThreadItem | undefined;
  if (thread) {
    if (isRepost) {
      const selectedThread = repliesQuery.data?.pages[0]?.data.find((item) => item.id === effectiveId);
      const base = selectedThread ?? thread;
      root = {
        ...base,
        reposted_by_id: thread.reposted_by_id,
        reposted_by_username: thread.reposted_by_username,
        reposted_by_name: thread.reposted_by_name,
        reposted_at: thread.reposted_at,
        ...(thread.poll ? { poll: thread.poll } : {}),
      };
    } else {
      root = thread;
    }
  }

  const replies = repliesQuery.data?.pages.flatMap((p) =>
    p.data.filter((item) => item.parent_id === effectiveId)
  ) ?? [];

  // Walk back from the anchor's parent through page-1 items to build the
  // ancestor chain in display order (root → … → immediate parent).
  // Backend returns [ancestors…, anchor, children…] in page 1, so all
  // ancestors are reachable for any reasonable chain depth.
  const ancestors: ThreadItem[] = [];
  if (root?.parent_id != null) {
    const byId = new Map<number, ThreadItem>();
    for (const item of repliesQuery.data?.pages[0]?.data ?? []) {
      byId.set(item.id, item);
    }
    const seen = new Set<number>();
    let cur = byId.get(root.parent_id);
    while (cur && !seen.has(cur.id)) {
      seen.add(cur.id);
      ancestors.unshift(cur);
      cur = cur.parent_id != null ? byId.get(cur.parent_id) : undefined;
    }
  }

  const rootThreadId =
    thread?.parent_id !== null && thread?.parent_id !== undefined
      ? (thread.root_id ?? thread.parent_id)
      : null;

  return {
    root,
    ancestors,
    replies,
    rootThreadId: rootThreadId === root?.id ? null : rootThreadId,
    hasNextPage: repliesQuery.hasNextPage ?? false,
    fetchNextPage: repliesQuery.fetchNextPage,
    isFetchingNextPage: repliesQuery.isFetchingNextPage,
    isLoading: rootQuery.isLoading || repliesQuery.isLoading,
    isError: rootQuery.isError || repliesQuery.isError,
  };
}
