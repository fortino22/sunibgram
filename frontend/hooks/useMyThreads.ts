'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useMyThreads() {
  const token = useAuthStore((s) => s.token);
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery<ThreadListResponse>({
    queryKey: ['threads', 'me'],
    queryFn: () => getMyThreads(userId!, { limit: 20 }, token!),
    enabled: !!token && !!userId,
  });
}
