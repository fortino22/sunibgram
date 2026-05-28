'use client';

import { useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { getMyTrendingCurrent } from '@/lib/services/trendingUsers';

export function useMyTrendingCurrent() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: ['trending-users', 'my-current'],
    queryFn: () => getMyTrendingCurrent(token!),
    enabled: !!token,
    staleTime: 30_000,
  });

  const refreshInSeconds = query.data?.meta.refreshInSeconds;

  useEffect(() => {
    if (refreshInSeconds === undefined) return;
    const delay = Math.max(5, refreshInSeconds) * 1000;
    const id = setTimeout(() => {
      queryClient.invalidateQueries({ queryKey: ['trending-users', 'my-current'] });
    }, delay);
    return () => clearTimeout(id);
  }, [refreshInSeconds, queryClient]);

  return query;
}
