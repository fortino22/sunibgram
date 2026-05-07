'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createThread, type CreateThreadParams } from '@/lib/api/threads';
import type { ThreadResponse } from '@/lib/services/threads';

export function useCreateThread() {
  const queryClient = useQueryClient();

  return useMutation<ThreadResponse, Error, CreateThreadParams>({
    mutationFn: createThread,
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] }),
        queryClient.invalidateQueries({ queryKey: ['threads'] }),
      ]);
    },
  });
}
