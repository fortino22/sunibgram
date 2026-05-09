'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useUserThreads(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  return useQuery<ThreadListResponse>({
    queryKey: ['threads', userId],
    queryFn: () => getMyThreads(userId, { limit: 20 }, token!),
    enabled: !!token && !!userId,
  });
}
