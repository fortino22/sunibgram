'use client';

import { useQuery } from '@tanstack/react-query';
import {
  getCommunityEvents,
  getCommunityEvent,
  getCommunityEventComments,
} from '@/lib/services/communityEvents';
import { useAuthStore } from '@/stores/authStore';

export function useCommunityEvents(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community-events', communityId],
    queryFn: () => getCommunityEvents(communityId!, token!),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityEvent(eventId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community-event', eventId],
    queryFn: () => getCommunityEvent(eventId!, token!),
    enabled: !!token && !!eventId,
  });
}

export function useCommunityEventComments(eventId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community-event-comments', eventId],
    queryFn: () => getCommunityEventComments(eventId!, token!),
    enabled: !!token && !!eventId,
  });
}
