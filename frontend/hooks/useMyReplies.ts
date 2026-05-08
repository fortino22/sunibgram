'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyReplies, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useMyReplies() {
  const token = useAuthStore((s) => s.token);
  return useQuery<ThreadListResponse>({
    queryKey: ['threads', 'me', 'replies'],
    queryFn: () => getMyReplies({ limit: 20 }, token!),
    enabled: !!token,
  });
}
