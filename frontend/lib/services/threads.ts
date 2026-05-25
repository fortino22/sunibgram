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

export interface PollOptionResponse {
  option_id: number;
  label: string;
  image_url?: string;
  vote_count: number;
  percent: number;
}

export interface PollResponse {
  poll_id: number;
  question?: string;
  ends_at: FormattedTime;
  total_votes: number;
  expired: boolean;
  user_voted_option_id?: number | null;
  options: PollOptionResponse[];
}

export interface ThreadResponse {
  thread_id: number;
  user_id: number;
  username: string;
  content: string;
  parent_id: number | null;
  root_id: number | null;
  media: ThreadMediaResponse[];
  poll?: PollResponse;
  like_count: number;
  repost_count: number;
  bookmark_count: number;
  comment_count: number;
  reposted_at?: FormattedTime;
  reposted_by_id?: number;
  reposted_by_username?: string;
  reposted_by_name?: string;
  created_at: FormattedTime;
  updated_at: FormattedTime;
  reply_permission?: string;
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
  name: string;
  username: string;
  is_premium: boolean;
  avatar_url?: string | null;
  faculty?: FeedAuthorFaculty;
}

export interface FeedStats {
  likes: number;
  reposts: number;
  bookmarks: number;
  comments: number;
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
  poll?: PollResponse;
  stats: FeedStats;
  created_at: FormattedTime;
  updated_at: FormattedTime;
  reply_permission?: string;
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
  reposted_by_id?: number;
  reposted_by_username?: string;
  reposted_by_name?: string;
  original_thread_id?: number;
  author: FeedAuthor;
  stats: FeedStats;
  media: ThreadMediaResponse[];
  poll?: PollResponse;
  is_liked?: boolean;
  is_bookmarked?: boolean;
  is_reposted?: boolean;
  is_pin?: boolean;
  pinned_by_admin?: boolean;
  reply_permission?: string;
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

export interface ThreadPollOptionInput {
  label: string;
  image?: File | null;
}

export interface ThreadPollInput {
  question?: string;
  options: ThreadPollOptionInput[];
  durationSeconds: number;
}

export interface ThreadMutationParams {
  content: string;
  files?: File[];
  parentId?: number | string;
  clearMedia?: boolean;
  isPrivate?: boolean;
  poll?: ThreadPollInput;
  replyPermission?: string;
  eventId?: string;
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
  if (params.isPrivate !== undefined) {
    formData.append('is_private', String(params.isPrivate));
  }
  if (params.replyPermission && !params.parentId) {
    formData.append('reply_permission', params.replyPermission);
  }
  if (params.eventId !== undefined) {
    formData.append('event_id', params.eventId);
  }
  params.files?.forEach((file) => {
    formData.append('files', file);
  });
  if (params.poll) {
    formData.append('poll_question', params.poll.question?.trim() ?? '');
    formData.append('poll_duration_seconds', String(params.poll.durationSeconds));
    params.poll.options.forEach((option) => {
      formData.append('poll_options', option.label);
    });
    const optionImages = params.poll.options.map((option) => option.image).filter((image): image is File => image instanceof File);
    if (optionImages.length === params.poll.options.length) {
      optionImages.forEach((image) => {
        formData.append('poll_option_images', image);
      });
    }
  }
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

export function getFollowingThreads(
  params: PaginationParams = {},
  token: string
): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/following${query(params)}`, token);
}

export function getFacultyThreads(
  facultyId: number | string,
  params: PaginationParams = {},
  token: string
): Promise<ThreadListResponse> {
  const search = new URLSearchParams();
  search.set('faculty_id', String(facultyId));
  if (params.page !== undefined) search.set('page', String(params.page));
  if (params.limit !== undefined) search.set('limit', String(params.limit));
  return api.get<ThreadListResponse>(`/api/threads/faculty?${search.toString()}`, token);
}

export function getMyThreads(userId: number | string, params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/user/${userId}${query(params)}`, token);
}

export function getMyReplies(userId: number | string, params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/user/${userId}/replies${query(params)}`, token);
}

export function getMyMediaThreads(userId: number | string, params: PaginationParams = {}, token: string): Promise<ThreadListResponse> {
  return api.get<ThreadListResponse>(`/api/threads/user/${userId}/media${query(params)}`, token);
}

export function getThread(id: number | string, token: string): Promise<ThreadResponse> {
  return api.get<ThreadResponse>(`/api/threads/${id}`, token);
}

export function getThreadFeed(id: number | string, token: string): Promise<FeedItemResponse> {
  return api.get<FeedItemResponse>(`/api/threads/${id}`, token);
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

export function voteOnPoll(
  id: number | string,
  optionId: number,
  token: string
): Promise<ThreadMessageResponse> {
  return api.post<ThreadMessageResponse>(`/api/threads/${id}/poll/vote`, { option_id: optionId }, token);
}

export function updateThread(
  id: number | string,
  params: ThreadMutationParams,
  token: string
): Promise<ThreadMessageResponse> {
  return putFormData<ThreadMessageResponse>(`/api/threads/${id}`, buildThreadFormData(params), token);
}

export function deleteThread(id: number | string, token: string): Promise<ThreadMessageResponse> {
  return api.delete<ThreadMessageResponse>(`/api/threads/${id}`, token);
}

export function updateReplyPermission(
  id: number | string,
  replyPermission: string,
  token: string
): Promise<ThreadMessageResponse> {
  return api.patch<ThreadMessageResponse>(`/api/threads/${id}/reply-permission`, { reply_permission: replyPermission }, token);
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
