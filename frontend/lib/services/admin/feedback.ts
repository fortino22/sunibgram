import { api } from '../../api';
import type { Feedback, MessageResponse, PaginatedResponse } from './types';

export function listFeedback(
  token: string,
  params: { type?: string; status?: string; page?: number; limit?: number } = {}
): Promise<PaginatedResponse<Feedback>> {
  const qs = new URLSearchParams();
  if (params.type) qs.set('type', params.type);
  if (params.status) qs.set('status', params.status);
  qs.set('page', String(params.page ?? 1));
  qs.set('limit', String(params.limit ?? 20));
  return api.get<PaginatedResponse<Feedback>>(`/api/feedback?${qs.toString()}`, token);
}

export function updateFeedbackStatus(token: string, id: string, status: string): Promise<MessageResponse> {
  return api.patch<MessageResponse>(`/api/feedback/${id}/status`, { status }, token);
}
