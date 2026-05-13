'use client';

import { useQuery } from '@tanstack/react-query';
import { getMessages, type ChatMessagesResponse } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useChatMessages(roomId: number | null) {
  const token = useAuthStore((s) => s.token);

  return useQuery<ChatMessagesResponse>({
    queryKey: ['chat', 'messages', roomId],
    queryFn: () => getMessages(roomId!, token!),
    enabled: !!roomId && !!token,
    staleTime: 10_000,
  });
}
