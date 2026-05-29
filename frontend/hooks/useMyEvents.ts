'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyEvents, type MyEventItem } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useMyEvents() {
  const token = useAuthStore((s) => s.token);
  return useQuery<MyEventItem[]>({
    queryKey: ['events', 'me'],
    queryFn: () => getMyEvents(token!),
    enabled: !!token,
    staleTime: 30_000,
  });
}
