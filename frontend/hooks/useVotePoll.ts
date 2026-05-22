'use client';

import { useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import type { ThreadDetailData } from '@/hooks/useThread';
import {
  voteOnPoll,
  type PollResponse,
  type ThreadItem,
  type ThreadListResponse,
  type ThreadResponse,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

interface VotePollContext {
  snapshots: Array<[readonly unknown[], unknown]>;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isPollResponse(value: unknown): value is PollResponse {
  return (
    isRecord(value) &&
    typeof value.poll_id === 'number' &&
    typeof value.total_votes === 'number' &&
    Array.isArray(value.options)
  );
}

function isThreadItem(value: unknown): value is ThreadItem {
  return (
    isRecord(value) &&
    typeof value.id === 'number' &&
    typeof value.content === 'string' &&
    (value.poll === undefined || isPollResponse(value.poll))
  );
}

function isThreadResponse(value: unknown): value is ThreadResponse {
  return (
    isRecord(value) &&
    typeof value.thread_id === 'number' &&
    typeof value.content === 'string' &&
    (value.poll === undefined || isPollResponse(value.poll))
  );
}

function isThreadListResponse(value: unknown): value is ThreadListResponse {
  return (
    isRecord(value) &&
    Array.isArray(value.data) &&
    isRecord(value.meta) &&
    typeof value.meta.total === 'number'
  );
}

function isThreadDetailData(value: unknown): value is ThreadDetailData {
  return isRecord(value) && isThreadItem(value.root) && Array.isArray(value.replies);
}

function roundPercent(value: number): number {
  return Math.round(value * 10) / 10;
}

function applyPollVote(poll: PollResponse, optionId: number): PollResponse {
  const previousOptionId = poll.user_voted_option_id ?? null;
  let totalVotes = poll.total_votes;

  if (previousOptionId === null) {
    totalVotes += 1;
  }

  const options = poll.options.map((option) => {
    let voteCount = option.vote_count;

    if (option.option_id === optionId) {
      voteCount += 1;
    }

    if (previousOptionId !== null && previousOptionId !== optionId && option.option_id === previousOptionId) {
      voteCount = Math.max(0, voteCount - 1);
    }

    return {
      ...option,
      vote_count: voteCount,
    };
  }).map((option) => ({
    ...option,
    percent: totalVotes > 0 ? roundPercent((option.vote_count / totalVotes) * 100) : 0,
  }));

  return {
    ...poll,
    total_votes: totalVotes,
    user_voted_option_id: optionId,
    options,
  };
}

function patchThreadItem(thread: ThreadItem, pollId: number, optionId: number): ThreadItem {
  if (thread.poll?.poll_id !== pollId) {
    return thread;
  }

  return {
    ...thread,
    poll: applyPollVote(thread.poll, optionId),
  };
}

function patchThreadResponse(thread: ThreadResponse, pollId: number, optionId: number): ThreadResponse {
  if (thread.poll?.poll_id !== pollId) {
    return thread;
  }

  return {
    ...thread,
    poll: applyPollVote(thread.poll, optionId),
  };
}

function patchPollCaches(queryClient: QueryClient, pollId: number, optionId: number, communityId?: number) {
  queryClient.setQueriesData(
    { queryKey: ['threads'] },
    (current: unknown) =>
      isThreadListResponse(current)
        ? { ...current, data: current.data.map((thread) => patchThreadItem(thread, pollId, optionId)) }
        : current
  );
  queryClient.setQueriesData(
    { queryKey: ['likes'] },
    (current: unknown) =>
      isThreadListResponse(current)
        ? { ...current, data: current.data.map((thread) => patchThreadItem(thread, pollId, optionId)) }
        : current
  );
  queryClient.setQueriesData(
    { queryKey: ['bookmarks'] },
    (current: unknown) =>
      isThreadListResponse(current)
        ? { ...current, data: current.data.map((thread) => patchThreadItem(thread, pollId, optionId)) }
        : current
  );
  queryClient.setQueriesData(
    { queryKey: ['explore'] },
    (current: unknown) =>
      isThreadListResponse(current)
        ? { ...current, data: current.data.map((thread) => patchThreadItem(thread, pollId, optionId)) }
        : current
  );
  queryClient.setQueriesData(
    { queryKey: ['communities', 'threads'] },
    (current: unknown) =>
      isThreadListResponse(current)
        ? { ...current, data: current.data.map((thread) => patchThreadItem(thread, pollId, optionId)) }
        : current
  );
  if (communityId !== undefined) {
    queryClient.setQueriesData(
      { queryKey: ['community', communityId, 'threads'] },
      (current: unknown) =>
        isThreadListResponse(current)
          ? { ...current, data: current.data.map((thread) => patchThreadItem(thread, pollId, optionId)) }
          : current
    );
  }
  queryClient.setQueriesData(
    { queryKey: ['thread-chain'] },
    (current: unknown) =>
      isThreadDetailData(current)
        ? {
            ...current,
            root: patchThreadItem(current.root, pollId, optionId),
            replies: current.replies.map((reply) => patchThreadItem(reply, pollId, optionId)),
          }
        : current
  );
  queryClient.setQueriesData(
    { queryKey: ['thread'] },
    (current: unknown) => {
      if (isThreadItem(current)) {
        return patchThreadItem(current, pollId, optionId);
      }
      if (isThreadResponse(current)) {
        return patchThreadResponse(current, pollId, optionId);
      }
      return current;
    }
  );
}

function collectSnapshots(queryClient: QueryClient, communityId?: number): Array<[readonly unknown[], unknown]> {
  const keys: Array<readonly unknown[]> = [
    ['threads'],
    ['likes'],
    ['bookmarks'],
    ['explore'],
    ['communities', 'threads'],
    ['thread-chain'],
    ['thread'],
  ];

  if (communityId !== undefined) {
    keys.push(['community', communityId, 'threads']);
  }

  const deduped = new Map<string, [readonly unknown[], unknown]>();

  keys.forEach((queryKey) => {
    queryClient.getQueriesData({ queryKey }).forEach(([key, value]) => {
      deduped.set(JSON.stringify(key), [key, value]);
    });
  });

  return [...deduped.values()];
}

async function cancelPollQueries(queryClient: QueryClient, communityId?: number) {
  const tasks = [
    queryClient.cancelQueries({ queryKey: ['threads'] }),
    queryClient.cancelQueries({ queryKey: ['likes'] }),
    queryClient.cancelQueries({ queryKey: ['bookmarks'] }),
    queryClient.cancelQueries({ queryKey: ['explore'] }),
    queryClient.cancelQueries({ queryKey: ['communities', 'threads'] }),
    queryClient.cancelQueries({ queryKey: ['thread-chain'] }),
    queryClient.cancelQueries({ queryKey: ['thread'] }),
  ];

  if (communityId !== undefined) {
    tasks.push(queryClient.cancelQueries({ queryKey: ['community', communityId, 'threads'] }));
  }

  await Promise.all(tasks);
}

async function invalidatePollQueries(queryClient: QueryClient, communityId?: number) {
  const tasks = [
    queryClient.invalidateQueries({ queryKey: ['threads'] }),
    queryClient.invalidateQueries({ queryKey: ['likes'] }),
    queryClient.invalidateQueries({ queryKey: ['bookmarks'] }),
    queryClient.invalidateQueries({ queryKey: ['explore'] }),
    queryClient.invalidateQueries({ queryKey: ['communities', 'threads'] }),
    queryClient.invalidateQueries({ queryKey: ['thread-chain'] }),
    queryClient.invalidateQueries({ queryKey: ['thread'] }),
  ];

  if (communityId !== undefined) {
    tasks.push(queryClient.invalidateQueries({ queryKey: ['community', communityId, 'threads'] }));
  }

  await Promise.all(tasks);
}

export function useVotePoll(thread: ThreadItem) {
  const token = useAuthStore((state) => state.token);
  const queryClient = useQueryClient();
  const pollId = thread.poll?.poll_id;
  const communityId = thread.community?.id;

  return useMutation<void, Error, number, VotePollContext>({
    mutationFn: async (optionId) => {
      if (!token) {
        throw new Error('Authentication token is required to vote on this poll.');
      }

      await voteOnPoll(thread.id, optionId, token);
    },
    onMutate: async (optionId) => {
      if (pollId === undefined) {
        throw new Error('Poll not found.');
      }

      await cancelPollQueries(queryClient, communityId);
      const snapshots = collectSnapshots(queryClient, communityId);
      patchPollCaches(queryClient, pollId, optionId, communityId);
      return { snapshots };
    },
    onError: (_error, _optionId, context) => {
      context?.snapshots.forEach(([key, value]) => {
        queryClient.setQueryData(key, value);
      });
    },
    onSettled: async () => {
      await invalidatePollQueries(queryClient, communityId);
    },
  });
}
