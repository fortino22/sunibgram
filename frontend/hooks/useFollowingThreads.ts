'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getFollowingThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

export function useFollowingThreads(enabled = true) {
  const token = useAuthStore((state) => state.token);
  return useInfiniteQuery<ThreadListResponse>({
    queryKey: ['threads', 'following'],
    queryFn: ({ pageParam }) => {
      if (!token) throw new Error('Authentication token is required to fetch following threads.');
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getFollowingThreads({ page, limit: PAGE_SIZE }, token);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled: token !== null && enabled,
  });
}
