'use client';

import { useMutation, useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getViolations } from '@/lib/services/violations';
import { reportPost, type ReportPostParams } from '@/lib/services/reports';
import { useAuthStore } from '@/stores/authStore';

export function useViolations() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['violations'],
    queryFn: () => getViolations(token!),
    enabled: !!token,
    staleTime: 10 * 60 * 1000,
  });
}

export function useReportPost() {
  const token = useAuthStore((s) => s.token);
  return useMutation({
    mutationFn: (params: ReportPostParams) => reportPost(params, token!),
    onSuccess: () => {
      toast.success('Report submitted. Our team will review this post.');
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to submit report.');
    },
  });
}
