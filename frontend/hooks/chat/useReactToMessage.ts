'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addReaction, removeReaction, type ChatMessagesResponse, type ChatReactionSummary } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

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

function swapReaction(
  reactions: ChatReactionSummary[],
  oldEmoji: string,
  newEmoji: string,
  userId: number,
): ChatReactionSummary[] {
  return upsertReaction(dropReaction(reactions, oldEmoji, userId), newEmoji, userId);
}

export function useAddReaction(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ chatId, emoji }: { chatId: number; emoji: string }) =>
      addReaction(chatId, emoji, token!),

    onMutate: async ({ chatId, emoji }) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });
      queryClient.setQueryData<ChatMessagesResponse>(
        ['chat', 'messages', roomId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === chatId
                ? { ...m, reactions: upsertReaction(m.reactions ?? [], emoji, user!.id) }
                : m,
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

export function useRemoveReaction(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ chatId, emoji }: { chatId: number; emoji: string }) =>
      removeReaction(chatId, emoji, token!),

    onMutate: async ({ chatId, emoji }) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });
      queryClient.setQueryData<ChatMessagesResponse>(
        ['chat', 'messages', roomId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === chatId
                ? { ...m, reactions: dropReaction(m.reactions ?? [], emoji, user!.id) }
                : m,
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

// Atomically replaces a user's existing reaction with a new emoji.
export function useSwapReaction(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ chatId, oldEmoji, newEmoji }: { chatId: number; oldEmoji: string; newEmoji: string }) => {
      await removeReaction(chatId, oldEmoji, token!);
      await addReaction(chatId, newEmoji, token!);
    },

    onMutate: async ({ chatId, oldEmoji, newEmoji }) => {
      await queryClient.cancelQueries({ queryKey: ['chat', 'messages', roomId] });
      queryClient.setQueryData<ChatMessagesResponse>(
        ['chat', 'messages', roomId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((m) =>
              m.chat_id === chatId
                ? { ...m, reactions: swapReaction(m.reactions ?? [], oldEmoji, newEmoji, user!.id) }
                : m,
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
