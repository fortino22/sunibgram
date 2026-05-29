import { api } from '../../api';
import type { AdminEvent, MessageResponse, PaginatedResponse } from './types';

export function listAdminEvents(
  token: string,
  params: { status?: string; name?: string; page?: number; limit?: number } = {}
): Promise<PaginatedResponse<AdminEvent>> {
  const qs = new URLSearchParams();
  if (params.status) qs.set('status', params.status);
  if (params.name) qs.set('name', params.name);
  qs.set('page', String(params.page ?? 1));
  qs.set('limit', String(params.limit ?? 20));
  return api.get<PaginatedResponse<AdminEvent>>(`/api/events/admin?${qs.toString()}`, token);
}

export function getEvent(token: string, id: string): Promise<AdminEvent> {
  return api.get<AdminEvent>(`/api/events/${id}`, token);
}

export function approveEvent(token: string, id: string): Promise<MessageResponse> {
  return api.patch<MessageResponse>(`/api/events/${id}/approve`, {}, token);
}

export function declineEvent(token: string, id: string, reason: string): Promise<MessageResponse> {
  return api.patch<MessageResponse>(`/api/events/${id}/decline`, { reason }, token);
}

export function requestRevision(token: string, id: string, note: string): Promise<MessageResponse> {
  return api.patch<MessageResponse>(`/api/events/${id}/revision`, { note }, token);
}

export function deleteEvent(token: string, id: string): Promise<MessageResponse> {
  return api.delete<MessageResponse>(`/api/events/${id}`, token);
}
