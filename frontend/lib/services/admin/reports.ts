import { api } from '../../api';
import type { MessageResponse, Report } from './types';

interface ReportsListResponse {
  data: Report[];
}

export async function listReports(token: string): Promise<Report[]> {
  const res = await api.get<ReportsListResponse | Report[]>('/api/reports', token);
  if (Array.isArray(res)) return res;
  return res.data ?? [];
}

export function getReport(token: string, id: string): Promise<Report> {
  return api.get<Report>(`/api/reports/${id}`, token);
}

export function updateReportStatus(token: string, id: string, status: string): Promise<MessageResponse> {
  return api.put<MessageResponse>(`/api/reports/${id}/status`, { status }, token);
}

export function deleteReport(token: string, id: string): Promise<MessageResponse> {
  return api.delete<MessageResponse>(`/api/reports/${id}`, token);
}
