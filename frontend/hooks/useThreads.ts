'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

export function useThreads() {
  const token = useAuthStore((state) => state.token);
  return useInfiniteQuery<ThreadListResponse>({
    queryKey: ['threads', 'feed'],
    queryFn: ({ pageParam }) => {
      if (!token) throw new Error('Authentication token is required to fetch threads.');
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getThreads({ page, limit: PAGE_SIZE }, token);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled: token !== null,
    // Don't reload the feed when the window regains focus — returning users should see
    // the same posts/scroll position they left. New posts surface via the banner tap only.
    refetchOnWindowFocus: false,
  });
}
