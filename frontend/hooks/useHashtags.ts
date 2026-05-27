'use client';

import { useQuery } from '@tanstack/react-query';
import { getTrendingHashtags, getTopTrendingHashtags, type TrendingHashtag } from '@/lib/services/hashtags';
import { useAuthStore } from '@/stores/authStore';

export function useTrendingHashtags(limit?: number) {
  const token = useAuthStore((state) => state.token);

  return useQuery<TrendingHashtag[]>({
    queryKey: ['hashtags', 'trending'],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch trending hashtags.');
      }

      return getTrendingHashtags(token);
    },
    enabled: token !== null,
    staleTime: 5 * 60 * 1000,
    select: (hashtags) => (typeof limit === 'number' ? hashtags.slice(0, limit) : hashtags),
  });
}

/** Compact top trending list (right panel / explore "For You" widget). */
export function useTopTrendingHashtags() {
  const token = useAuthStore((state) => state.token);

  return useQuery<TrendingHashtag[]>({
    queryKey: ['hashtags', 'trending', 'top'],
    queryFn: () => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch trending hashtags.');
      }

      return getTopTrendingHashtags(token);
    },
    enabled: token !== null,
    staleTime: 5 * 60 * 1000,
  });
}
