'use client';

import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { getMyTrendingHistory } from '@/lib/services/trendingUsers';

export function useMyTrendingHistory(year?: number, month?: number) {
  const token = useAuthStore((s) => s.token);

  return useQuery({
    queryKey: ['trending-users', 'my-history', year ?? null, month ?? null],
    queryFn: () => getMyTrendingHistory(token!, year, month),
    enabled: !!token,
    staleTime: 60_000,
  });
}
