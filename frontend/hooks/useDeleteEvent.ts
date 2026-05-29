'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteEvent } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useDeleteEvent() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => {
      if (!token) throw new Error('Not authenticated.');
      return deleteEvent(id, token);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['events', 'feed'] }),
        queryClient.invalidateQueries({ queryKey: ['events', 'me'] }),
        queryClient.invalidateQueries({ queryKey: ['events', 'joined', 'me'] }),
      ]);
    },
  });
}
