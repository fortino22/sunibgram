'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getFoxSessions, type FoxSessionsResponse } from '@/lib/services/fox';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

/**
 * Paginated Fox session history for the dedicated /summaries page (infinite
 * scroll). Search is server-side: `q` is sent to the backend and is part of the
 * query key so changing it starts a fresh paginated list. Distinct from
 * {@link useFoxSessions}, which fetches just the 6 most recent for the fox card.
 */
export function useFoxSessionsInfinite(q: string) {
  const token = useAuthStore((s) => s.token);
  const query = q.trim();

  return useInfiniteQuery<FoxSessionsResponse>({
    queryKey: ['fox', 'sessions', 'all', query],
    queryFn: ({ pageParam }) => {
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getFoxSessions({ page, limit: PAGE_SIZE, q: query }, token as string);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled: typeof token === 'string' && token.length > 0,
  });
}
