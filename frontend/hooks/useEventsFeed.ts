'use client';

import { useQuery } from '@tanstack/react-query';
import { getEvents, type EventGroups } from '@/lib/services/events';
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
