'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getMutedUsers, unmuteUser, type MutedUser } from '@/lib/services/mutes';
import { useAuthStore } from '@/stores/authStore';

const MUTED_USERS_KEY = ['muted-users', 'me'] as const;

export function useMutedUsers() {
  const token = useAuthStore((s) => s.token);
  return useQuery<MutedUser[]>({
    queryKey: MUTED_USERS_KEY,
    queryFn: () => getMutedUsers(token!),
    enabled: !!token,
    staleTime: 30_000,
  });
}

export function useUnmute() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: number) => unmuteUser(userId, token!),
    onMutate: async (userId: number) => {
      await queryClient.cancelQueries({ queryKey: MUTED_USERS_KEY });
      const previous = queryClient.getQueryData<MutedUser[]>(MUTED_USERS_KEY);
      // Optimistically drop the row from the list.
      queryClient.setQueryData<MutedUser[]>(MUTED_USERS_KEY, (old) =>
        old ? old.filter((u) => u.id !== userId) : old
      );
      return { previous };
    },
    onError: (err: Error, _userId, context) => {
      if (context?.previous) {
        queryClient.setQueryData(MUTED_USERS_KEY, context.previous);
      }
      toast.error(err.message || 'Failed to unmute user.');
    },
    onSuccess: (_data, userId) => {
      toast.success('User unmuted.');
      // Keep profile/feed views consistent with the unmute.
      void queryClient.invalidateQueries({ queryKey: ['profile', userId] });
      void queryClient.invalidateQueries({ queryKey: ['threads', userId] });
      void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
      void queryClient.invalidateQueries({ queryKey: ['threads'] });
      void queryClient.invalidateQueries({ queryKey: ['explore'] });
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: MUTED_USERS_KEY });
    },
  });
}
