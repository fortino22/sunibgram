'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteNotification, type PaginatedNotificationResponse } from '@/lib/services/notifications';
import { useAuthStore } from '@/stores/authStore';

export function useDeleteNotification() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => deleteNotification(id, token!),

    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ['notifications'] });

      const previous = queryClient.getQueryData<{
        pages: PaginatedNotificationResponse[];
        pageParams: unknown[];
      }>(['notifications']);

      queryClient.setQueryData<{
        pages: PaginatedNotificationResponse[];
        pageParams: unknown[];
      }>(['notifications'], (old) => {
        if (!old) return old;
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            data: page.data.filter((n) => n.id !== id),
          })),
        };
      });

      return { previous };
    },

    onError: (_err, _id, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['notifications'], context.previous);
      }
    },

    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}
