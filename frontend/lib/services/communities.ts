import { api } from '../api';
import type { ThreadListResponse } from './threads';

export type CommunityVisibility = 'PUBLIC' | 'PRIVATE';
export type CommunityRole = 'MEMBER' | 'MODERATOR' | 'ADMIN' | 'OWNER';

// Backend DTO types — match exact JSON field names from Go backend
export interface CommunityData {
  community_id: number;
  user_id: number;
  username: string;
  name: string;
  description: string;
  rules: { id: number; text: string }[];
  category_name: string;
  visibility: CommunityVisibility;
  banner_url: string;
  avatar_url?: string | null;
  member_count: number;
  top_members: CommunityTopMemberData[];
  created_at: string;
  updated_at: string;
}

export interface CommunityTopMemberData {
  user_id: string;
  avatar_url: string;
}

export interface CommunityMeta {
  total: number;
  page: number;
  limit: number;
}

export interface PaginatedCommunityData {
  data: CommunityData[];
  meta: CommunityMeta;
}

export interface CommunityMemberData {
  community_id: number;
  user_id: number;
  username: string;
  avatar_url: string;
  role: CommunityRole;
  joined_at: string;
}

export interface PaginatedCommunityMemberData {
  data: CommunityMemberData[];
  meta: CommunityMeta;
}

export interface CommunityRequestData {
  community_id: number;
  user_id: number;
  name: string;
  username: string;
  avatar_url: string;
  is_premium: boolean;
  status: string;
  created_at: string;
}

export interface PaginatedCommunityRequestData {
  data: CommunityRequestData[];
  meta: CommunityMeta;
}

export interface DiscoverCommunityMemberData {
  user_id: string;
  avatar_url: string;
}

export interface DiscoverCommunityData {
  community_id: number;
  name: string;
  member_count: number;
  category_name: string;
  avatar_url?: string;
  banner_url?: string;
  top_members: DiscoverCommunityMemberData[];
}

export interface CommunityMessageResponse {
  message: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  created_at?: string;
}

export interface CategoryListResponse {
  data: Category[];
}

export interface CommunityParams {
  name: string;
  description: string;
  rules: string[];
  categoryId: string;
  visibility: CommunityVisibility;
  banner?: File | null;
}

export interface UpdateCommunityParams {
  name?: string;
  description?: string;
  keep_rule_ids?: number[];
  new_rules?: string[];
  categoryId?: string;
  visibility?: CommunityVisibility;
  banner?: File | null;
  clearBanner?: boolean;
}

export interface CommunityThreadParams {
  content: string;
  files?: File[];
}

function appendOptionalString(formData: FormData, key: string, value?: string | boolean): void {
  if (value !== undefined) {
    formData.append(key, String(value));
  }
}

function buildCreateCommunityFormData(params: CommunityParams): FormData {
  const formData = new FormData();
  appendOptionalString(formData, 'name', params.name);
  appendOptionalString(formData, 'description', params.description);
  params.rules.forEach((rule) => formData.append('rules', rule));
  appendOptionalString(formData, 'category_id', params.categoryId);
  appendOptionalString(formData, 'visibility', params.visibility);
  if (params.banner) {
    formData.append('banner', params.banner);
  }
  return formData;
}

function buildUpdateCommunityFormData(params: UpdateCommunityParams): FormData {
  const formData = new FormData();
  appendOptionalString(formData, 'name', params.name);
  appendOptionalString(formData, 'description', params.description);
  appendOptionalString(formData, 'category_id', params.categoryId);
  appendOptionalString(formData, 'visibility', params.visibility);
  if (params.keep_rule_ids) {
    params.keep_rule_ids.forEach((id) => formData.append('keep_rule_ids', String(id)));
  }
  if (params.new_rules) {
    params.new_rules.forEach((rule) => formData.append('new_rules', rule));
  }
  if ('clearBanner' in params) {
    appendOptionalString(formData, 'clear_banner', params.clearBanner);
  }
  if (params.banner) {
    formData.append('banner', params.banner);
  }
  return formData;
}

function buildCommunityThreadFormData(params: CommunityThreadParams): FormData {
  const formData = new FormData();
  formData.append('content', params.content);
  params.files?.forEach((file) => {
    formData.append('files', file);
  });
  return formData;
}

async function patchJson<T>(path: string, body: Record<string, string>, token: string): Promise<T> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const res = await fetch(`${baseUrl}${normalizedPath}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const error: unknown = await res.json().catch(() => ({ message: 'Request failed' }));
    const message =
      typeof error === 'object' && error !== null && 'message' in error && typeof error.message === 'string'
        ? error.message
        : 'Request failed';
    throw new Error(message);
  }

  return res.json() as Promise<T>;
}

async function putFormData<T>(path: string, formData: FormData, token: string): Promise<T> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const res = await fetch(`${baseUrl}${normalizedPath}`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });

  if (!res.ok) {
    const error: unknown = await res.json().catch(() => ({ message: 'Request failed' }));
    const message =
      typeof error === 'object' && error !== null && 'message' in error && typeof error.message === 'string'
        ? error.message
        : 'Request failed';
    throw new Error(message);
  }

  return res.json() as Promise<T>;
}

function paginationQuery(params?: { page?: number; limit?: number }): string {
  const search = new URLSearchParams();
  if (params?.page !== undefined) search.set('page', String(params.page));
  if (params?.limit !== undefined) search.set('limit', String(params.limit));
  const value = search.toString();
  return value ? `?${value}` : '';
}

export function getCategories(token?: string): Promise<CategoryListResponse> {
  return api.get<CategoryListResponse>('/api/categories', token);
}

export function getMyCommunities(
  token: string,
  params?: { page?: number; limit?: number }
): Promise<PaginatedCommunityData> {
  return api.get<PaginatedCommunityData>(`/api/communities/me${paginationQuery(params)}`, token);
}

export function getJoinedCommunities(
  token: string,
  params?: { page?: number; limit?: number }
): Promise<PaginatedCommunityData> {
  return api.get<PaginatedCommunityData>(`/api/communities/joined${paginationQuery(params)}`, token);
}

export function getCommunity(id: number | string, token: string): Promise<CommunityData> {
  return api.get<CommunityData>(`/api/communities/${id}`, token);
}

export async function getDiscoverCommunities(
  token: string,
  categoryId?: string
): Promise<DiscoverCommunityData[]> {
  const q = categoryId ? `?category=${encodeURIComponent(categoryId)}` : '';
  const res = await api.get<{ data: DiscoverCommunityData[] }>(`/api/communities/discover${q}`, token);
  return res.data;
}

export function createCommunity(params: CommunityParams, token: string): Promise<CommunityData> {
  return api.post<CommunityData>('/api/communities', buildCreateCommunityFormData(params), token);
}

export function updateCommunity(
  id: number | string,
  params: UpdateCommunityParams,
  token: string
): Promise<CommunityData> {
  return putFormData<CommunityData>(`/api/communities/${id}`, buildUpdateCommunityFormData(params), token);
}

export function deleteCommunity(id: number | string, token: string): Promise<CommunityMessageResponse> {
  return api.delete<CommunityMessageResponse>(`/api/communities/${id}`, token);
}

export function joinCommunity(id: number | string, token: string): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(`/api/communities/${id}/join`, undefined, token);
}

export function leaveCommunity(id: number | string, token: string): Promise<CommunityMessageResponse> {
  return api.delete<CommunityMessageResponse>(`/api/communities/${id}/leave`, token);
}

export function getCommunityMembers(
  id: number | string,
  token: string,
  params?: { page?: number; limit?: number }
): Promise<PaginatedCommunityMemberData> {
  return api.get<PaginatedCommunityMemberData>(
    `/api/communities/${id}/members${paginationQuery(params)}`,
    token
  );
}

export function kickCommunityMember(
  id: number | string,
  userId: number | string,
  token: string
): Promise<CommunityMessageResponse> {
  return api.delete<CommunityMessageResponse>(`/api/communities/${id}/members/${userId}`, token);
}

export function updateCommunityMemberRole(
  id: number | string,
  userId: number | string,
  role: CommunityRole,
  token: string
): Promise<CommunityMessageResponse> {
  return patchJson<CommunityMessageResponse>(
    `/api/communities/${id}/members/${userId}/role`,
    { role },
    token
  );
}

export function requestToJoinCommunity(id: number | string, token: string): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(`/api/communities/${id}/requests`, undefined, token);
}

export function cancelCommunityJoinRequest(id: number | string, token: string): Promise<CommunityMessageResponse> {
  return api.delete<CommunityMessageResponse>(`/api/communities/${id}/requests`, token);
}

export function getCommunityJoinRequests(
  id: number | string,
  token: string,
  params?: { page?: number; limit?: number }
): Promise<PaginatedCommunityRequestData> {
  return api.get<PaginatedCommunityRequestData>(
    `/api/communities/${id}/requests${paginationQuery(params)}`,
    token
  );
}

export function approveCommunityJoinRequest(
  id: number | string,
  userId: number | string,
  token: string
): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(
    `/api/communities/${id}/requests/${userId}/approve`,
    undefined,
    token
  );
}

export function declineCommunityJoinRequest(
  id: number | string,
  userId: number | string,
  token: string
): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(
    `/api/communities/${id}/requests/${userId}/decline`,
    undefined,
    token
  );
}

export function getCommunityThreadsLatest(
  id: number | string,
  token: string,
  params?: { page?: number; limit?: number }
): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(
    `/api/communities/${id}/threads/latest${paginationQuery(params)}`,
    token
  );
}

export function getCommunityThreadsTop(
  id: number | string,
  token: string,
  params?: { page?: number; limit?: number }
): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(
    `/api/communities/${id}/threads/top${paginationQuery(params)}`,
    token
  );
}

interface CommunityMediaRaw {
  thread_id: number;
  media: { media_id: number; url: string }[];
}

export async function getCommunityThreadsMedia(
  id: number | string,
  token: string,
  params?: { page?: number; limit?: number }
): Promise<{ data: { id: number; media: { url: string }[] }[] }> {
  const raw = await api.get<{ data: CommunityMediaRaw[] }>(
    `/api/communities/${id}/threads/media${paginationQuery(params)}`,
    token
  );
  return {
    data: raw.data.map((item) => ({ id: item.thread_id, media: item.media })),
  };
}

export function createCommunityThread(
  id: number | string,
  params: CommunityThreadParams,
  token: string
): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(
    `/api/communities/${id}/threads`,
    buildCommunityThreadFormData(params),
    token
  );
}

export function getMyCommunitiesThreadsFeed(
  token: string,
  params?: { page?: number; limit?: number }
): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/communities/me/threads${paginationQuery(params)}`, token);
}

export function getExploreCommunitiesThreads(
  token: string,
  categoryId?: string,
  params?: { page?: number; limit?: number }
): Promise<ThreadListResponse> {
  const search = new URLSearchParams();
  if (categoryId) search.set('categoryId', categoryId);
  if (params?.page !== undefined) search.set('page', String(params.page));
  if (params?.limit !== undefined) search.set('limit', String(params.limit));
  const q = search.toString() ? `?${search.toString()}` : '';
  return api.get<ThreadListResponse>(`/api/communities/threads${q}`, token);
}

export function searchCommunities(
  query: string,
  token: string
): Promise<{ data: DiscoverCommunityData[] }> {
  return api.get<{ data: DiscoverCommunityData[] }>(
    `/api/communities/search?community_name=${encodeURIComponent(query)}`,
    token
  );
}

export function searchCommunityThreadsTop(
  token: string,
  query: string,
  params?: { page?: number; limit?: number }
): Promise<ThreadListResponse> {
  const search = new URLSearchParams({ query });
  if (params?.page !== undefined) search.set('page', String(params.page));
  if (params?.limit !== undefined) search.set('limit', String(params.limit));
  return api.get<ThreadListResponse>(`/api/communities/threads/top?${search.toString()}`, token);
}

export function searchCommunityThreadsLatest(
  token: string,
  query: string,
  params?: { page?: number; limit?: number }
): Promise<ThreadListResponse> {
  const search = new URLSearchParams({ query });
  if (params?.page !== undefined) search.set('page', String(params.page));
  if (params?.limit !== undefined) search.set('limit', String(params.limit));
  return api.get<ThreadListResponse>(`/api/communities/threads/latest?${search.toString()}`, token);
}
