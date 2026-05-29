'use client';

import { useRef } from 'react';
import {
  useMutation,
  useQueryClient,
  type InfiniteData,
  type MutateOptions,
} from '@tanstack/react-query';
import { getFeedSince, type ThreadListResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

interface ShowNewPostsResult {
  fetched: number;
  prepended: number;
}

// Backs the "Show N new posts" banner. Fetches threads with thread_id > sinceId
// and prepends them onto the existing For You feed (page 1) — does NOT call
// /feed/refresh, so the seed stays put and older threads cannot resurface to
// re-trigger the banner. New-count is set to 0 synchronously so the banner
// disappears immediately on click.
export function useShowNewPosts() {
  const token = useAuthStore((state) => state.token);
  const queryClient = useQueryClient();
  // Synchronous in-flight guard. React Query's `isPending` updates via setState
  // and so lags one render — a rapid double-click in the same tick can pass an
  // `if (isPending) return` check twice and fire two mutations. A ref flips
  // immediately on enter and clears in onSettled.
  const inflightRef = useRef(false);

  const mutation = useMutation<ShowNewPostsResult, Error, number>({
    mutationFn: async (sinceId: number) => {
      if (!token) throw new Error('Authentication token is required.');
      const result = await getFeedSince(sinceId, token, 50);

      let prepended = 0;
      queryClient.setQueryData<InfiniteData<ThreadListResponse>>(
        ['threads', 'feed'],
        (old) => {
          if (!old || old.pages.length === 0) {
            return {
              pages: [{ data: result.data, meta: result.meta }],
              pageParams: [1],
            };
          }
          const firstPage = old.pages[0];
          const existingIds = new Set<number>();
          for (const page of old.pages) {
            for (const item of page.data) existingIds.add(item.id);
          }
          const fresh = result.data.filter((t) => !existingIds.has(t.id));
          prepended = fresh.length;
          if (fresh.length === 0) return old;
          const mergedFirst = {
            ...firstPage,
            data: [...fresh, ...firstPage.data],
          };
          return { ...old, pages: [mergedFirst, ...old.pages.slice(1)] };
        },
      );
      // Hide the banner immediately — banner is rendered off the cached count.
      queryClient.setQueryData(['threads', 'feed', 'new-count', sinceId], { count: 0 });
      return { fetched: result.data.length, prepended };
    },
    onError: (err) => {
      // Log to console; banner stays visible so user can retry. Without this,
      // a transient 5xx silently swallows the click and the user has no signal
      // that anything went wrong.
      console.error('[showNewPosts] failed to fetch new threads:', err);
    },
    onSettled: () => {
      inflightRef.current = false;
    },
  });

  return {
    ...mutation,
    mutate: (
      sinceId: number,
      options?: MutateOptions<ShowNewPostsResult, Error, number>,
    ) => {
      if (inflightRef.current) return;
      inflightRef.current = true;
      mutation.mutate(sinceId, options);
    },
  };
}
