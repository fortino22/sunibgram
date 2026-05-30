'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { searchShareUsers, type ShareUser } from '@/lib/services/share';

export function useDebouncedValue<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

export function useSearchShareUsers(query: string) {
  const token = useAuthStore((s) => s.token);
  const debounced = useDebouncedValue(query.trim(), 300);
  return useQuery<ShareUser[]>({
    queryKey: ['share-search', debounced],
    queryFn: () => {
      if (!token) return Promise.resolve([]);
      return searchShareUsers(debounced, token);
    },
    enabled: debounced.length > 0 && !!token,
    staleTime: 30_000,
    placeholderData: (prev) => prev,
  });
}
