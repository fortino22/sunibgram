'use client';

import { useQuery } from '@tanstack/react-query';
import { getFollowing, type FollowListResponse } from '@/lib/services/follows';
import { useAuthStore } from '@/stores/authStore';

export function useFollowing(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  return useQuery<FollowListResponse>({
    queryKey: ['following', userId],
    queryFn: () => getFollowing(userId, token!, { limit: 100 }),
    enabled: !!token && !!userId,
  });
}
