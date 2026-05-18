'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toggleMuteRoom, type ChatRoomsResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useMuteRoom() {
  const queryClient = useQueryClient();
  const token = useAuthStore((s) => s.token);

  return useMutation({
    mutationFn: ({ roomId }: { roomId: number }) =>
      toggleMuteRoom(roomId, token!),
    onMutate: async ({ roomId }) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'rooms'] });
      const prev = queryClient.getQueryData<ChatRoomsResponse>(['chat', 'rooms']);
      queryClient.setQueryData<ChatRoomsResponse>(['chat', 'rooms'], (old) => {
        if (!old) return old;
        return {
          ...old,
          data: old.data.map((r) =>
            r.room_id === roomId ? { ...r, is_muted: !r.is_muted } : r
          ),
        };
      });
      return { prev };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) queryClient.setQueryData(['chat', 'rooms'], ctx.prev);
    },
    onSuccess: (data, { roomId }) => {
      queryClient.setQueryData<ChatRoomsResponse>(['chat', 'rooms'], (old) => {
        if (!old) return old;
        return {
          ...old,
          data: old.data.map((r) =>
            r.room_id === roomId ? { ...r, is_muted: data.is_muted } : r
          ),
        };
      });
    },
  });
}
