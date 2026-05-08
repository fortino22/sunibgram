'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyProfile, type UserProfile } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useMyProfile() {
  const token = useAuthStore((s) => s.token);
  return useQuery<UserProfile>({
    queryKey: ['profile', 'me'],
    queryFn: () => getMyProfile(token!),
    enabled: !!token,
    staleTime: 60_000,
  });
}
