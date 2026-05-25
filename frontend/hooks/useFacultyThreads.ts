'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getFacultyThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

export function useFacultyThreads(
  facultyId: number | null | undefined,
  enabled = true
) {
  const token = useAuthStore((s) => s.token);
  return useInfiniteQuery<ThreadListResponse>({
    queryKey: ['threads', 'faculty', facultyId],
    queryFn: ({ pageParam }) => {
      if (!token) throw new Error('Authentication required.');
      if (!facultyId) throw new Error('Faculty ID required.');
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getFacultyThreads(facultyId, { page, limit: PAGE_SIZE }, token);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled: !!token && !!facultyId && enabled,
  });
}
