'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  deleteThread,
  updateThread,
  type ThreadMessageResponse,
  type ThreadMutationParams,
  type ThreadResponse,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { removeThreadFromCache } from '@/lib/threadCache';
import { INACTIVE_COMMUNITY_TOAST, isInactiveCommunityError } from '@/lib/api';
import { toast } from 'sonner';

interface UseThreadManagementOptions {
  onDeleteSuccess?: () => void;
}

function removeSingleThread(
  current: ThreadResponse | undefined,
  threadId: number
): ThreadResponse | undefined {
  if (current === undefined || current.thread_id !== threadId) {
    return current;
  }

  return undefined;
}

export function useThreadManagement(
  threadId: number,
  options: UseThreadManagementOptions = {}
) {
  const token = useAuthStore((state) => state.token);
  const queryClient = useQueryClient();

  const updateMutation = useMutation<ThreadMessageResponse, Error, ThreadMutationParams>({
    mutationFn: async (params) => {
      if (token === null) {
        throw new Error('Authentication token is required to edit this thread.');
      }

      return updateThread(threadId, params, token);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['threads'] }),
        queryClient.invalidateQueries({ queryKey: ['likes'] }),
        queryClient.invalidateQueries({ queryKey: ['thread'] }),
        queryClient.invalidateQueries({ queryKey: ['thread-root'] }),
        queryClient.invalidateQueries({ queryKey: ['thread-replies'] }),
      ]);
    },
    onError: (err) => {
      if (isInactiveCommunityError(err)) {
        toast.error(INACTIVE_COMMUNITY_TOAST);
        return;
      }
      toast.error(err.message || 'Failed to update thread.');
    },
  });

  const deleteMutation = useMutation<ThreadMessageResponse, Error, void>({
    mutationFn: async () => {
      if (token === null) {
        throw new Error('Authentication token is required to delete this thread.');
      }

      return deleteThread(threadId, token);
    },
    onSuccess: async () => {
      queryClient.setQueriesData({ queryKey: ['threads'] }, (current) => removeThreadFromCache(current, threadId));
      queryClient.setQueriesData({ queryKey: ['likes'] }, (current) => removeThreadFromCache(current, threadId));
      queryClient.setQueriesData({ queryKey: ['thread-replies'] }, (current) => removeThreadFromCache(current, threadId));
      queryClient.setQueriesData<ThreadResponse | undefined>(
        { queryKey: ['thread'] },
        (current) => removeSingleThread(current, threadId)
      );
      queryClient.setQueriesData({ queryKey: ['event'] }, (current) => removeThreadFromCache(current, threadId));

      options.onDeleteSuccess?.();

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['threads'] }),
        queryClient.invalidateQueries({ queryKey: ['likes'] }),
        queryClient.invalidateQueries({ queryKey: ['thread'] }),
        queryClient.invalidateQueries({ queryKey: ['thread-root'] }),
        queryClient.invalidateQueries({ queryKey: ['thread-replies'] }),
        queryClient.invalidateQueries({ queryKey: ['event'], refetchType: 'none' }),
      ]);
    },
    onError: (err) => {
      if (isInactiveCommunityError(err)) {
        toast.error(INACTIVE_COMMUNITY_TOAST);
        return;
      }
      toast.error(err.message || 'Failed to delete thread.');
    },
  });

  return {
    updateMutation,
    deleteMutation,
  };
}
