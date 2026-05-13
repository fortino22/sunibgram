'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { sendMessage, type ChatMessage, type ChatMessagesResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

interface SendVars {
  roomId: number;
  message: string | null;
  files: File[];
}

export function useSendMessage() {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roomId, message, files }: SendVars) =>
      sendMessage(roomId, message, files, token!),

    onMutate: async ({ roomId, message }) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });

      const optimistic: ChatMessage = {
        chat_id: -Date.now(),
        room_id: roomId,
        sender: { id: user!.id, username: user!.username },
        message,
        media: [],
        is_read: false,
        deleted_at: null,
        created_at: new Date().toISOString(),
      };

      queryClient.setQueryData(
        ['chat', 'messages', roomId],
        (old: ChatMessagesResponse | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: [optimistic, ...old.data],
            meta: { ...old.meta, total: old.meta.total + 1 },
          };
        },
      );

      return { optimistic };
    },

    onSuccess: (_data, { roomId }) => {
      // Fallback: if the socket message:new event was missed, refetch ensures
      // the real message replaces any lingering optimistic entry.
      void queryClient.invalidateQueries({ queryKey: ['chat', 'messages', roomId] });
    },

    onError: (_err, { roomId }, ctx) => {
      if (!ctx?.optimistic) return;
      queryClient.setQueryData(
        ['chat', 'messages', roomId],
        (old: ChatMessagesResponse | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.filter((m) => m.chat_id !== ctx.optimistic.chat_id),
            meta: { ...old.meta, total: old.meta.total - 1 },
          };
        },
      );
    },
  });
}
