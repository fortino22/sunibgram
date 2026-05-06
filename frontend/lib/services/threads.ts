import { api } from '../api';

export interface ThreadMessageResponse {
  message: string;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface ThreadAuthor {
  id: number;
  username: string;
  is_premium: boolean;
}

export interface ThreadItem {
  type: 'POST';
  id: number;
  parent_id: null | number;
  root_id: null | number;
  content: string;
  created_at: string;
  updated_at: string;
  time_ago: string;
  author: ThreadAuthor;
  stats: {
    likes: number;
    reposts: number;
    bookmarks: number;
  };
  media: Array<{
    media_id: number;
    url: string;
  }>;
  is_liked: boolean;
  is_bookmarked: boolean;
  is_reposted: boolean;
  is_pin: boolean;
}

export interface ThreadListResponse {
  data: ThreadItem[];
  meta: {
    total: number;
    page: number;
    limit: number;
  };
}

export interface ThreadResponse {
  data: ThreadItem;
  message?: string;
  [key: string]: unknown;
}

export interface ThreadMutationParams {
  content: string;
  files?: File[];
  parentId?: number | string;
  clearMedia?: boolean;
}

function query(params?: PaginationParams): string {
  const search = new URLSearchParams();
  if (params?.page !== undefined) {
    search.set('page', String(params.page));
  }
  if (params?.limit !== undefined) {
    search.set('limit', String(params.limit));
  }
  const value = search.toString();
  return value ? `?${value}` : '';
}

function buildThreadFormData(params: ThreadMutationParams): FormData {
  const formData = new FormData();
  formData.append('content', params.content);
  if (params.parentId !== undefined) {
    formData.append('parent_id', String(params.parentId));
  }
  if (params.clearMedia !== undefined) {
    formData.append('clear_media', String(params.clearMedia));
  }
  params.files?.forEach((file) => {
    formData.append('files', file);
  });
  return formData;
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

export function getThreads(params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads${query(params)}`, token);
}

export function getFollowingThreads(token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>('/api/threads/following', token);
}

export function getMyThreads(params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/me${query(params)}`, token);
}

export function getMyReplies(params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/me/replies${query(params)}`, token);
}

export function getThread(id: number | string, token: string): Promise<ThreadResponse> {
  return api.get<ThreadResponse>(`/api/threads/${id}`, token);
}

export function getThreadChain(
  id: number | string,
  params: PaginationParams = {},
  token: string
): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/${id}/chain${query(params)}`, token);
}

export function createThread(params: ThreadMutationParams, token: string): Promise<ThreadResponse> {
  return api.post<ThreadResponse>('/api/threads', buildThreadFormData(params), token);
}

export function updateThread(
  id: number | string,
  params: ThreadMutationParams,
  token: string
): Promise<ThreadResponse> {
  return putFormData<ThreadResponse>(`/api/threads/${id}`, buildThreadFormData(params), token);
}

export function deleteThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.delete<ThreadMessageResponse>(`/api/threads/${id}`, token);
}

export function likeThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.post<ThreadMessageResponse>(`/api/threads/${id}/like`, undefined, token);
}

export function unlikeThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.delete<ThreadMessageResponse>(`/api/threads/${id}/unlike`, token);
}

export function bookmarkThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.post<ThreadMessageResponse>(`/api/threads/${id}/bookmark`, undefined, token);
}

export function unbookmarkThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.delete<ThreadMessageResponse>(`/api/threads/${id}/unbookmark`, token);
}

export function repostThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.post<ThreadMessageResponse>(`/api/threads/${id}/repost`, undefined, token);
}

export function unrepostThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.delete<ThreadMessageResponse>(`/api/threads/${id}/unrepost`, token);
}

export function pinThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.post<ThreadMessageResponse>(`/api/threads/${id}/pin`, undefined, token);
}

export function getPins(token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>('/api/pins', token);
}

export function getLikes(token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>('/api/likes', token);
}

export function getBookmarks(token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>('/api/bookmarks', token);
}
