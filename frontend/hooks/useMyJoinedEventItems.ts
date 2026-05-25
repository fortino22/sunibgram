'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyJoinedEvents, type EventItem } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useMyJoinedEventItems() {
  const token = useAuthStore((s) => s.token);
  return useQuery<EventItem[]>({
    queryKey: ['events', 'joined', 'me'],
    queryFn: () => getMyJoinedEvents(token!),
    enabled: !!token,
    staleTime: 30_000,
  });
}
