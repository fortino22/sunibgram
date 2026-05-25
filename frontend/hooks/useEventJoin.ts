'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { joinEvent, leaveEvent, getMyJoinedEvents, type EventItem } from '@/lib/services/events';
import { useAuthStore } from '@/stores/authStore';

export function useEventJoin(eventId: string) {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  const joinedQuery = useQuery<EventItem[]>({
    queryKey: ['events', 'joined', 'me'],
    queryFn: () => getMyJoinedEvents(token!),
    enabled: !!token,
    staleTime: 30_000,
  });

  const isJoined = joinedQuery.data?.some((e) => e.id === eventId) ?? false;

  const join = useMutation({
    mutationFn: () => {
      if (!token) throw new Error('Not authenticated.');
      return joinEvent(eventId, token);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: ['events', 'joined', 'me'] }),
  });

  const leave = useMutation({
    mutationFn: () => {
      if (!token) throw new Error('Not authenticated.');
      return leaveEvent(eventId, token);
    },
    onMutate: () => {
      queryClient.setQueryData<EventItem[]>(['events', 'joined', 'me'], (old) =>
        old?.filter((e) => e.id !== eventId) ?? []
      );
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: ['events', 'joined', 'me'] }),
  });

  return {
    isJoined,
    isLoading: joinedQuery.isLoading,
    join,
    leave,
  };
}
