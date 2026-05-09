'use client';

import { useQuery } from '@tanstack/react-query';
import { getProfile, type UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useProfile(userId: number | string) {
  const token = useAuthStore((s) => s.token);
  return useQuery<UserProfile>({
    queryKey: ['profile', userId],
    queryFn: () => getProfile(userId, token!),
    enabled: !!token && !!userId,
    staleTime: 30_000,
  });
}
