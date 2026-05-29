'use client';

import { useMutation, useQueryClient, type InfiniteData } from '@tanstack/react-query';
import {
  deleteThread,
  updateThread,
  type ThreadListResponse,
  type ThreadMessageResponse,
  type ThreadMutationParams,
  type ThreadResponse,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { INACTIVE_COMMUNITY_TOAST, isInactiveCommunityError } from '@/lib/api';
import { toast } from 'sonner';

interface UseThreadManagementOptions {
  onDeleteSuccess?: () => void;
}

function removeThreadFromList(
  current: ThreadListResponse | undefined,
  threadId: number
): ThreadListResponse | undefined {
  if (current === undefined) {
    return current;
  }

  return {
    ...current,
    data: current.data.filter((item) => item.id !== threadId),
  };
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
      queryClient.setQueriesData<ThreadListResponse>(
        { queryKey: ['threads'] },
        (current) => removeThreadFromList(current, threadId)
      );
      queryClient.setQueriesData<ThreadListResponse>(
        { queryKey: ['likes'] },
        (current) => removeThreadFromList(current, threadId)
      );
      queryClient.setQueriesData<InfiniteData<ThreadListResponse>>(
        { queryKey: ['thread-replies'] },
        (current) => {
          if (!current) return current;
          return {
            ...current,
            pages: current.pages.map((page) => ({
              ...page,
              data: page.data.filter((item) => item.id !== threadId),
            })),
          };
        }
      );
      queryClient.setQueriesData<ThreadResponse | undefined>(
        { queryKey: ['thread'] },
        (current) => removeSingleThread(current, threadId)
      );
      queryClient.setQueriesData<InfiniteData<ThreadListResponse>>(
        { queryKey: ['event'] },
        (current) => {
          if (!current) return current;
          return {
            ...current,
            pages: current.pages.map((page) => ({
              ...page,
              data: page.data.filter((item) => item.id !== threadId),
            })),
          };
        }
      );

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
