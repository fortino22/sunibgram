'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyEvents, type MyEventItem } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

interface Options {
  status?: string;
  name?: string;
  enabled?: boolean;
}

export function useMyEvents(opts: Options = {}) {
  const { status, name, enabled = true } = opts;
  const token = useAuthStore((s) => s.token);
  const trimmedName = name?.trim() ?? '';
  return useQuery<MyEventItem[]>({
    queryKey: ['events', 'me', status ?? 'all', trimmedName],
    queryFn: () => getMyEvents(token!, status, trimmedName || undefined),
    enabled: !!token && enabled,
    staleTime: 30_000,
  });
}
