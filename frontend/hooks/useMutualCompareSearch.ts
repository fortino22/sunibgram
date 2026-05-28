'use client';

import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { searchMutualCompareUsers } from '@/lib/services/trendingUsers';

export function useMutualCompareSearch(query: string, enabled = true) {
  const token = useAuthStore((s) => s.token);

  return useQuery({
    queryKey: ['trending-users', 'mutual-search', query],
    queryFn: () => searchMutualCompareUsers(token!, query),
    enabled: !!token && enabled,
    staleTime: 30_000,
  });
}
