'use client';

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getMyRooms, type ChatRoom, type ChatRoomsResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';
import { useEffect } from 'react';
import { socketManager } from '@/lib/socket';

export function useChatRooms() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  // Refresh room list when a new message arrives in any room (updates updated_at ordering)
  useEffect(() => {
    const handleNewMessage = () => {
      queryClient.invalidateQueries({ queryKey: ['chat', 'rooms'] });
    };
    const unsub = socketManager.on('message:new', handleNewMessage);
    return unsub;
  }, [queryClient]);

  return useQuery<ChatRoomsResponse>({
    queryKey: ['chat', 'rooms'],
    queryFn: () => getMyRooms(token!),
    enabled: !!token,
    staleTime: 30_000,
  });
}

export function useChatRoom(roomId: number | null): ChatRoom | undefined {
  const { data } = useChatRooms();
  if (!roomId || !data) return undefined;
  return data.data.find((r) => r.room_id === roomId);
}
