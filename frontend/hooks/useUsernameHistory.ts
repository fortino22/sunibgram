'use client';

import { useQuery } from '@tanstack/react-query';
import { getUsernameHistory, type UsernameHistory } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useUsernameHistory(userId: number | string | null | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery<UsernameHistory>({
    queryKey: ['username-history', userId],
    queryFn: () => {
      if (!token || userId === null || userId === undefined || userId === '') {
        throw new Error('A valid user id is required to load username history.');
      }
      return getUsernameHistory(userId, token);
    },
    enabled: !!token && userId !== null && userId !== undefined && userId !== '',
    staleTime: 30_000,
  });
}
