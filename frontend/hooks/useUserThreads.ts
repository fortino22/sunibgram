'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getMyThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

export function useUserThreads(
  userId: number | string | null | undefined,
  options?: { enabled?: boolean }
) {
  const token = useAuthStore((s) => s.token);
  return useInfiniteQuery<ThreadListResponse>({
    queryKey: ['threads', userId],
    queryFn: ({ pageParam }) => {
      if (!token || userId === null || userId === undefined || userId === '') {
        throw new Error('A valid user id is required to load user threads.');
      }
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getMyThreads(userId, { page, limit: PAGE_SIZE }, token);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled:
      (options?.enabled ?? true) &&
      !!token &&
      userId !== null &&
      userId !== undefined &&
      userId !== '',
  });
}
