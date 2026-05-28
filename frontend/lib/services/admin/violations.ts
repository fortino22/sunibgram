import { api } from '../../api';
import type { MessageResponse, Violation } from './types';

interface ListResponse {
  data: Violation[];
}

export async function listViolations(token: string): Promise<Violation[]> {
  const res = await api.get<ListResponse | Violation[]>('/api/violations', token);
  if (Array.isArray(res)) return res;
  return res.data ?? [];
}

export function createViolation(token: string, body: { name: string; description: string }): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/violations', body, token);
}

export function updateViolation(token: string, id: string, body: { name: string; description: string }): Promise<MessageResponse> {
  return api.put<MessageResponse>(`/api/violations/${id}`, body, token);
}

export function deleteViolation(token: string, id: string): Promise<MessageResponse> {
  return api.delete<MessageResponse>(`/api/violations/${id}`, token);
}
