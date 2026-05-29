import { api } from '../../api';
import type { TrendReason, TrendReasonStats } from './types';

interface ListReasons {
  data: TrendReason[];
}

interface StatsResponse {
  data: TrendReasonStats[];
}

export async function listTrendReasons(token: string): Promise<TrendReason[]> {
  const res = await api.get<ListReasons>('/api/trend-reports/reasons', token);
  return res.data ?? [];
}

export async function getTrendStats(token: string): Promise<TrendReasonStats[]> {
  const res = await api.get<StatsResponse>('/api/trend-reports/stats', token);
  return res.data ?? [];
}
