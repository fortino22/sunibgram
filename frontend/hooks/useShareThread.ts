'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { recordShareCopy, shareViaChat, type ChatShareResponse, type CopyShareResponse, type ShareViaChatParams } from '@/lib/services/share';
import type { ThreadItem } from '@/lib/services/threads';

type ThreadListData = { data?: ThreadItem[] } | undefined;
type InfiniteData = { pages?: ThreadListData[] } | undefined;

function bumpShareCount(thread: ThreadItem): ThreadItem {
  return {
    ...thread,
    stats: {
      ...thread.stats,
      shares: (thread.stats.shares ?? 0) + 1,
    },
  };
}

function patchThreadEverywhere(qc: ReturnType<typeof useQueryClient>, threadId: number) {
  const visit = (item: ThreadItem): ThreadItem => {
    const targetId = item.original_thread_id ?? item.id;
    if (targetId === threadId) return bumpShareCount(item);
    return item;
  };
  const patchAny = (current: unknown): unknown => {
    if (current && typeof current === 'object' && 'pages' in current) {
      const inf = current as InfiniteData;
      if (!inf?.pages) return current;
      return {
        ...inf,
        pages: inf.pages.map((p) => {
          if (!p?.data) return p;
          return { ...p, data: p.data.map(visit) };
        }),
      };
    }
    if (current && typeof current === 'object' && 'data' in current) {
      const list = current as ThreadListData;
      if (!list?.data) return current;
      return { ...list, data: list.data.map(visit) };
    }
    return current;
  };
  ['threads', 'likes', 'bookmarks', 'explore', 'thread-replies', 'community', 'communities'].forEach((key) => {
    qc.setQueriesData({ queryKey: [key] }, patchAny);
  });
  qc.setQueriesData({ queryKey: ['thread'] }, (current: unknown) => {
    if (current && typeof current === 'object' && 'id' in current) {
      const item = current as ThreadItem;
      const targetId = item.original_thread_id ?? item.id;
      if (targetId === threadId) return bumpShareCount(item);
    }
    return current;
  });
}

export function useRecordShare(threadId: number) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();
  return useMutation<CopyShareResponse, Error>({
    mutationFn: () => {
      if (!token) throw new Error('Authentication required');
      return recordShareCopy(threadId, token);
    },
    onMutate: () => {
      patchThreadEverywhere(queryClient, threadId);
    },
  });
}

export function useShareViaChat() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();
  return useMutation<ChatShareResponse, Error, ShareViaChatParams>({
    mutationFn: (params) => {
      if (!token) throw new Error('Authentication required');
      return shareViaChat(params, token);
    },
    onSuccess: (_data, vars) => {
      patchThreadEverywhere(queryClient, Number(vars.threadId));
    },
  });
}
