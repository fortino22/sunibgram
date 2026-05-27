'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { refreshFeed } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

// Resets the For You feed seed, then refetches the feed and new-count queries.
// Invalidating ['threads', 'feed'] matches both the infinite feed and the
// ['threads', 'feed', 'new-count', sinceId] queries by prefix.
export function useRefreshFeed() {
  const token = useAuthStore((state) => state.token);
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => {
      if (!token) throw new Error('Authentication token is required to refresh the feed.');
      return refreshFeed(token);
    },
    // Refetch even on backend 5xx so the user still sees a fresh attempt.
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] });
    },
  });
}
