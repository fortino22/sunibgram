'use client';

import { useQuery } from '@tanstack/react-query';
import { getProfileByUserId, type UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useProfile(userId: number | string | null | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery<UserProfile>({
    queryKey: ['profile', userId],
    queryFn: () => {
      if (!token || userId === null || userId === undefined || userId === '') {
        throw new Error('A valid user id is required to load this profile.');
      }

      return getProfileByUserId(userId, token);
    },
    enabled: !!token && userId !== null && userId !== undefined && userId !== '',
    staleTime: 30_000,
  });
}
