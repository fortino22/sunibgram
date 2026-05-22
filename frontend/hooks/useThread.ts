'use client';

import { useQuery } from '@tanstack/react-query';
import {
  DEFAULT_THREAD_CHAIN_PARAMS,
  type ThreadItem,
  type ThreadResponse,
  getThread,
  getThreadFeed,
  getThreadChain,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

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
        getThreadFeed(id, token),
        getThreadChain(id, { page, limit }, token),
      ]);

      const isRepost = thread.type === 'REPOST' && thread.original_thread_id != null;
      const effectiveId = isRepost ? thread.original_thread_id! : numericId;

      const selectedThread = chain.data.find((item) => item.id === effectiveId);

      let root: ThreadItem;
      if (isRepost) {
        const base = selectedThread ?? thread;
        root = {
          ...base,
          reposted_by_id: thread.reposted_by_id,
          reposted_by_username: thread.reposted_by_username,
          reposted_by_name: thread.reposted_by_name,
          reposted_at: thread.reposted_at,
          // Prefer poll from direct fetch — chain endpoint may omit image_url
          ...(thread.poll ? { poll: thread.poll } : {}),
        };
      } else {
        // Use direct-fetch thread as root: it includes poll option image_url
        // Chain endpoint may omit image_url from poll options
        root = thread;
      }

      const rootThreadId =
        thread.parent_id !== null && thread.parent_id !== undefined
          ? (thread.root_id ?? thread.parent_id)
          : null;

      return {
        root,
        replies: chain.data.filter((item) => item.parent_id === effectiveId),
        rootThreadId: rootThreadId === root.id ? null : rootThreadId,
      };
    },
    enabled: token !== null && hasValidId,
  });
}
