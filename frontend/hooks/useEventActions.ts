'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  bookmarkEvent,
  shareEvent,
  unbookmarkEvent,
  type EventShareResult,
} from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useEventActions(eventId: string | undefined) {
  const token = useAuthStore((state) => state.token);
  const queryClient = useQueryClient();

  const bookmark = useMutation({
    mutationFn: (isCurrentlyBookmarked: boolean) => {
      if (!eventId) throw new Error('Event ID missing.');
      if (!token) throw new Error('Not authenticated.');
      return isCurrentlyBookmarked
        ? unbookmarkEvent(eventId, token)
        : bookmarkEvent(eventId, token);
    },
    onSettled: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['events', 'feed'] }),
        ...(eventId ? [queryClient.invalidateQueries({ queryKey: ['event', eventId] })] : []),
      ]);
    },
  });

  const share = useMutation<EventShareResult>({
    mutationFn: () => {
      if (!eventId) throw new Error('Event ID missing.');
      if (!token) throw new Error('Not authenticated.');
      return shareEvent(eventId, token);
    },
    onSettled: async () => {
      if (eventId) {
        await queryClient.invalidateQueries({ queryKey: ['event', eventId] });
      }
    },
  });

  return { bookmark, share };
}
