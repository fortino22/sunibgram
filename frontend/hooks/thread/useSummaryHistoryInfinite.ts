'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getSummaryHistory, type SummaryHistoryResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 20;

/**
 * Paginated summary history for the dedicated /summaries page (infinite scroll).
 * Search is server-side: `q` is sent to the backend and is part of the query key
 * so changing it starts a fresh paginated list. Distinct from
 * {@link useSummaryHistory}, which fetches just the 6 most recent for the fox card.
 */
export function useSummaryHistoryInfinite(q: string) {
  const token = useAuthStore((s) => s.token);
  const query = q.trim();

  return useInfiniteQuery<SummaryHistoryResponse>({
    queryKey: ['summaries', 'history', 'all', query],
    queryFn: ({ pageParam }) => {
      const page = typeof pageParam === 'number' ? pageParam : 1;
      return getSummaryHistory({ page, limit: PAGE_SIZE, q: query }, token as string);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.data.length, 0);
      return loaded < lastPage.meta.total ? allPages.length + 1 : undefined;
    },
    enabled: typeof token === 'string' && token.length > 0,
  });
}
