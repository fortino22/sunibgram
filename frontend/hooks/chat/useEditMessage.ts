'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { editMessage, type ChatMessagesResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

interface EditVars {
  chatId: number;
  message: string;
}

export function useEditMessage(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ chatId, message }: EditVars) => editMessage(chatId, message, token!),

    onMutate: async ({ chatId, message }) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });
      queryClient.setQueryData<ChatMessagesResponse>(
        ['chat', 'messages', roomId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === chatId ? { ...m, message, is_edited: true } : m,
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
