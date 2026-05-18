'use client';

import { useQuery } from '@tanstack/react-query';
import { getEvent, type EventDetail } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useEvent(id: string) {
  const token = useAuthStore((state) => state.token);

  return useQuery<EventDetail>({
    queryKey: ['event', id],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch this event.');
      }
      return getEvent(id, token);
    },
    enabled: token !== null && id.length > 0,
  });
}
