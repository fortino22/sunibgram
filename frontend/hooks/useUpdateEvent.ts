'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateEvent, type UpdateEventInput } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useUpdateEvent(id: string) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateEventInput) => {
      if (!token) throw new Error('Not authenticated.');
      return updateEvent(id, input, token);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['events', 'feed'] }),
        queryClient.invalidateQueries({ queryKey: ['event', id] }),
      ]);
    },
  });
}
