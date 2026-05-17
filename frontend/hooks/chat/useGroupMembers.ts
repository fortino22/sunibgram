'use client';

import { useQuery } from '@tanstack/react-query';
import { getGroupMembers, type GroupMember } from '@/lib/services/chat';
import { useAuthStore } from '@/stores/authStore';

export function useGroupMembers(roomId: number | null) {
  const token = useAuthStore((s) => s.token);

  return useQuery<GroupMember[]>({
    queryKey: ['chat', 'rooms', roomId, 'members'],
    queryFn: async () => {
      const res = await getGroupMembers(roomId!, token!);
      return res.data;
    },
    enabled: !!token && !!roomId,
    staleTime: 30_000,
  });
}
