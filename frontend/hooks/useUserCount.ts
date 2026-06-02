'use client';

import { useQuery } from '@tanstack/react-query';
import { getUserCount } from '@/lib/services/users';

export function useUserCount() {
  return useQuery<number>({
    queryKey: ['users', 'count'],
    queryFn: getUserCount,
    staleTime: 5 * 60_000,
  });
}
