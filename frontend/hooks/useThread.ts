'use client';

import { useQuery } from '@tanstack/react-query';
import {
  getThread,
  getThreadChain,
  type ThreadListResponse,
  type ThreadResponse,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useThread(id: number | string) {
  const token = useAuthStore((state) => state.token);

  return useQuery<ThreadResponse>({
    queryKey: ['thread', id],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch this thread.');
      }

      return getThread(id, token);
    },
    enabled: token !== null && String(id).length > 0,
  });
}

export function useThreadChain(id: number | string) {
  const token = useAuthStore((state) => state.token);

  return useQuery<ThreadListResponse>({
    queryKey: ['thread-chain', id],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch this thread chain.');
      }

      return getThreadChain(id, {}, token);
    },
    enabled: token !== null && String(id).length > 0,
  });
}
