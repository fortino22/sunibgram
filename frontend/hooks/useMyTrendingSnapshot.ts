'use client';

import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { getMyTrendingSnapshot } from '@/lib/services/trendingUsers';

export function useMyTrendingSnapshot(snapshotId: number | null) {
  const token = useAuthStore((s) => s.token);

  return useQuery({
    queryKey: ['trending-users', 'snapshot', snapshotId],
    queryFn: () => getMyTrendingSnapshot(token!, snapshotId!),
    enabled: !!token && snapshotId !== null,
    staleTime: 5 * 60_000,
  });
}
