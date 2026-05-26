'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getFollowing } from '@/lib/services/follows';
import { useAuthStore } from '@/stores/authStore';

export function useFollowing(userId: number | string | undefined) {
  const token = useAuthStore((s) => s.token);
  return useInfiniteQuery({
    queryKey: ['following', userId],
    queryFn: ({ pageParam }) =>
      getFollowing(userId!, token!, { page: pageParam as number, limit: 20 }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const { page, limit, total } = lastPage.meta;
      return page * limit < total ? page + 1 : undefined;
    },
    enabled: !!token && !!userId,
  });
}
