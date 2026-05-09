'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { followUser, unfollowUser, type UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useFollowUser(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  function patch(is_following: boolean, delta: number) {
    queryClient.setQueryData<UserProfile>(['profile', userId], (old) =>
      old
        ? { ...old, is_following, followers_count: Math.max(0, (old.followers_count ?? 0) + delta) }
        : old
    );
  }

  const follow = useMutation({
    mutationFn: () => followUser(userId, token!),
    onMutate: () => patch(true, 1),
    onError: () => patch(false, -1),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['profile', userId] }),
  });

  const unfollow = useMutation({
    mutationFn: () => unfollowUser(userId, token!),
    onMutate: () => patch(false, -1),
    onError: () => patch(true, 1),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['profile', userId] }),
  });

  return { follow, unfollow };
}
