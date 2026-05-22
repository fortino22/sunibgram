'use client';

import { useQuery } from '@tanstack/react-query';
import { getFacultyThreads, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useFacultyThreads(
  facultyId: number | null | undefined,
  page = 1,
  limit = 10,
  enabled = true
) {
  const token = useAuthStore((s) => s.token);

  return useQuery<ThreadListResponse>({
    queryKey: ['threads', 'faculty', facultyId, page, limit],
    queryFn: () => {
      if (!token) throw new Error('Authentication required.');
      if (!facultyId) throw new Error('Faculty ID required.');
      return getFacultyThreads(facultyId, { page, limit }, token);
    },
    enabled: !!token && !!facultyId && enabled,
  });
}
