import { api } from '@/lib/api';

export type FeedbackType = 'bug' | 'feature_request';
export type FeedbackStatus = 'open' | 'planned' | 'in_progress' | 'resolved' | 'rejected';
export type FeedbackVote = 'like' | 'dislike';

export const FEEDBACK_TYPES: FeedbackType[] = ['bug', 'feature_request'];
export const FEEDBACK_STATUSES: FeedbackStatus[] = [
  'open',
  'planned',
  'in_progress',
  'resolved',
  'rejected',
];

export const FEEDBACK_TITLE_MAX = 200;
export const FEEDBACK_BODY_MAX = 5000;

export interface FeedbackAuthor {
  user_id: number;
  username: string;
  name: string;
  user_avatar: string;
  is_premium: boolean;
}

export interface FeedbackResponse {
  id: string;
  type: FeedbackType;
  title: string;
  body: string;
  status: FeedbackStatus;
  author: FeedbackAuthor;
  likes: number;
  dislikes: number;
  is_vote: FeedbackVote | null;
  created_at: string;
  updated_at: string;
}

export interface FeedbackListMeta {
  total: number;
  page: number;
  limit: number;
}

export interface PaginatedFeedbackResponse {
  data: FeedbackResponse[];
  meta: FeedbackListMeta;
}

export interface ListFeedbackParams {
  type?: FeedbackType | null;
  status?: FeedbackStatus | null;
  page?: number;
  limit?: number;
}

export interface CreateFeedbackInput {
  type: FeedbackType;
  title: string;
  body: string;
}

function buildListPath(params: ListFeedbackParams): string {
  const q = new URLSearchParams();
  if (params.type) q.set('type', params.type);
  if (params.status) q.set('status', params.status);
  if (params.page) q.set('page', String(params.page));
  if (params.limit) q.set('limit', String(params.limit));
  const qs = q.toString();
  return qs ? `/api/feedback?${qs}` : '/api/feedback';
}

export function listFeedback(
  params: ListFeedbackParams,
  token: string,
): Promise<PaginatedFeedbackResponse> {
  return api.get<PaginatedFeedbackResponse>(buildListPath(params), token);
}

export function getFeedback(id: string, token: string): Promise<FeedbackResponse> {
  return api.get<FeedbackResponse>(`/api/feedback/${id}`, token);
}

export function createFeedback(
  input: CreateFeedbackInput,
  token: string,
): Promise<{ message: string }> {
  return api.post<{ message: string }>('/api/feedback', input, token);
}

export function deleteFeedback(id: string, token: string): Promise<{ message: string }> {
  return api.delete<{ message: string }>(`/api/feedback/${id}`, token);
}

export function likeFeedback(id: string, token: string): Promise<{ message: string }> {
  return api.post<{ message: string }>(`/api/feedback/${id}/like`, undefined, token);
}

export function dislikeFeedback(id: string, token: string): Promise<{ message: string }> {
  return api.post<{ message: string }>(`/api/feedback/${id}/dislike`, undefined, token);
}

export function updateFeedbackStatus(
  id: string,
  status: FeedbackStatus,
  token: string,
): Promise<{ message: string }> {
  return api.patch<{ message: string }>(`/api/feedback/${id}/status`, { status }, token);
}

export const FEEDBACK_TYPE_LABEL: Record<FeedbackType, string> = {
  bug: 'Bug',
  feature_request: 'Feature',
};

export const FEEDBACK_STATUS_LABEL: Record<FeedbackStatus, string> = {
  open: 'Open',
  planned: 'Planned',
  in_progress: 'In progress',
  resolved: 'Resolved',
  rejected: 'Rejected',
};
