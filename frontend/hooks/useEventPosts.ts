'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import {
  getEventDiscussion,
  getOfficialEventPosts,
} from '@/lib/services/events';
import type { PaginatedFeedResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

type EventPostMode = 'official' | 'discussion';

const EVENT_POST_PAGE_SIZE = 10;

export function useEventPosts(eventId: string, mode: EventPostMode, enabled = true) {
  const token = useAuthStore((state) => state.token);

  return useInfiniteQuery<PaginatedFeedResponse>({
    queryKey: ['event', eventId, mode, EVENT_POST_PAGE_SIZE],
    queryFn: ({ pageParam }) => {
      if (token === null) {
        throw new Error('Authentication token is required to fetch event posts.');
      }

      const page = typeof pageParam === 'number' ? pageParam : 1;
      const params = { page, limit: EVENT_POST_PAGE_SIZE };

      return mode === 'official'
        ? getOfficialEventPosts(eventId, params, token)
        : getEventDiscussion(eventId, params, token);
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loadedCount = allPages.reduce((sum, page) => sum + page.data.length, 0);
      if (loadedCount >= lastPage.meta.total) {
        return undefined;
      }
      return allPages.length + 1;
    },
    enabled: token !== null && enabled && eventId.length > 0,
  });
}
