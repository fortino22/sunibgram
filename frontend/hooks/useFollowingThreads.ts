'use client';

import { useQuery } from '@tanstack/react-query';
import { getFollowingThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useFollowingThreads(page = 1, limit = 10, enabled = true) {
  const token = useAuthStore((state) => state.token);

  return useQuery<ThreadListResponse>({
    queryKey: ['threads', 'following', page, limit],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch following threads.');
      }

      return getFollowingThreads({ page, limit }, token);
    },
    enabled: token !== null && enabled,
  });
}
