'use client';

import { useQuery } from '@tanstack/react-query';
import { getSummaryHistory, type SummaryHistoryResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

/**
 * Fetches the current user's thread-summary history (newest first). Enabled only
 * when `enabled` is true (the fox card is open) so the list isn't fetched on
 * every page. Keyed by ['summaries','history'] so a new summary can invalidate it.
 */
export function useSummaryHistory(enabled: boolean) {
  const token = useAuthStore((state) => state.token);

  return useQuery<SummaryHistoryResponse>({
    queryKey: ['summaries', 'history'],
    queryFn: () => getSummaryHistory({ page: 1, limit: 6 }, token as string),
    enabled: enabled && typeof token === 'string' && token.length > 0,
    staleTime: 30_000,
  });
}
