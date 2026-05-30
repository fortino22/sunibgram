'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { reportTrend, type ReportTrendParams } from '@/lib/services/trendReports';
import type { TrendingHashtag } from '@/lib/services/hashtags';
import { useAuthStore } from '@/stores/authStore';

const TOP_KEY = ['hashtags', 'trending', 'top'] as const;
const FULL_KEY = ['hashtags', 'trending'] as const;

function stripHash(name: string) {
  return name.replace(/^#+/, '').trim().toLowerCase();
}

export function useReportTrend() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: ReportTrendParams) => reportTrend(params, token!),
    onMutate: async (params) => {
      const target = stripHash(params.hashtag_name);
      await Promise.all([
        queryClient.cancelQueries({ queryKey: TOP_KEY }),
        queryClient.cancelQueries({ queryKey: FULL_KEY }),
      ]);
      const prevTop = queryClient.getQueryData<TrendingHashtag[]>(TOP_KEY);
      const prevFull = queryClient.getQueryData<TrendingHashtag[]>(FULL_KEY);
      const filter = (list: TrendingHashtag[] | undefined) =>
        list?.filter((h) => stripHash(h.name) !== target);
      queryClient.setQueryData<TrendingHashtag[]>(TOP_KEY, (old) => filter(old) ?? []);
      queryClient.setQueryData<TrendingHashtag[]>(FULL_KEY, (old) => filter(old) ?? []);
      return { prevTop, prevFull };
    },
    onError: (err: Error, _params, ctx) => {
      if (ctx?.prevTop !== undefined) {
        queryClient.setQueryData(TOP_KEY, ctx.prevTop);
      }
      if (ctx?.prevFull !== undefined) {
        queryClient.setQueryData(FULL_KEY, ctx.prevFull);
      }
      toast.error(err.message || 'Failed to submit report.');
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: TOP_KEY });
      void queryClient.invalidateQueries({ queryKey: FULL_KEY });
    },
  });
}
