'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  createPrivateRoom,
  createGroupRoom,
  updateGroupRoom,
  addGroupMembers,
  leaveGroup,
} from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useCreatePrivateRoom() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (targetUserId: number) => createPrivateRoom(targetUserId, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
    },
  });
}

interface CreateGroupVars {
  name: string;
  memberIds: number[];
  description?: string;
}

export function useCreateGroupRoom() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ name, memberIds, description }: CreateGroupVars) =>
      createGroupRoom(name, memberIds, description, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
    },
  });
}

interface UpdateGroupVars {
  name: string;
  description?: string;
  avatar?: File | null;
}

export function useUpdateGroupRoom(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ name, description, avatar }: UpdateGroupVars) =>
      updateGroupRoom(roomId, name, description, avatar ?? null, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
    },
  });
}

export function useAddGroupMembers(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userIds: number[]) => addGroupMembers(roomId, userIds, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms', roomId, 'members'] });
    },
  });
}

export function useLeaveGroup(roomId: number) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => leaveGroup(roomId, token!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
    },
  });
}
