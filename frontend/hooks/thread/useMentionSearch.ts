'use client';

import { useQuery } from '@tanstack/react-query';
import { searchMentions, type MentionUser } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useMentionSearch(query: string, enabled: boolean) {
  const token = useAuthStore((s) => s.token);
  const trimmed = query.trim();
  return useQuery<MentionUser[]>({
    queryKey: ['mentions', trimmed],
    queryFn: () => searchMentions(trimmed, token!),
    enabled: enabled && trimmed.length > 0 && token !== null,
    staleTime: 30_000,
  });
}
