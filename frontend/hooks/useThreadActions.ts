'use client';

import { useMutation, useQueryClient, type QueryKey } from '@tanstack/react-query';
import {
  likeThread, unlikeThread,
  bookmarkThread, unbookmarkThread,
  repostThread, unrepostThread,
  type ThreadListResponse, type ThreadItem,
} from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export function useThreadActions(threadId: number, queryKey: QueryKey) {
  const token = useAuthStore((s) => s.token!);
  const queryClient = useQueryClient();

  function patch(updater: (t: ThreadItem) => ThreadItem) {
    queryClient.setQueryData<ThreadListResponse>(queryKey, (old) => {
      if (!old) return old;
      return { ...old, data: old.data.map((t) => (t.id === threadId ? updater(t) : t)) };
    });
  }

  const like = useMutation({
    mutationFn: (isLiked: boolean) =>
      isLiked ? unlikeThread(threadId, token) : likeThread(threadId, token),
    onMutate: (isLiked) => {
      patch((t) => ({
        ...t,
        is_liked: !isLiked,
        stats: { ...t.stats, likes: t.stats.likes + (isLiked ? -1 : 1) },
      }));
    },
    onError: (_, isLiked) => {
      patch((t) => ({
        ...t,
        is_liked: isLiked,
        stats: { ...t.stats, likes: t.stats.likes + (isLiked ? 1 : -1) },
      }));
    },
  });

  const bookmark = useMutation({
    mutationFn: (isBookmarked: boolean) =>
      isBookmarked ? unbookmarkThread(threadId, token) : bookmarkThread(threadId, token),
    onMutate: (isBookmarked) => {
      patch((t) => ({
        ...t,
        is_bookmarked: !isBookmarked,
        stats: { ...t.stats, bookmarks: t.stats.bookmarks + (isBookmarked ? -1 : 1) },
      }));
    },
    onError: (_, isBookmarked) => {
      patch((t) => ({
        ...t,
        is_bookmarked: isBookmarked,
        stats: { ...t.stats, bookmarks: t.stats.bookmarks + (isBookmarked ? 1 : -1) },
      }));
    },
  });

  const repost = useMutation({
    mutationFn: (isReposted: boolean) =>
      isReposted ? unrepostThread(threadId, token) : repostThread(threadId, token),
    onMutate: (isReposted) => {
      patch((t) => ({
        ...t,
        is_reposted: !isReposted,
        stats: { ...t.stats, reposts: t.stats.reposts + (isReposted ? -1 : 1) },
      }));
    },
    onError: (_, isReposted) => {
      patch((t) => ({
        ...t,
        is_reposted: isReposted,
        stats: { ...t.stats, reposts: t.stats.reposts + (isReposted ? 1 : -1) },
      }));
    },
  });

  return { like, bookmark, repost };
}
