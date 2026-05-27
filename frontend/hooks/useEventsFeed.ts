'use client';

import { useQuery } from '@tanstack/react-query';
import { getEvents, getTrendingEvents, type EventGroups, type EventItem } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useEventsFeed(enabled = true) {
  const token = useAuthStore((state) => state.token);

  return useQuery<EventGroups>({
    queryKey: ['events', 'feed'],
    queryFn: () => {
      if (token === null) throw new Error('Authentication token required.');
      return getEvents(token);
    },
    enabled: token !== null && enabled,
  });
}

export function useTrendingEvents() {
  const token = useAuthStore((state) => state.token);

  return useQuery<EventItem[]>({
    queryKey: ['events', 'trending'],
    queryFn: () => {
      if (token === null) throw new Error('Authentication token required.');
      return getTrendingEvents(token);
    },
    enabled: token !== null,
    staleTime: 5 * 60 * 1000,
  });
}
