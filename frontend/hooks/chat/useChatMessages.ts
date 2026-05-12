'use client';

import { useEffect, useRef } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getMessages,
  type ChatMessage,
  type ChatMessagesResponse,
  type MessageEditedPayload,
  type MessageDeletedPayload,
} from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';
import { socketManager } from '@/lib/socket';

export function useChatMessages(roomId: number | null) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();
  const joinedRef = useRef<number | null>(null);

  // Join / leave socket room as activeRoomId changes
  useEffect(() => {
    if (!roomId) return;

    // Leave previous room if needed
    if (joinedRef.current && joinedRef.current !== roomId) {
      socketManager.emit('leave_room', joinedRef.current);
    }

    socketManager.emit('join_room', roomId);
    joinedRef.current = roomId;

    return () => {
      socketManager.emit('leave_room', roomId);
      joinedRef.current = null;
    };
  }, [roomId]);

  // Handle new messages via socket
  useEffect(() => {
    if (!roomId) return;

    const handleNew = (data: unknown) => {
      const msg = data as ChatMessage;
      if (msg.room_id !== roomId) return;

      queryClient.setQueryData(
        ['chat', 'messages', roomId],
        (old: ChatMessagesResponse | undefined) => {
          if (!old) return old;
          // Replace any optimistic messages from the same sender, add real message
          const filtered = old.data.filter(
            (m) => m.chat_id > 0 || m.sender.id !== msg.sender.id
          );
          // Avoid duplicate real message
          if (filtered.some((m) => m.chat_id === msg.chat_id)) return old;
          return {
            ...old,
            data: [msg, ...filtered],
            meta: { ...old.meta, total: old.meta.total + 1 },
          };
        }
      );
    };

    const handleEdited = (data: unknown) => {
      const payload = data as MessageEditedPayload;
      if (payload.room_id !== roomId) return;

      queryClient.setQueryData(
        ['chat', 'messages', roomId],
        (old: ChatMessagesResponse | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === payload.chat_id ? { ...m, message: payload.message } : m
            ),
          };
        }
      );
    };

    const handleDeleted = (data: unknown) => {
      const payload = data as { chat_id: number; room_id: number };
      if (payload.room_id !== roomId) return;

      queryClient.setQueryData(
        ['chat', 'messages', roomId],
        (old: ChatMessagesResponse | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === payload.chat_id
                ? { ...m, deleted_at: new Date().toISOString() }
                : m
            ),
          };
        }
      );
    };

    const unsubNew = socketManager.on('message:new', handleNew);
    const unsubEdited = socketManager.on('message:edited', handleEdited);
    const unsubDeleted = socketManager.on('message:deleted', handleDeleted);

    return () => {
      unsubNew();
      unsubEdited();
      unsubDeleted();
    };
  }, [roomId, queryClient]);

  return useQuery<ChatMessagesResponse>({
    queryKey: ['chat', 'messages', roomId],
    queryFn: () => getMessages(roomId!, token!),
    enabled: !!roomId && !!token,
    staleTime: 10_000,
  });
}
