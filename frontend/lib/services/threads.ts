import { api } from '../api';

export type FormattedTime = string;

export interface ThreadMessageResponse {
  message: string;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface ThreadMediaResponse {
  media_id: number;
  url: string;
}

export interface ThreadResponse {
  thread_id: number;
  user_id: number;
  username: string;
  content: string;
  parent_id: number | null;
  root_id: number | null;
  media: ThreadMediaResponse[];
  like_count: number;
  repost_count: number;
  bookmark_count: number;
  reposted_at?: FormattedTime;
  created_at: FormattedTime;
  updated_at: FormattedTime;
}

export interface PaginatedThreadsResponse {
  data: ThreadResponse[];
  total: number;
  page: number;
  limit: number;
}

export interface ReplyResponse {
  parent: ThreadResponse | null;
  reply: ThreadResponse;
}

export interface FeedAuthorFaculty {
  faculty_id: number;
  faculty_name: string;
}

export interface FeedAuthor {
  id: number;
  username: string;
  is_premium: boolean;
  avatar_url?: string | null;
  faculty?: FeedAuthorFaculty;
}

export interface FeedStats {
  likes: number;
  reposts: number;
  bookmarks: number;
}

export interface FeedCommunity {
  id: number;
  name: string;
}

export interface FeedParentContext {
  id: number;
  parent_id: number | null;
  root_id: number | null;
  community?: FeedCommunity;
  author: FeedAuthor;
  content: string;
  media: ThreadMediaResponse[];
  stats: FeedStats;
  created_at: FormattedTime;
  updated_at: FormattedTime;
}

export interface FeedItemResponse {
  type: string;
  parent_context?: FeedParentContext;
  id: number;
  parent_id: number | null;
  root_id: number | null;
  community?: FeedCommunity;
  content: string;
  created_at: FormattedTime;
  updated_at: FormattedTime;
  time_ago: string;
  action_at?: FormattedTime;
  reposted_at?: FormattedTime;
  author: FeedAuthor;
  stats: FeedStats;
  media: ThreadMediaResponse[];
  is_liked?: boolean;
  is_bookmarked?: boolean;
  is_reposted?: boolean;
  is_pin?: boolean;
}

export interface FeedMeta {
  total: number;
  page: number;
  limit: number;
}

export interface PaginatedFeedResponse {
  data: FeedItemResponse[];
  meta: FeedMeta;
}

export type ThreadListResponse = PaginatedFeedResponse;

export type ThreadItem = FeedItemResponse;

export interface ThreadMutationParams {
  content: string;
  files?: File[];
  parentId?: number | string;
  clearMedia?: boolean;
}

export const DEFAULT_THREAD_CHAIN_PARAMS: Required<PaginationParams> = {
  page: 1,
  limit: 10,
};

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

export function getMyThreads(userId: number | string, params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/${userId}${query(params)}`, token);
}

export function getMyReplies(userId: number | string, params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/${userId}/replies${query(params)}`, token);
}

export function getThread(id: number | string, token: string): Promise<ThreadResponse> {
  return api.get<ThreadResponse>(`/api/threads/${id}`, token);
}

export function getThreadChain(
  id: number | string,
  params: PaginationParams = DEFAULT_THREAD_CHAIN_PARAMS,
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
