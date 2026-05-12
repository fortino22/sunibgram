'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { followUser, unfollowUser, type UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useFollowUser(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  const myId = useAuthStore((s) => s.user?.id);
  const queryClient = useQueryClient();

  function patchTarget(is_following: boolean, delta: number) {
    queryClient.setQueryData<UserProfile>(['profile', userId], (old) =>
      old ? { ...old, is_following, followers_count: Math.max(0, (old.followers_count ?? 0) + delta) } : old
    );
  }

  function patchMe(delta: number) {
    queryClient.setQueryData<UserProfile>(['profile', 'me'], (old) =>
      old ? { ...old, following_count: Math.max(0, (old.following_count ?? 0) + delta) } : old
    );
  }

  function invalidateAll() {
    void queryClient.invalidateQueries({ queryKey: ['profile', userId] });
    void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    void queryClient.invalidateQueries({ queryKey: ['followers', userId] });
    if (myId) void queryClient.invalidateQueries({ queryKey: ['following', myId] });
  }

  const follow = useMutation({
    mutationFn: () => followUser(userId, token!),
    onMutate: () => { patchTarget(true, 1); patchMe(1); },
    onError: () => { patchTarget(false, -1); patchMe(-1); },
    onSuccess: invalidateAll,
  });

  const unfollow = useMutation({
    mutationFn: () => unfollowUser(userId, token!),
    onMutate: () => { patchTarget(false, -1); patchMe(-1); },
    onError: () => { patchTarget(true, 1); patchMe(1); },
    onSuccess: invalidateAll,
  });

  return { follow, unfollow };
}
