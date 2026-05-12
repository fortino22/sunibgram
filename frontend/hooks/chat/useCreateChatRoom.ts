'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createPrivateRoom, createGroupRoom } from '@/lib/services/chat';
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
