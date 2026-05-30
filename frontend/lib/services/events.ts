import { api } from '../api';
import type { PaginatedFeedResponse } from './threads';

// ── Backend response shapes ──────────────────────────────────────────────────

interface EventListItemResponse {
  id: string;
  user_id: number;
  title: string;
  description: string;
  location: string;
  type: 'official' | 'unofficial';
  start_time: string | null;
  end_time: string | null;
  banner_url?: string;
  created_at: string;
  updated_at: string;
}

interface EventDetailResponse extends EventListItemResponse {
  status: string;
  attachment_url?: string;
  doc_urls?: string[];
}

interface EventReviewNoteResponse {
  revision_number: number;
  notes: string;
  created_at: string;
}

interface MyEventItemResponse extends EventDetailResponse {
  notes: EventReviewNoteResponse[];
}

interface MyEventsApiResponse {
  data: MyEventItemResponse[];
  meta: { total: number; page: number; limit: number };
}

interface EventGroupsResponse {
  trending: EventListItemResponse[];
  running: EventListItemResponse[];
  future: EventListItemResponse[];
  past: EventListItemResponse[];
}

interface EventShareApiResponse {
  shareable_link: string;
  share_count: number;
}

// ── Frontend model shapes ────────────────────────────────────────────────────

export interface EventItem {
  id: string;
  userId: number;
  title: string;
  description: string;
  location: string;
  type: 'official' | 'unofficial';
  startTime: string | null;
  endTime: string | null;
  banner?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EventDetail extends EventItem {
  status: string;
  documentUrl?: string;
}

export interface EventReviewNote {
  revisionNumber: number;
  notes: string;
  createdAt: string;
}

export interface MyEventItem extends EventItem {
  status: string;
  notes: EventReviewNote[];
}

export interface EventGroups {
  trending: EventItem[];
  running: EventItem[];
  future: EventItem[];
  past: EventItem[];
}

export interface CreateEventInput {
  title: string;
  description: string;
  location: string;
  startTime: string;
  endTime: string;
  type: 'official' | 'unofficial';
  banner: File;
  document?: File;
}

export interface UpdateEventInput {
  title: string;
  description: string;
  location: string;
  startTime: string;
  endTime: string;
  type: 'official' | 'unofficial';
  banner?: File;
  document?: File;
}

export interface EventShareResult {
  shareableLink: string;
  shareCount: number;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

// ── Mappers ──────────────────────────────────────────────────────────────────

function mapEventItem(raw: EventListItemResponse): EventItem {
  return {
    id: raw.id,
    userId: raw.user_id,
    title: raw.title,
    description: raw.description,
    location: raw.location,
    type: raw.type as 'official' | 'unofficial',
    startTime: raw.start_time,
    endTime: raw.end_time,
    banner: raw.banner_url,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

function mapEventDetail(raw: EventDetailResponse): EventDetail {
  return {
    ...mapEventItem(raw),
    status: raw.status,
    documentUrl: raw.attachment_url ?? raw.doc_urls?.[0],
  };
}

function queryString(params?: PaginationParams): string {
  const search = new URLSearchParams();
  if (params?.page !== undefined) search.set('page', String(params.page));
  if (params?.limit !== undefined) search.set('limit', String(params.limit));
  const value = search.toString();
  return value ? `?${value}` : '';
}

function buildCreateEventFormData(input: CreateEventInput): FormData {
  const formData = new FormData();
  formData.append('title', input.title);
  formData.append('description', input.description);
  formData.append('location', input.location);
  formData.append('start_time', input.startTime);
  formData.append('end_time', input.endTime);
  formData.append('type', input.type);
  formData.append('banner', input.banner);
  if (input.document) {
    formData.append('docs', input.document);
  }
  return formData;
}

// ── API calls ────────────────────────────────────────────────────────────────

export async function getEvents(token: string): Promise<EventGroups> {
  const response = await api.get<EventGroupsResponse>('/api/events', token);
  return {
    trending: (response.trending ?? []).map(mapEventItem),
    running: (response.running ?? []).map(mapEventItem),
    future: (response.future ?? []).map(mapEventItem),
    past: (response.past ?? []).map(mapEventItem),
  };
}

export async function getTrendingEvents(token: string): Promise<EventItem[]> {
  const response = await api.get<EventListItemResponse[]>('/api/events/trending', token);
  return (response ?? []).map(mapEventItem);
}

export async function getEvent(id: string, token: string): Promise<EventDetail> {
  const response = await api.get<EventDetailResponse>(`/api/events/${id}`, token);
  return mapEventDetail(response);
}

export function createEvent(input: CreateEventInput, token: string) {
  return api.post<{ message: string }>('/api/events', buildCreateEventFormData(input), token);
}

export function updateEvent(id: string, input: UpdateEventInput, token: string) {
  const formData = new FormData();
  formData.append('title', input.title);
  formData.append('description', input.description);
  formData.append('location', input.location);
  formData.append('start_time', input.startTime);
  formData.append('end_time', input.endTime);
  formData.append('type', input.type);
  if (input.banner) formData.append('banner', input.banner);
  if (input.document) formData.append('docs', input.document);
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  return fetch(`${baseUrl}/api/events/${id}`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  }).then(async (res) => {
    if (!res.ok) {
      const err = await res.json().catch(() => ({ message: 'Request failed' }));
      throw new Error((err as { message?: string }).message ?? 'Request failed');
    }
    return res.json() as Promise<{ message: string }>;
  });
}

export function deleteEvent(id: string, token: string) {
  return api.delete<{ message: string }>(`/api/events/${id}`, token);
}

export function bookmarkEvent(id: string, token: string) {
  return api.post<{ message: string }>(`/api/events/${id}/bookmark`, undefined, token);
}

export function unbookmarkEvent(id: string, token: string) {
  return api.delete<{ message: string }>(`/api/events/${id}/bookmark`, token);
}

export async function shareEvent(id: string, token: string): Promise<EventShareResult> {
  const response = await api.post<EventShareApiResponse>(`/api/events/${id}/share`, undefined, token);
  return {
    shareableLink: response.shareable_link,
    shareCount: response.share_count,
  };
}

export function joinEvent(id: string, token: string) {
  return api.post<{ message: string }>(`/api/events/${id}/join`, undefined, token);
}

export function leaveEvent(id: string, token: string) {
  return api.delete<{ message: string }>(`/api/events/${id}/leave`, token);
}

export async function getMyJoinedEvents(token: string): Promise<EventItem[]> {
  const response = await api.get<unknown>('/api/events/joined/me', token);
  const items: EventListItemResponse[] = Array.isArray(response)
    ? (response as EventListItemResponse[])
    : Array.isArray((response as { data?: unknown }).data)
      ? (response as { data: EventListItemResponse[] }).data
      : [];
  return items.map(mapEventItem);
}

export async function getMyEvents(token: string, status?: string, name?: string): Promise<MyEventItem[]> {
  const params = new URLSearchParams();
  if (status) params.set('status', status);
  if (name && name.trim()) params.set('name', name.trim());
  const qs = params.toString();
  const response = await api.get<MyEventsApiResponse>(`/api/events/me${qs ? `?${qs}` : ''}`, token);
  return (response.data ?? []).map((raw) => ({
    ...mapEventItem(raw),
    status: raw.status,
    notes: (raw.notes ?? []).map((n) => ({
      revisionNumber: n.revision_number,
      notes: n.notes,
      createdAt: n.created_at,
    })),
  }));
}

export function getOfficialEventPosts(
  id: string,
  params: PaginationParams,
  token: string
): Promise<PaginatedFeedResponse> {
  const qs = new URLSearchParams({ filter: 'official' });
  if (params.page !== undefined) qs.set('page', String(params.page));
  if (params.limit !== undefined) qs.set('limit', String(params.limit));
  return api.get<PaginatedFeedResponse>(`/api/events/${id}/threads?${qs.toString()}`, token);
}

export function getEventDiscussion(
  id: string,
  params: PaginationParams,
  token: string
): Promise<PaginatedFeedResponse> {
  const qs = new URLSearchParams({ filter: 'unofficial' });
  if (params.page !== undefined) qs.set('page', String(params.page));
  if (params.limit !== undefined) qs.set('limit', String(params.limit));
  return api.get<PaginatedFeedResponse>(`/api/events/${id}/threads?${qs.toString()}`, token);
}
