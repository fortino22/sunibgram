'use client';

import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { reportTrend, type ReportTrendParams } from '@/lib/services/trendReports';
import { useAuthStore } from '@/stores/authStore';

export function useReportTrend() {
  const token = useAuthStore((s) => s.token);
  return useMutation({
    mutationFn: (params: ReportTrendParams) => reportTrend(params, token!),
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to submit report.');
    },
  });
}
