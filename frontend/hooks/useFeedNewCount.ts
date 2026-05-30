'use client';

import { useQuery } from '@tanstack/react-query';
import { getFeedNewCount } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const POLL_INTERVAL_MS = 180_000; // 3 minutes

// Polls the "new posts" count for the For You feed. `sinceId` is the max loaded thread id;
// `active` gates polling to when the For You tab is showing.
export function useFeedNewCount(sinceId: number, active: boolean): number {
  const token = useAuthStore((state) => state.token);
  const query = useQuery({
    queryKey: ['threads', 'feed', 'new-count', sinceId],
    queryFn: () => {
      if (!token) throw new Error('Authentication token is required to fetch the feed count.');
      return getFeedNewCount(sinceId, token);
    },
    enabled: active && Boolean(token) && sinceId > 0,
    refetchInterval: POLL_INTERVAL_MS,
    refetchIntervalInBackground: false,
    retry: false,
  });
  return query.data?.count ?? 0;
}
