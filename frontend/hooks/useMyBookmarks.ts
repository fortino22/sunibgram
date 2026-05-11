'use client';

import { useQuery } from '@tanstack/react-query';
import { getBookmarks, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useMyBookmarks() {
  const token = useAuthStore((s) => s.token);

  return useQuery<ThreadListResponse>({
    queryKey: ['bookmarks', 'me'],
    queryFn: () => getBookmarks(token!),
    enabled: !!token,
  });
}
