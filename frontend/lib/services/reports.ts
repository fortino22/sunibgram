import { api } from '@/lib/api';

export interface ReportPostParams {
  violation_id: string;
  reported_post_id: number;
  explanation: string;
}

export interface ReportUserParams {
  violation_id: string;
  reported_user_id: number;
  explanation: string;
}

export interface ReportResponse {
  message: string;
}

export function reportPost(params: ReportPostParams, token: string): Promise<ReportResponse> {
  return api.post<ReportResponse>('/api/reports', params, token);
}

export function reportUser(params: ReportUserParams, token: string): Promise<ReportResponse> {
  return api.post<ReportResponse>('/api/reports', params, token);
}
