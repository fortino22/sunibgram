'use client';

import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { getUserTrendingHistory } from '@/lib/services/trendingUsers';

export function useUserTrendingHistory(
  userId: number | null,
  startDate?: string,
  endDate?: string,
) {
  const token = useAuthStore((s) => s.token);

  return useQuery({
    queryKey: [
      'trending-users',
      'user-history',
      userId,
      startDate ?? null,
      endDate ?? null,
    ],
    queryFn: () => getUserTrendingHistory(token!, userId!, startDate, endDate),
    enabled: !!token && userId !== null,
    staleTime: 60_000,
  });
}
