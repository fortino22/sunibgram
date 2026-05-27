'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import { getMutualSuggestions, getFacultySuggestions } from '@/lib/services/explore';
import { useAuthStore } from '@/stores/authStore';

const PAGE_SIZE = 10;

export function useMutualSuggestions(enabled = true) {
  const token = useAuthStore((s) => s.token);
  return useInfiniteQuery({
    queryKey: ['suggestions', 'mutual'],
    queryFn: ({ pageParam }) => getMutualSuggestions(token!, pageParam as number, PAGE_SIZE),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.length === PAGE_SIZE ? allPages.length + 1 : undefined,
    enabled: token !== null && enabled,
  });
}

export function useFacultySuggestions(enabled = true) {
  const token = useAuthStore((s) => s.token);
  return useInfiniteQuery({
    queryKey: ['suggestions', 'faculty'],
    queryFn: ({ pageParam }) => getFacultySuggestions(token!, pageParam as number, PAGE_SIZE),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.length === PAGE_SIZE ? allPages.length + 1 : undefined,
    enabled: token !== null && enabled,
  });
}
