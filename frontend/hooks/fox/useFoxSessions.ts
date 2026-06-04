'use client';

import { useQuery } from '@tanstack/react-query';
import { getFoxSessions, type FoxSessionsResponse } from '@/lib/services/fox';
import { useAuthStore } from '@/stores/authStore';

/**
 * Fetches the current user's Fox sessions (newest first) for the fox card's
 * History tab — just the 6 most recent. Enabled only when `enabled` is true (the
 * card is open). Keyed ['fox','sessions'] so a new chat turn / summary can
 * invalidate it.
 */
export function useFoxSessions(enabled: boolean) {
  const token = useAuthStore((state) => state.token);

  return useQuery<FoxSessionsResponse>({
    queryKey: ['fox', 'sessions'],
    queryFn: () => getFoxSessions({ page: 1, limit: 6 }, token as string),
    enabled: enabled && typeof token === 'string' && token.length > 0,
    staleTime: 30_000,
  });
}
