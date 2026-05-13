'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { markMessagesRead, type ChatMessagesResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useMarkMessagesRead() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roomId, chatIds }: { roomId: number; chatIds: number[] }) =>
      markMessagesRead(roomId, chatIds, token!),
    onMutate: ({ roomId, chatIds }) => {
      const chatIdSet = new Set(chatIds);
      queryClient.setQueryData<ChatMessagesResponse>(
        ['chat', 'messages', roomId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              chatIdSet.has(m.chat_id) ? { ...m, is_read: true } : m,
            ),
          };
        },
      );
    },
  });
}
