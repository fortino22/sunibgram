'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getBlockedUsers, unblockUser, type BlockedUser } from '@/lib/services/blocks';
import { useAuthStore } from '@/stores/authStore';

const BLOCKED_USERS_KEY = ['blocked-users', 'me'] as const;

export function useBlockedUsers() {
  const token = useAuthStore((s) => s.token);
  return useQuery<BlockedUser[]>({
    queryKey: BLOCKED_USERS_KEY,
    queryFn: () => getBlockedUsers(token!),
    enabled: !!token,
    staleTime: 30_000,
  });
}

export function useUnblock() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: number) => unblockUser(userId, token!),
    onMutate: async (userId: number) => {
      await queryClient.cancelQueries({ queryKey: BLOCKED_USERS_KEY });
      const previous = queryClient.getQueryData<BlockedUser[]>(BLOCKED_USERS_KEY);
      // Optimistically drop the row from the list.
      queryClient.setQueryData<BlockedUser[]>(BLOCKED_USERS_KEY, (old) =>
        old ? old.filter((u) => u.id !== userId) : old
      );
      return { previous };
    },
    onError: (err: Error, _userId, context) => {
      if (context?.previous) {
        queryClient.setQueryData(BLOCKED_USERS_KEY, context.previous);
      }
      toast.error(err.message || 'Failed to unblock user.');
    },
    onSuccess: (_data, userId) => {
      toast.success('User unblocked.');
      // Keep profile/feed views consistent with the unblock.
      void queryClient.invalidateQueries({ queryKey: ['profile', userId] });
      void queryClient.invalidateQueries({ queryKey: ['threads', userId] });
      void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
      void queryClient.invalidateQueries({ queryKey: ['threads'] });
      void queryClient.invalidateQueries({ queryKey: ['explore'] });
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: BLOCKED_USERS_KEY });
    },
  });
}
