'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getNotifications, type PaginatedNotificationResponse } from '@/lib/services/notifications';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

export function useNotifications() {
  const token = useAuthStore((s) => s.token);

  return useInfiniteQuery<PaginatedNotificationResponse>({
    queryKey: ['notifications'],
    queryFn: ({ pageParam }) => {
      const offset = typeof pageParam === 'number' ? pageParam : 0;
      return getNotifications(token!, PAGE_SIZE, offset);
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage) => {
      if (lastPage.data.length < PAGE_SIZE) return undefined;
      return lastPage.meta.offset + lastPage.meta.limit;
    },
    enabled: !!token,
  });
}
