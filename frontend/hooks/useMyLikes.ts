'use client';

import { useQuery } from '@tanstack/react-query';
import { getLikes, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useMyLikes() {
  const token = useAuthStore((s) => s.token);
  return useQuery<ThreadListResponse>({
    queryKey: ['likes', 'me'],
    queryFn: () => getLikes(token!),
    enabled: !!token,
  });
}
