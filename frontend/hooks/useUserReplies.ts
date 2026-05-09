'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyReplies, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useUserReplies(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  return useQuery<ThreadListResponse>({
    queryKey: ['threads', userId, 'replies'],
    queryFn: () => getMyReplies(userId, { limit: 20 }, token!),
    enabled: !!token && !!userId,
  });
}
