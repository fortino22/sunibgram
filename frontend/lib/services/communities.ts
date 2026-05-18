import { api } from '../api';

export type CommunityVisibility = 'PUBLIC' | 'PRIVATE';
export type CommunityRole = 'MEMBER' | 'MODERATOR' | 'ADMIN' | 'OWNER';

export interface CommunityMessageResponse {
  message: string;
}

export interface Community {
  id: number;
  name: string;
  description?: string;
  rules?: string;
  category_id?: string;
  category?: Category | null;
  visibility?: CommunityVisibility;
  avatar?: string | null;
  banner?: string | null;
  [key: string]: unknown;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  created_at?: string;
}

export interface CategoryListResponse {
  data: Category[];
  [key: string]: unknown;
}

export interface CommunityMember {
  id: number;
  user_id?: number;
  role?: CommunityRole;
  [key: string]: unknown;
}

export interface CommunityThread {
  id: number;
  content: string;
  community_id?: number;
  [key: string]: unknown;
}

export interface CommunityListResponse {
  data: Community[];
  [key: string]: unknown;
}

export interface CommunityResponse {
  data: Community;
  message?: string;
  [key: string]: unknown;
}

export interface CommunityMembersResponse {
  data: CommunityMember[];
  [key: string]: unknown;
}

export interface CommunityThreadsResponse {
  data: CommunityThread[];
  [key: string]: unknown;
}

export interface CommunityParams {
  name: string;
  description: string;
  rules: string;
  categoryId: string;
  visibility: CommunityVisibility;
  avatar?: File | null;
  banner?: File | null;
}

export interface UpdateCommunityParams extends Partial<CommunityParams> {
  clearAvatar?: boolean;
  clearBanner?: boolean;
}

export interface CommunityThreadParams {
  content: string;
  files?: File[];
  clearMedia?: boolean;
}

function appendOptionalString(formData: FormData, key: string, value?: string | boolean): void {
  if (value !== undefined) {
    formData.append(key, String(value));
  }
}

function buildCommunityFormData(params: CommunityParams | UpdateCommunityParams): FormData {
  const formData = new FormData();
  appendOptionalString(formData, 'name', params.name);
  appendOptionalString(formData, 'description', params.description);
  appendOptionalString(formData, 'rules', params.rules);
  appendOptionalString(formData, 'category_id', params.categoryId);
  appendOptionalString(formData, 'visibility', params.visibility);

  if ('clearAvatar' in params) {
    appendOptionalString(formData, 'clear_avatar', params.clearAvatar);
  }
  if ('clearBanner' in params) {
    appendOptionalString(formData, 'clear_banner', params.clearBanner);
  }
  if (params.avatar) {
    formData.append('avatar', params.avatar);
  }
  if (params.banner) {
    formData.append('banner', params.banner);
  }

  return formData;
}

function buildCommunityThreadFormData(params: CommunityThreadParams): FormData {
  const formData = new FormData();
  formData.append('content', params.content);
  if (params.clearMedia !== undefined) {
    formData.append('clear_media', String(params.clearMedia));
  }
  params.files?.forEach((file) => {
    formData.append('files', file);
  });
  return formData;
}

async function requestWithMethod<T>(
  method: 'PATCH' | 'PUT',
  path: string,
  body: FormData | Record<string, string>,
  token: string
): Promise<T> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const isFormData = body instanceof FormData;
  const res = await fetch(`${baseUrl}${normalizedPath}`, {
    method,
    headers: isFormData
      ? { Authorization: `Bearer ${token}` }
      : { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: isFormData ? body : JSON.stringify(body),
  });

  if (!res.ok) {
    const error: unknown = await res.json().catch(() => ({ message: 'Request failed' }));
    const message =
      typeof error === 'object' &&
      error !== null &&
      'message' in error &&
      typeof error.message === 'string'
        ? error.message
        : 'Request failed';
    throw new Error(message);
  }

  return res.json() as Promise<T>;
}

export function getCategories(token?: string): Promise<CategoryListResponse> {
  return api.get<CategoryListResponse>('/api/categories', token);
}

export function getMyCommunities(token: string): Promise<CommunityListResponse> {
  return api.get<CommunityListResponse>('/api/communities/me', token);
}

export function getJoinedCommunities(token: string): Promise<CommunityListResponse> {
  return api.get<CommunityListResponse>('/api/communities/joined', token);
}

export function getCommunity(id: number | string, token: string): Promise<CommunityResponse> {
  return api.get<CommunityResponse>(`/api/communities/${id}`, token);
}

export function createCommunity(params: CommunityParams, token: string): Promise<CommunityResponse> {
  return api.post<CommunityResponse>('/api/communities', buildCommunityFormData(params), token);
}

export function updateCommunity(
  id: number | string,
  params: UpdateCommunityParams,
  token: string
): Promise<CommunityResponse> {
  return requestWithMethod<CommunityResponse>('PUT', `/api/communities/${id}`, buildCommunityFormData(params), token);
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
  token: string
): Promise<CommunityMembersResponse> {
  return api.get<CommunityMembersResponse>(`/api/communities/${id}/members`, token);
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
  return requestWithMethod<CommunityMessageResponse>('PATCH', `/api/communities/${id}/members/${userId}/role`, { role }, token);
}

export function requestToJoinCommunity(id: number | string, token: string): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(`/api/communities/${id}/requests`, undefined, token);
}

export function cancelCommunityJoinRequest(
  id: number | string,
  token: string
): Promise<CommunityMessageResponse> {
  return api.delete<CommunityMessageResponse>(`/api/communities/${id}/requests`, token);
}

export function getCommunityJoinRequests(
  id: number | string,
  token: string
): Promise<CommunityMembersResponse> {
  return api.get<CommunityMembersResponse>(`/api/communities/${id}/requests`, token);
}

export function approveCommunityJoinRequest(
  id: number | string,
  userId: number | string,
  token: string
): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(`/api/communities/${id}/requests/${userId}/approve`, undefined, token);
}

export function declineCommunityJoinRequest(
  id: number | string,
  userId: number | string,
  token: string
): Promise<CommunityMessageResponse> {
  return api.post<CommunityMessageResponse>(`/api/communities/${id}/requests/${userId}/decline`, undefined, token);
}

export function getCommunityThreads(
  id: number | string,
  token: string
): Promise<CommunityThreadsResponse> {
  return api.get<CommunityThreadsResponse>(`/api/communities/${id}/threads`, token);
}

export function createCommunityThread(
  id: number | string,
  params: CommunityThreadParams,
  token: string
): Promise<CommunityThreadsResponse> {
  return api.post<CommunityThreadsResponse>(
    `/api/communities/${id}/threads`,
    buildCommunityThreadFormData(params),
    token
  );
}
