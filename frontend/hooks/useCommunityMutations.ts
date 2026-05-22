'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  joinCommunity,
  leaveCommunity,
  requestToJoinCommunity,
  cancelCommunityJoinRequest,
  approveCommunityJoinRequest,
  declineCommunityJoinRequest,
  kickCommunityMember,
  updateCommunityMemberRole,
  createCommunity,
  updateCommunity,
  deleteCommunity,
  createCommunityThread,
  pinCommunity,
  unpinCommunity,
  pinCommunityThread,
  unpinCommunityThread,
  type CommunityParams,
  type UpdateCommunityParams,
  type CommunityRole,
  type CommunityThreadParams,
} from '@/lib/services/communities';
import { useAuthStore } from '@/stores/authStore';

export function useJoinCommunity(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => joinCommunity(communityId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId] });
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'members'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'joined'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'threads', 'feed'] });
    },
  });
}

export function useLeaveCommunity(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => leaveCommunity(communityId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId] });
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'members'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'joined'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'threads', 'feed'] });
    },
  });
}

export function useRequestToJoin(communityId: number) {
  const token = useAuthStore((s) => s.token);
  return useMutation({
    mutationFn: () => requestToJoinCommunity(communityId, token!),
  });
}

export function useCancelJoinRequest(communityId: number) {
  const token = useAuthStore((s) => s.token);
  return useMutation({
    mutationFn: () => cancelCommunityJoinRequest(communityId, token!),
  });
}

export function useApproveRequest(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: number) => approveCommunityJoinRequest(communityId, userId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'requests'] });
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'members'] });
    },
  });
}

export function useDeclineRequest(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: number) => declineCommunityJoinRequest(communityId, userId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'requests'] });
    },
  });
}

export function useKickMember(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: number) => kickCommunityMember(communityId, userId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'members'] });
      void qc.invalidateQueries({ queryKey: ['community', communityId] });
    },
  });
}

export function useUpdateMemberRole(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: CommunityRole }) =>
      updateCommunityMemberRole(communityId, userId, role, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'members'] });
    },
  });
}

export function useCreateCommunity() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (params: CommunityParams) => createCommunity(params, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['communities', 'mine'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'discover'] });
    },
  });
}

export function useUpdateCommunity(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (params: UpdateCommunityParams) => updateCommunity(communityId, params, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId] });
      void qc.invalidateQueries({ queryKey: ['communities', 'mine'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'discover'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'joined'] });
    },
  });
}

export function useDeleteCommunity() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (communityId: number) => deleteCommunity(communityId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['communities', 'mine'] });
      void qc.invalidateQueries({ queryKey: ['communities', 'discover'] });
    },
  });
}

export function useCreateCommunityThread(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (params: CommunityThreadParams) => createCommunityThread(communityId, params, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'threads'] });
    },
  });
}

export function usePinCommunity(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => pinCommunity(communityId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['communities', 'pins'] });
    },
  });
}

export function useUnpinCommunity(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => unpinCommunity(communityId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['communities', 'pins'] });
    },
  });
}

export function usePinCommunityThread(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (threadId: number) => pinCommunityThread(communityId, threadId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'threads'] });
    },
  });
}

export function useUnpinCommunityThread(communityId: number) {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (threadId: number) => unpinCommunityThread(communityId, threadId, token!),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['community', communityId, 'threads'] });
    },
  });
}
