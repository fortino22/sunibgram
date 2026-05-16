'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { unsendMessage, type ChatMessagesResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useDeleteMessage(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (chatId: number) => unsendMessage(chatId, token!),

    onMutate: async (chatId) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });
      queryClient.setQueryData<ChatMessagesResponse>(
        ['chat', 'messages', roomId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === chatId ? { ...m, deleted_at: new Date().toISOString() } : m,
            ),
          };
        },
      );
    },

    onError: () => {
      void queryClient.invalidateQueries({ queryKey: ['chat', 'messages', roomId] });
    },
  });
}
