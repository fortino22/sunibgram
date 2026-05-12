'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useUserThreads(userId: number | string | null | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery<ThreadListResponse>({
    queryKey: ['threads', userId],
    queryFn: () => {
      if (!token || userId === null || userId === undefined || userId === '') {
        throw new Error('A valid user id is required to load user threads.');
      }

      return getMyThreads(userId, { limit: 20 }, token);
    },
    enabled: !!token && userId !== null && userId !== undefined && userId !== '',
  });
}
