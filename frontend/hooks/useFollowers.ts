'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getFollowers } from '@/lib/services/follows';
import { useAuthStore } from '@/stores/authStore';

export function useFollowers(userId: number | string | undefined) {
  const token = useAuthStore((s) => s.token);
  return useInfiniteQuery({
    queryKey: ['followers', userId],
    queryFn: ({ pageParam }) =>
      getFollowers(userId!, token!, { page: pageParam as number, limit: 20 }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const { page, limit, total } = lastPage.meta;
      return page * limit < total ? page + 1 : undefined;
    },
    enabled: !!token && !!userId,
  });
}
