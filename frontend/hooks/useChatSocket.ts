'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import {
  useWebSocket,
  type ServerEnvelope,
  type SocketErrorPayload,
  type WebSocketStatus,
  type ClientEnvelope,
} from './useWebSocket';
import type { ChatMessage, ChatMessagesResponse } from '@/lib/services/chat';

// ── Payload types matching actual backend broadcasts ──────────────────────────

interface TypingPayload {
  room_id: number;
  user_id: number;
}

interface MessageEditedPayload {
  chat_id: number;
  room_id: number;
  sender_id: number;
  message: string;
}

interface MessageDeletedPayload {
  chat_id: number;
  room_id: number;
  sender_id: number;
  deleted_by: number;
}

interface MessageReadPayload {
  room_id: number;
  chat_ids: number[];
  read_by: number;
}

// ── Client → server message shapes ───────────────────────────────────────────

interface JoinRoomMessage extends ClientEnvelope {
  event: 'join_room';
  room_id: number;
}

interface LeaveRoomMessage extends ClientEnvelope {
  event: 'leave_room';
  room_id: number;
}

interface TypingStartMessage extends ClientEnvelope {
  event: 'typing:start';
  room_id: number;
}

interface TypingStopMessage extends ClientEnvelope {
  event: 'typing:stop';
  room_id: number;
}

// ── Hook interface ────────────────────────────────────────────────────────────

export interface UseChatSocketOptions {
  accessToken: string | null;
  currentUserId?: number;
  onSocketError?: (payload: SocketErrorPayload) => void;
}

export interface UseChatSocketResult {
  status: WebSocketStatus;
  lastError: string | null;
  typingUsersByRoom: Record<number, number[]>;
  joinRoom: (roomId: number) => boolean;
  leaveRoom: (roomId: number) => boolean;
  startTyping: (roomId: number) => boolean;
  stopTyping: (roomId: number) => boolean;
}

// ── Typing state helpers ──────────────────────────────────────────────────────

function addTypingUser(
  current: Record<number, number[]>,
  payload: TypingPayload,
): Record<number, number[]> {
  const existing = current[payload.room_id] ?? [];
  if (existing.includes(payload.user_id)) return current;
  return { ...current, [payload.room_id]: [...existing, payload.user_id] };
}

function removeTypingUser(
  current: Record<number, number[]>,
  payload: TypingPayload,
): Record<number, number[]> {
  const existing = current[payload.room_id] ?? [];
  return {
    ...current,
    [payload.room_id]: existing.filter((id) => id !== payload.user_id),
  };
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function useChatSocket({
  accessToken,
  currentUserId,
  onSocketError,
}: UseChatSocketOptions): UseChatSocketResult {
  const queryClient = useQueryClient();
  const [typingUsersByRoom, setTypingUsersByRoom] = useState<Record<number, number[]>>({});

  // Keep currentUserId in a ref so the message handler closure stays stable
  const currentUserIdRef = useRef(currentUserId);
  currentUserIdRef.current = currentUserId;

  const handleMessage = useCallback(
    (message: ServerEnvelope): void => {
      switch (message.event) {
        // ── New message ────────────────────────────────────────────────────
        case 'message:new': {
          const raw = message.data as Record<string, unknown>;
          // Normalise optional fields so ChatMessage types are satisfied
          const msg: ChatMessage = {
            chat_id: raw.chat_id as number,
            room_id: raw.room_id as number,
            sender: raw.sender as ChatMessage['sender'],
            message: (raw.message as string | undefined) ?? null,
            media: (raw.media as ChatMessage['media']) ?? [],
            is_read: (raw.is_read as boolean) ?? false,
            deleted_at: (raw.deleted_at as string | undefined) ?? null,
            created_at: raw.created_at as string,
          };

          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', msg.room_id],
            (old) => {
              if (!old) return old;
              // Replace any pending optimistic messages from the same sender
              const filtered = old.data.filter(
                (m) => m.chat_id > 0 || m.sender.id !== msg.sender.id,
              );
              // Deduplicate: skip if we already have this real message
              if (filtered.some((m) => m.chat_id === msg.chat_id)) return old;
              return {
                ...old,
                data: [msg, ...filtered],
                meta: { ...old.meta, total: old.meta.total + 1 },
              };
            },
          );

          // Refresh room list so updated_at order reflects the new message
          void queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
          break;
        }

        // ── Message edited ────────────────────────────────────────────────
        case 'message:edited': {
          const payload = message.data as MessageEditedPayload;
          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', payload.room_id],
            (old) => {
              if (!old) return old;
              return {
                ...old,
                data: old.data.map((m) =>
                  m.chat_id === payload.chat_id ? { ...m, message: payload.message } : m,
                ),
              };
            },
          );
          break;
        }

        // ── Message deleted ───────────────────────────────────────────────
        case 'message:deleted': {
          const payload = message.data as MessageDeletedPayload;
          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', payload.room_id],
            (old) => {
              if (!old) return old;
              return {
                ...old,
                data: old.data.map((m) =>
                  m.chat_id === payload.chat_id
                    ? { ...m, deleted_at: new Date().toISOString() }
                    : m,
                ),
              };
            },
          );
          break;
        }

        // ── Read receipts (sent only to the original message author) ──────
        case 'message:read': {
          const payload = message.data as MessageReadPayload;
          const chatIdSet = new Set(payload.chat_ids);
          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', payload.room_id],
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
          break;
        }

        // ── Typing indicators ─────────────────────────────────────────────
        case 'typing:start': {
          const payload = message.data as TypingPayload;
          // Suppress own typing events echoed back by the server
          if (payload.user_id === currentUserIdRef.current) break;
          setTypingUsersByRoom((curr) => addTypingUser(curr, payload));
          break;
        }

        case 'typing:stop': {
          const payload = message.data as TypingPayload;
          setTypingUsersByRoom((curr) => removeTypingUser(curr, payload));
          break;
        }
      }
    },
    [queryClient],
  );

  const socket = useWebSocket({
    accessToken,
    onMessage: handleMessage,
    onErrorMessage: onSocketError,
  });

  const joinRoom = useCallback(
    (roomId: number): boolean =>
      socket.sendJson<JoinRoomMessage>({ event: 'join_room', room_id: roomId }),
    [socket],
  );

  const leaveRoom = useCallback(
    (roomId: number): boolean =>
      socket.sendJson<LeaveRoomMessage>({ event: 'leave_room', room_id: roomId }),
    [socket],
  );

  const startTyping = useCallback(
    (roomId: number): boolean =>
      socket.sendJson<TypingStartMessage>({ event: 'typing:start', room_id: roomId }),
    [socket],
  );

  const stopTyping = useCallback(
    (roomId: number): boolean =>
      socket.sendJson<TypingStopMessage>({ event: 'typing:stop', room_id: roomId }),
    [socket],
  );

  return useMemo(
    () => ({
      status: socket.status,
      lastError: socket.lastError,
      typingUsersByRoom,
      joinRoom,
      leaveRoom,
      startTyping,
      stopTyping,
    }),
    [joinRoom, leaveRoom, socket.lastError, socket.status, startTyping, stopTyping, typingUsersByRoom],
  );
}
