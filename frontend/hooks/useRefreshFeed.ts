'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { refreshFeed } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

// Resets the For You feed seed, then restarts the feed and new-count queries.
// resetQueries (not invalidate) discards all loaded pages and refetches from
// page 1: with a freshly reshuffled seed the offset of page 2+ no longer lines
// up with page 1, so refetching old pages in place yields overlapping/missing
// threads (duplicate React keys → unclickable cards). The banner already
// scrolls to top, so a clean page-1 restart is the correct behaviour.
// Prefix ['threads', 'feed'] matches both the infinite feed and the
// ['threads', 'feed', 'new-count', sinceId] queries.
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
      void queryClient.resetQueries({ queryKey: ['threads', 'feed'] });
    },
  });
}
