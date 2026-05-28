'use client';

import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { getMyTrendingHistory } from '@/lib/services/trendingUsers';

export function useMyTrendingHistory(startDate?: string, endDate?: string) {
  const token = useAuthStore((s) => s.token);

  return useQuery({
    queryKey: ['trending-users', 'my-history', startDate ?? null, endDate ?? null],
    queryFn: () => getMyTrendingHistory(token!, startDate, endDate),
    enabled: !!token,
    staleTime: 60_000,
  });
}
