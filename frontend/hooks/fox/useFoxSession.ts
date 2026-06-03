'use client';

import { useQuery } from '@tanstack/react-query';
import { getFoxSession, type FoxSessionDetail } from '@/lib/services/fox';
import { useAuthStore } from '@/stores/authStore';

/**
 * Fetches one chat session's turns (oldest-first) so the Ask Fox panel can
 * hydrate and continue it, and the /summaries page can show a read-only
 * transcript. Disabled until a session id is provided.
 */
export function useFoxSession(sessionId: number | null) {
  const token = useAuthStore((s) => s.token);

  return useQuery<FoxSessionDetail>({
    queryKey: ['fox', 'session', sessionId],
    queryFn: () => getFoxSession(sessionId as number, token as string),
    enabled: sessionId !== null && typeof token === 'string' && token.length > 0,
    staleTime: 30_000,
  });
}
