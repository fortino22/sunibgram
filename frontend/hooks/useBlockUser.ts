'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { blockUser, unblockUser } from '@/lib/services/blocks';
import type { UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

function isProfileQueryKey(value: number | string | null | undefined): value is number | string {
  return value !== null && value !== undefined && value !== '';
}

export function useBlockUser(
  userId: number | string | null | undefined,
  profileQueryKeys: Array<number | string | null | undefined> = [userId]
) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();
  const resolvedProfileQueryKeys = profileQueryKeys.filter(isProfileQueryKey);

  function patchTarget(is_blocked: boolean) {
    resolvedProfileQueryKeys.forEach((profileQueryKey) => {
      queryClient.setQueryData<UserProfile>(['profile', profileQueryKey], (old) =>
        old ? { ...old, is_blocked, ...(is_blocked ? { is_following: false } : {}) } : old
      );
    });
  }

  function invalidateAll() {
    resolvedProfileQueryKeys.forEach((profileQueryKey) => {
      void queryClient.invalidateQueries({ queryKey: ['profile', profileQueryKey] });
      void queryClient.invalidateQueries({ queryKey: ['threads', profileQueryKey] });
    });
    void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    // Refresh feeds so the (un)blocked user's threads appear / disappear.
    void queryClient.invalidateQueries({ queryKey: ['threads'] });
    void queryClient.invalidateQueries({ queryKey: ['explore'] });
  }

  function ensureUserId(): number | string {
    if (!token || userId === null || userId === undefined || userId === '') {
      throw new Error('A valid user id is required.');
    }
    return userId;
  }

  const block = useMutation({
    mutationFn: () => blockUser(ensureUserId(), token!),
    onMutate: () => patchTarget(true),
    onError: (err: Error) => {
      patchTarget(false);
      toast.error(err.message || 'Failed to block user.');
    },
    onSuccess: () => {
      toast.success('User blocked.');
      invalidateAll();
    },
  });

  const unblock = useMutation({
    mutationFn: () => unblockUser(ensureUserId(), token!),
    onMutate: () => patchTarget(false),
    onError: (err: Error) => {
      patchTarget(true);
      toast.error(err.message || 'Failed to unblock user.');
    },
    onSuccess: () => {
      toast.success('User unblocked.');
      invalidateAll();
    },
  });

  return { block, unblock };
}
