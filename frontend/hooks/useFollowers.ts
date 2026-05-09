'use client';

import { useQuery } from '@tanstack/react-query';
import { getFollowers, type FollowListResponse } from '@/lib/services/follows';
import { useAuthStore } from '@/stores/authStore';

export function useFollowers(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  return useQuery<FollowListResponse>({
    queryKey: ['followers', userId],
    queryFn: () => getFollowers(userId, token!),
    enabled: !!token && !!userId,
  });
}
