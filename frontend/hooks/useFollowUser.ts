'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { followUser, unfollowUser } from '@/lib/services/follows';
import type { UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

function isProfileQueryKey(value: number | string | null | undefined): value is number | string {
  return value !== null && value !== undefined && value !== '';
}

export function useFollowUser(
  userId: number | string | null | undefined,
  profileQueryKeys: Array<number | string | null | undefined> = [userId]
) {
  const token = useAuthStore((s) => s.token);
  const myId = useAuthStore((s) => s.user?.id);
  const queryClient = useQueryClient();
  const resolvedProfileQueryKeys = profileQueryKeys.filter(isProfileQueryKey);

  function patchTarget(is_following: boolean, delta: number) {
    resolvedProfileQueryKeys.forEach((profileQueryKey) => {
      queryClient.setQueryData<UserProfile>(['profile', profileQueryKey], (old) =>
        old ? { ...old, is_following, followers_count: Math.max(0, (old.followers_count ?? 0) + delta) } : old
      );
    });
  }

  function patchMe(delta: number) {
    queryClient.setQueryData<UserProfile>(['profile', 'me'], (old) =>
      old ? { ...old, following_count: Math.max(0, (old.following_count ?? 0) + delta) } : old
    );
  }

  function invalidateAll() {
    resolvedProfileQueryKeys.forEach((profileQueryKey) => {
      void queryClient.invalidateQueries({ queryKey: ['profile', profileQueryKey] });
    });
    void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    void queryClient.invalidateQueries({ queryKey: ['followers', userId] });
    if (myId) void queryClient.invalidateQueries({ queryKey: ['following', myId] });
  }

  const follow = useMutation({
    mutationFn: () => {
      if (!token || userId === null || userId === undefined || userId === '') {
        throw new Error('A valid user id is required to follow this user.');
      }

      return followUser(userId, token);
    },
    onMutate: () => { patchTarget(true, 1); patchMe(1); },
    onError: () => { patchTarget(false, -1); patchMe(-1); },
    onSuccess: invalidateAll,
  });

  const unfollow = useMutation({
    mutationFn: () => {
      if (!token || userId === null || userId === undefined || userId === '') {
        throw new Error('A valid user id is required to unfollow this user.');
      }

      return unfollowUser(userId, token);
    },
    onMutate: () => { patchTarget(false, -1); patchMe(-1); },
    onError: () => { patchTarget(true, 1); patchMe(1); },
    onSuccess: invalidateAll,
  });

  return { follow, unfollow };
}
