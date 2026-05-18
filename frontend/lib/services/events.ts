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
  pdf_urls?: string[];
}

interface EventGroupsResponse {
  trending: EventListItemResponse[];
  running: EventListItemResponse[];
  future: EventListItemResponse[];
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

export interface EventGroups {
  trending: EventItem[];
  running: EventItem[];
  future: EventItem[];
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
    documentUrl: raw.attachment_url ?? raw.pdf_urls?.[0],
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
    formData.append('attachment', input.document);
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
  };
}

export async function getEvent(id: string, token: string): Promise<EventDetail> {
  const response = await api.get<EventDetailResponse>(`/api/events/${id}`, token);
  return mapEventDetail(response);
}

export function createEvent(input: CreateEventInput, token: string) {
  return api.post<{ message: string }>('/api/events', buildCreateEventFormData(input), token);
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

export function getOfficialEventPosts(
  id: string,
  params: PaginationParams,
  token: string
): Promise<PaginatedFeedResponse> {
  return api.get<PaginatedFeedResponse>(`/api/events/${id}/official-posts${queryString(params)}`, token);
}

export function getEventDiscussion(
  id: string,
  params: PaginationParams,
  token: string
): Promise<PaginatedFeedResponse> {
  return api.get<PaginatedFeedResponse>(`/api/events/${id}/discussion${queryString(params)}`, token);
}
