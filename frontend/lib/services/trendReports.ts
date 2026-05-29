import { api } from '@/lib/api';

export type TrendReportReasonCode =
  | 'not_relevant'
  | 'spam'
  | 'abusive_or_harmful'
  | 'not_interested'
  | 'duplicate'
  | 'harmful_or_spammy';

export interface ReportTrendParams {
  hashtag_name: string;
  reason_code: TrendReportReasonCode;
}

export const reportTrend = (params: ReportTrendParams, token: string) =>
  api.post<{ message: string }>('/api/trend-reports', params, token);
