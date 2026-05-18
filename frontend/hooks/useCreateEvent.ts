'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createEvent, type CreateEventInput } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useCreateEvent() {
  const token = useAuthStore((state) => state.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateEventInput) => {
      if (token === null) {
        throw new Error('Authentication token is required to create an event.');
      }
      return createEvent(input, token);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['events', 'feed'] });
    },
  });
}
