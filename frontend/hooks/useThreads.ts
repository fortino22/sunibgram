'use client';

import { useQuery } from '@tanstack/react-query';
import { getThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useThreads(page = 1, limit = 10) {
  const token = useAuthStore((state) => state.token);

  return useQuery<ThreadListResponse>({
    queryKey: ['threads', page, limit],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch threads.');
      }

      return getThreads({ page, limit }, token);
    },
    enabled: token !== null,
  });
}
