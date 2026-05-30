'use client';

import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import {
  createFeedback,
  deleteFeedback,
  dislikeFeedback,
  getFeedback,
  likeFeedback,
  listFeedback,
  updateFeedbackStatus,
  type CreateFeedbackInput,
  type FeedbackResponse,
  type FeedbackStatus,
  type FeedbackVote,
  type ListFeedbackParams,
  type PaginatedFeedbackResponse,
} from '@/lib/services/feedback';
import { useAuthStore } from '@/stores/authStore';

function isPaginatedFeedback(value: unknown): value is PaginatedFeedbackResponse {
  return (
    typeof value === 'object' &&
    value !== null &&
    Array.isArray((value as PaginatedFeedbackResponse).data) &&
    typeof (value as PaginatedFeedbackResponse).meta === 'object'
  );
}

function applyVote(
  current: FeedbackResponse,
  nextVote: FeedbackVote | null,
): FeedbackResponse {
  let likes = current.likes;
  let dislikes = current.dislikes;

  if (current.is_vote === 'like') likes = Math.max(0, likes - 1);
  if (current.is_vote === 'dislike') dislikes = Math.max(0, dislikes - 1);
  if (nextVote === 'like') likes += 1;
  if (nextVote === 'dislike') dislikes += 1;

  return { ...current, is_vote: nextVote, likes, dislikes };
}

function patchFeedbackLists(
  queryClient: QueryClient,
  id: string,
  updater: (item: FeedbackResponse) => FeedbackResponse,
) {
  queryClient.setQueriesData({ queryKey: ['feedback', 'list'] }, (current: unknown) => {
    if (!isPaginatedFeedback(current)) return current;
    return {
      ...current,
      data: current.data.map((item) => (item.id === id ? updater(item) : item)),
    };
  });

  queryClient.setQueriesData({ queryKey: ['feedback', 'item', id] }, (current: unknown) => {
    if (!current || typeof current !== 'object') return current;
    const item = current as FeedbackResponse;
    if (item.id !== id) return current;
    return updater(item);
  });
}

function removeFromFeedbackLists(queryClient: QueryClient, id: string) {
  queryClient.setQueriesData({ queryKey: ['feedback', 'list'] }, (current: unknown) => {
    if (!isPaginatedFeedback(current)) return current;
    const nextData = current.data.filter((item) => item.id !== id);
    if (nextData.length === current.data.length) return current;
    return {
      ...current,
      data: nextData,
      meta: { ...current.meta, total: Math.max(0, current.meta.total - 1) },
    };
  });
}

export function useFeedbackList(params: ListFeedbackParams) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['feedback', 'list', params],
    queryFn: () => listFeedback(params, token!),
    enabled: !!token,
    staleTime: 30 * 1000,
  });
}

export function useFeedbackItem(id: string | null | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['feedback', 'item', id ?? ''],
    queryFn: () => getFeedback(id!, token!),
    enabled: !!token && !!id,
    staleTime: 30 * 1000,
  });
}

export function useCreateFeedback() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateFeedbackInput) => createFeedback(input, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feedback', 'list'] });
    },
  });
}

export function useDeleteFeedback() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteFeedback(id, token!),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ['feedback', 'list'] });
      const snapshot = queryClient.getQueriesData({ queryKey: ['feedback', 'list'] });
      removeFromFeedbackLists(queryClient, id);
      return { snapshot };
    },
    onError: (_err, _id, context) => {
      context?.snapshot.forEach(([key, value]) => queryClient.setQueryData(key, value));
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['feedback', 'list'] });
    },
  });
}

interface VoteArgs {
  id: string;
  current: FeedbackVote | null;
  next: FeedbackVote | null;
}

export function useVoteFeedback() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, next }: VoteArgs) => {
      if (next === 'like') return likeFeedback(id, token!);
      if (next === 'dislike') return dislikeFeedback(id, token!);
      throw new Error('Cannot send empty vote — backend has no unvote endpoint.');
    },
    onMutate: async ({ id, next }) => {
      await queryClient.cancelQueries({ queryKey: ['feedback'] });
      const snapshot = [
        ...queryClient.getQueriesData({ queryKey: ['feedback', 'list'] }),
        ...queryClient.getQueriesData({ queryKey: ['feedback', 'item', id] }),
      ];
      patchFeedbackLists(queryClient, id, (item) => applyVote(item, next));
      return { snapshot };
    },
    onError: (_err, _vars, context) => {
      context?.snapshot.forEach(([key, value]) => queryClient.setQueryData(key, value));
    },
  });
}

interface UpdateStatusArgs {
  id: string;
  status: FeedbackStatus;
}

export function useUpdateFeedbackStatus() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: UpdateStatusArgs) => updateFeedbackStatus(id, status, token!),
    onMutate: async ({ id, status }) => {
      await queryClient.cancelQueries({ queryKey: ['feedback'] });
      const snapshot = [
        ...queryClient.getQueriesData({ queryKey: ['feedback', 'list'] }),
        ...queryClient.getQueriesData({ queryKey: ['feedback', 'item', id] }),
      ];
      patchFeedbackLists(queryClient, id, (item) => ({ ...item, status }));
      return { snapshot };
    },
    onError: (_err, _vars, context) => {
      context?.snapshot.forEach(([key, value]) => queryClient.setQueryData(key, value));
    },
    onSettled: (_data, _err, vars) => {
      queryClient.invalidateQueries({ queryKey: ['feedback', 'item', vars.id] });
    },
  });
}
