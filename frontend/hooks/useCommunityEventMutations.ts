'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  createCommunityEvent,
  updateCommunityEvent,
  deleteCommunityEvent,
  addCommunityEventComment,
  editCommunityEventComment,
  deleteCommunityEventComment,
  pinCommunityEventComment,
  unpinCommunityEventComment,
  setCommunityEventReaction,
  removeCommunityEventReaction,
  type CommunityEventParams,
} from '@/lib/services/communityEvents';
import { useAuthStore } from '@/stores/authStore';

export function useCreateCommunityEvent(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (params: CommunityEventParams) => createCommunityEvent(communityId, params, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-events', communityId] });
    },
  });
}

export function useUpdateCommunityEvent(eventId: number, communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (params: CommunityEventParams) => updateCommunityEvent(eventId, params, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
      qc.invalidateQueries({ queryKey: ['community-events', communityId] });
    },
  });
}

export function useDeleteCommunityEvent(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (eventId: number) => deleteCommunityEvent(eventId, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-events', communityId] });
    },
  });
}

export function useAddCommunityEventComment(eventId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (content: string) => addCommunityEventComment(eventId, content, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event-comments', eventId] });
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
    },
  });
}

export function useEditCommunityEventComment(eventId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ commentId, content }: { commentId: number; content: string }) =>
      editCommunityEventComment(eventId, commentId, content, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event-comments', eventId] });
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
    },
  });
}

export function useDeleteCommunityEventComment(eventId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (commentId: number) => deleteCommunityEventComment(eventId, commentId, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event-comments', eventId] });
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
    },
  });
}

export function usePinCommunityEventComment(eventId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ commentId, pinned }: { commentId: number; pinned: boolean }) =>
      pinned
        ? unpinCommunityEventComment(eventId, commentId, token!)
        : pinCommunityEventComment(eventId, commentId, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event-comments', eventId] });
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
    },
  });
}

export function useSetCommunityEventReaction(eventId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (emoji: string) => setCommunityEventReaction(eventId, emoji, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
    },
  });
}

export function useRemoveCommunityEventReaction(eventId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => removeCommunityEventReaction(eventId, token!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['community-event', eventId] });
    },
  });
}
