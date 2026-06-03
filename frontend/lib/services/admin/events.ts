import { api } from '../../api';
import type { AdminEvent, AdminEventAudience, MessageResponse, PaginatedResponse } from './types';

// Backend returns audience entries in snake_case (is_all_regions); normalise to
// the camelCase shape the shared <AudienceBadges /> component expects.
interface RawAudience {
  id: number;
  name: string;
  region: string;
  is_all_regions: boolean;
}

function mapAudience(raw?: RawAudience[]): AdminEventAudience[] {
  return (raw ?? []).map((a) => ({
    id: a.id,
    name: a.name,
    region: a.region,
    isAllRegions: a.is_all_regions,
  }));
}

function mapEvent(raw: AdminEvent & { audience?: RawAudience[] }): AdminEvent {
  return { ...raw, audience: mapAudience(raw.audience as RawAudience[] | undefined) };
}

export async function listAdminEvents(
  token: string,
  params: { status?: string; name?: string; page?: number; limit?: number } = {}
): Promise<PaginatedResponse<AdminEvent>> {
  const qs = new URLSearchParams();
  if (params.status) qs.set('status', params.status);
  if (params.name) qs.set('name', params.name);
  qs.set('page', String(params.page ?? 1));
  qs.set('limit', String(params.limit ?? 20));
  const res = await api.get<PaginatedResponse<AdminEvent & { audience?: RawAudience[] }>>(
    `/api/events/admin?${qs.toString()}`,
    token,
  );
  return { ...res, data: res.data.map(mapEvent) };
}

export async function getEvent(token: string, id: string): Promise<AdminEvent> {
  const res = await api.get<AdminEvent & { audience?: RawAudience[] }>(`/api/events/${id}`, token);
  return mapEvent(res);
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
