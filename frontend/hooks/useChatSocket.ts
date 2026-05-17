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
import type { ChatMessage, ChatMessagesResponse, ChatRoom, ChatRoomsResponse, ChatReactionSummary } from '@/lib/services/chat';

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
  is_edited: boolean;
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

interface MessageReactionPayload {
  chat_id: number;
  room_id: number;
  user_id: number;
  emoji: string;
}

interface UserProfileUpdatedPayload {
  user_id: number;
  avatar_url: string;
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

// ── Reaction helpers (mirrors the optimistic helpers in useReactToMessage) ────

function upsertReaction(
  reactions: ChatReactionSummary[],
  emoji: string,
  userId: number,
): ChatReactionSummary[] {
  const existing = reactions.find((r) => r.emoji === emoji);
  if (existing) {
    if (existing.user_ids.includes(userId)) return reactions;
    return reactions.map((r) =>
      r.emoji === emoji
        ? { ...r, count: r.count + 1, user_ids: [...r.user_ids, userId] }
        : r,
    );
  }
  return [...reactions, { emoji, count: 1, user_ids: [userId] }];
}

function dropReaction(
  reactions: ChatReactionSummary[],
  emoji: string,
  userId: number,
): ChatReactionSummary[] {
  return reactions
    .map((r) =>
      r.emoji === emoji
        ? { ...r, count: r.count - 1, user_ids: r.user_ids.filter((id) => id !== userId) }
        : r,
    )
    .filter((r) => r.count > 0);
}

// ── Hook interface ────────────────────────────────────────────────────────────

export interface UseChatSocketOptions {
  accessToken: string | null;
  currentUserId?: number;
  activeRoomId?: number | null;
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
  activeRoomId,
  onSocketError,
}: UseChatSocketOptions): UseChatSocketResult {
  const queryClient = useQueryClient();
  const [typingUsersByRoom, setTypingUsersByRoom] = useState<Record<number, number[]>>({});

  const currentUserIdRef = useRef(currentUserId);
  currentUserIdRef.current = currentUserId;

  const activeRoomIdRef = useRef(activeRoomId);
  activeRoomIdRef.current = activeRoomId;

  // Ref so we can call sendJson from inside handleMessage without a circular dep
  const sendJsonRef = useRef<((msg: ClientEnvelope) => boolean) | null>(null);

  const handleMessage = useCallback(
    (message: ServerEnvelope): void => {
      switch (message.event) {
        // ── New message ────────────────────────────────────────────────────
        case 'message:new': {
          const raw = message.data as Record<string, unknown>;
          const msg: ChatMessage = {
            chat_id: raw.chat_id as number,
            room_id: raw.room_id as number,
            sender: raw.sender as ChatMessage['sender'],
            reply_to: raw.reply_to
              ? (raw.reply_to as ChatMessage['reply_to'])
              : undefined,
            message: (raw.message as string | undefined) ?? null,
            gif_url: (raw.gif_url as string | null | undefined) ?? null,
            media: (raw.media as ChatMessage['media']) ?? [],
            reactions: (raw.reactions as ChatReactionSummary[]) ?? [],
            is_read: (raw.is_read as boolean) ?? false,
            deleted_at: (raw.deleted_at as string | undefined) ?? null,
            created_at: raw.created_at as string,
          };

          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', msg.room_id],
            (old) => {
              if (!old) return old;
              const filtered = old.data.filter(
                (m) => m.chat_id > 0 || m.sender.id !== msg.sender.id,
              );
              if (filtered.some((m) => m.chat_id === msg.chat_id)) return old;
              const replacedOptimistic = filtered.length < old.data.length;
              return {
                ...old,
                data: [msg, ...filtered],
                meta: { ...old.meta, total: replacedOptimistic ? old.meta.total : old.meta.total + 1 },
              };
            },
          );

          // Update room list: move to top, update last_message, increment unread for non-active rooms
          queryClient.setQueryData<ChatRoomsResponse>(
            ['chat', 'rooms'],
            (old) => {
              if (!old) return old;
              const now = new Date().toISOString();
              const isFromOther = msg.sender.id !== currentUserIdRef.current;
              const isActiveRoom = msg.room_id === activeRoomIdRef.current;
              const updated = old.data.map((r) =>
                r.room_id === msg.room_id
                  ? {
                      ...r,
                      updated_at: now,
                      last_message: {
                        chat_id: msg.chat_id,
                        sender_id: msg.sender.id,
                        message: msg.message,
                        gif_url: msg.gif_url ?? null,
                        has_media: (msg.media?.length ?? 0) > 0,
                        created_at: msg.created_at,
                      },
                      unread_count: isFromOther && !isActiveRoom
                        ? (r.unread_count ?? 0) + 1
                        : (r.unread_count ?? 0),
                    }
                  : r,
              );
              updated.sort(
                (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
              );
              return { ...old, data: updated };
            },
          );
          break;
        }

        // ── New room (created by someone else or first DM) ─────────────────
        case 'room:new': {
          const room = message.data as ChatRoom;
          queryClient.setQueryData<ChatRoomsResponse>(
            ['chat', 'rooms'],
            (old) => {
              if (!old) return old;
              // Skip if already in the list
              if (old.data.some((r) => r.room_id === room.room_id)) return old;
              const updated = [room, ...old.data];
              return { ...old, data: updated, meta: { ...old.meta, total: old.meta.total + 1 } };
            },
          );
          // Auto-join the socket room so we receive future broadcasts for it
          sendJsonRef.current?.({ event: 'join_room', room_id: room.room_id } as ClientEnvelope);
          break;
        }

        // ── Avatar / profile updated ───────────────────────────────────────
        case 'user:profile_updated': {
          const payload = message.data as UserProfileUpdatedPayload;
          // Update avatar in all rooms where the updated user is the other participant
          queryClient.setQueryData<ChatRoomsResponse>(
            ['chat', 'rooms'],
            (old) => {
              if (!old) return old;
              return {
                ...old,
                data: old.data.map((r) =>
                  r.other_user?.user_id === payload.user_id
                    ? { ...r, other_user: { ...r.other_user!, avatar_url: payload.avatar_url } }
                    : r,
                ),
              };
            },
          );
          // Also invalidate the profile cache so ChatHeader and other components update
          void queryClient.invalidateQueries({ queryKey: ['profile', payload.user_id] });
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
                  m.chat_id === payload.chat_id
                    ? { ...m, message: payload.message, is_edited: true }
                    : m,
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

        // ── Read receipts ─────────────────────────────────────────────────
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

        // ── Reaction added ────────────────────────────────────────────────
        case 'message:reaction:added': {
          const payload = message.data as MessageReactionPayload;
          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', payload.room_id],
            (old) => {
              if (!old) return old;
              return {
                ...old,
                data: old.data.map((m) =>
                  m.chat_id === payload.chat_id
                    ? { ...m, reactions: upsertReaction(m.reactions ?? [], payload.emoji, payload.user_id) }
                    : m,
                ),
              };
            },
          );
          break;
        }

        // ── Reaction removed ──────────────────────────────────────────────
        case 'message:reaction:removed': {
          const payload = message.data as MessageReactionPayload;
          queryClient.setQueryData<ChatMessagesResponse>(
            ['chat', 'messages', payload.room_id],
            (old) => {
              if (!old) return old;
              return {
                ...old,
                data: old.data.map((m) =>
                  m.chat_id === payload.chat_id
                    ? { ...m, reactions: dropReaction(m.reactions ?? [], payload.emoji, payload.user_id) }
                    : m,
                ),
              };
            },
          );
          break;
        }

        // ── Typing indicators ─────────────────────────────────────────────
        case 'typing:start': {
          const payload = message.data as TypingPayload;
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

  // Keep ref in sync so handleMessage can emit join_room for room:new events
  sendJsonRef.current = socket.sendJson;

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
