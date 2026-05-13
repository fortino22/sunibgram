'use client';

import { useQuery } from '@tanstack/react-query';
import { getMyRooms, type ChatRoom, type ChatRoomsResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useChatRooms() {
  const token = useAuthStore((s) => s.token);

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
