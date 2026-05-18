'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { sendMessage, type ChatMessage, type ChatMedia, type ChatMessagesResponse, type ChatReplyInfo } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

interface SendVars {
  roomId: number;
  message: string | null;
  files: File[];
  gifUrl?: string | null;
  replyToId?: number | null;
}

interface MutationContext {
  optimistic: ChatMessage;
  blobUrls: string[];
}

export function useSendMessage() {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roomId, message, files, gifUrl, replyToId }: SendVars) =>
      sendMessage(roomId, message, files, token!, gifUrl, replyToId),

    onMutate: async ({ roomId, message, files, gifUrl, replyToId }): Promise<MutationContext> => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });

      const audioFiles = files.filter((f) => f.type.startsWith('audio/'));
      const blobUrls = audioFiles.map((f) => URL.createObjectURL(f));

      const optimisticMedia: ChatMedia[] = audioFiles.map((f, i) => ({
        media_id: -(Date.now() + i),
        url: blobUrls[i],
        media_type: f.type,
        file_size: f.size,
      }));

      // Try to resolve the reply_to preview from the cache
      let replyTo: ChatReplyInfo | undefined;
      if (replyToId) {
        const cached = queryClient.getQueryData<ChatMessagesResponse>(['chat', 'messages', roomId]);
        const original = cached?.data.find((m) => m.chat_id === replyToId);
        if (original) {
          replyTo = {
            chat_id: original.chat_id,
            sender: original.sender,
            message: original.message,
            gif_url: original.gif_url,
          };
        }
      }

      const optimistic: ChatMessage = {
        chat_id: -Date.now(),
        room_id: roomId,
        sender: { id: user!.id, username: user!.username },
        reply_to: replyTo ?? null,
        message,
        gif_url: gifUrl ?? null,
        media: optimisticMedia,
        reactions: [],
        is_read: false,
        is_edited: false,
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

      return { optimistic, blobUrls };
    },

    onSuccess: (_data, { roomId }, ctx) => {
      void queryClient.invalidateQueries({ queryKey: ['chat', 'messages', roomId] });
      if (ctx && ctx.blobUrls.length > 0) {
        const urls = ctx.blobUrls;
        setTimeout(() => urls.forEach((u) => URL.revokeObjectURL(u)), 10_000);
      }
    },

    onError: (_err, { roomId }, ctx) => {
      if (!ctx) return;
      ctx.blobUrls.forEach((u) => URL.revokeObjectURL(u));
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
