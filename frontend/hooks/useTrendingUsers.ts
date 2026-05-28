'use client';

import { useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { getTrendingUsers } from '@/lib/services/trendingUsers';

export function useTrendingUsers(page = 1, limit = 20) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: ['trending-users', page, limit],
    queryFn: () => getTrendingUsers(token!, page, limit),
    enabled: !!token,
    staleTime: 30_000,
  });

  const refreshInSeconds = query.data?.meta.refreshInSeconds;

  useEffect(() => {
    if (refreshInSeconds === undefined) return;
    const delay = Math.max(5, refreshInSeconds) * 1000;
    const id = setTimeout(() => {
      queryClient.invalidateQueries({ queryKey: ['trending-users', page, limit] });
    }, delay);
    return () => clearTimeout(id);
  }, [refreshInSeconds, queryClient, page, limit]);

  return query;
}
