import { api } from '../api';

// Backend DTO shapes — snake_case to match the Go JSON tags (same convention as
// communities.ts).

export interface CommunityEventAuthor {
  user_id: number;
  username: string;
  name: string;
  avatar_url?: string;
}

export interface CommunityEventReactionSummaryItem {
  emoji: string;
  count: number;
}

export interface CommunityEventCommentData {
  id: number;
  author: CommunityEventAuthor;
  content: string;
  is_pinned: boolean;
  created_at: string;
}

export interface CommunityEventDetail {
  id: number;
  community_id: number;
  created_by: CommunityEventAuthor;
  title: string;
  description: string;
  location: string;
  start_time: string;
  end_time: string;
  banner_url?: string;
  invitees: CommunityEventAuthor[];
  reaction_summary: CommunityEventReactionSummaryItem[];
  my_reaction?: string;
  pinned_comment?: CommunityEventCommentData | null;
  comment_count: number;
  can_comment: boolean;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

export interface CommunityEventListItem {
  id: number;
  community_id: number;
  title: string;
  description: string;
  location: string;
  start_time: string;
  end_time: string;
  banner_url?: string;
  comment_count: number;
  created_at: string;
}

export interface CommunityEventMeta {
  total: number;
  page: number;
  limit: number;
}

export interface PaginatedCommunityEvents {
  data: CommunityEventListItem[];
  meta: CommunityEventMeta;
}

export interface PaginatedCommunityEventComments {
  data: CommunityEventCommentData[];
  pinned_comment_id?: number | null;
  is_admin: boolean;
  meta: CommunityEventMeta;
}

export interface CommunityEventMessageResponse {
  message: string;
}

export interface SetReactionResponse {
  my_reaction: string;
}

export interface CommunityEventParams {
  title: string;
  description: string;
  location: string;
  /** datetime-local value, e.g. "2026-06-10T14:30" */
  startTime: string;
  endTime: string;
  banner?: File | null;
  inviteeEmails?: string[];
  inviteeUserIds?: number[];
  clearBanner?: boolean;
}

export interface InviteeCandidate {
  user_id: number;
  username: string;
  name: string;
  avatar_url?: string;
}

// datetime-local "2026-06-10T14:30" -> backend layout "2026-06-10 14:30"
function toBackendTime(value: string): string {
  return value.includes('T') ? value.replace('T', ' ') : value;
}

function buildCommunityEventFormData(params: CommunityEventParams): FormData {
  const formData = new FormData();
  formData.append('title', params.title);
  formData.append('description', params.description);
  formData.append('location', params.location);
  formData.append('start_time', toBackendTime(params.startTime));
  formData.append('end_time', toBackendTime(params.endTime));
  (params.inviteeEmails ?? []).forEach((email) => formData.append('invitee_emails', email));
  (params.inviteeUserIds ?? []).forEach((id) => formData.append('invitee_user_ids', String(id)));
  if (params.clearBanner) {
    formData.append('clear_banner', 'true');
  }
  if (params.banner) {
    formData.append('banner', params.banner);
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
      typeof error === 'object' && error !== null && 'message' in error && typeof error.message === 'string'
        ? error.message
        : 'Request failed';
    throw new Error(message);
  }
  return res.json() as Promise<T>;
}

// ─── Events ─────────────────────────────────────────────────────────────────

export function getCommunityEvents(
  communityId: number | string,
  token: string
): Promise<PaginatedCommunityEvents> {
  return api.get<PaginatedCommunityEvents>(`/api/communities/${communityId}/community-events`, token);
}

export function getCommunityEvent(
  eventId: number | string,
  token: string
): Promise<CommunityEventDetail> {
  return api.get<CommunityEventDetail>(`/api/community-events/${eventId}`, token);
}

export async function searchInviteeCandidates(
  communityId: number | string,
  query: string,
  token: string
): Promise<InviteeCandidate[]> {
  const res = await api.get<{ data: InviteeCandidate[] | null }>(
    `/api/communities/${communityId}/community-events/invitee-candidates?q=${encodeURIComponent(query)}`,
    token
  );
  return res.data ?? [];
}

export function createCommunityEvent(
  communityId: number | string,
  params: CommunityEventParams,
  token: string
): Promise<CommunityEventDetail> {
  return api.post<CommunityEventDetail>(
    `/api/communities/${communityId}/community-events`,
    buildCommunityEventFormData(params),
    token
  );
}

export function updateCommunityEvent(
  eventId: number | string,
  params: CommunityEventParams,
  token: string
): Promise<CommunityEventDetail> {
  return putFormData<CommunityEventDetail>(
    `/api/community-events/${eventId}`,
    buildCommunityEventFormData(params),
    token
  );
}

export function deleteCommunityEvent(
  eventId: number | string,
  token: string
): Promise<CommunityEventMessageResponse> {
  return api.delete<CommunityEventMessageResponse>(`/api/community-events/${eventId}`, token);
}

// ─── Comments ─────────────────────────────────────────────────────────────────

export function getCommunityEventComments(
  eventId: number | string,
  token: string
): Promise<PaginatedCommunityEventComments> {
  return api.get<PaginatedCommunityEventComments>(`/api/community-events/${eventId}/comments`, token);
}

export function addCommunityEventComment(
  eventId: number | string,
  content: string,
  token: string
): Promise<CommunityEventCommentData> {
  return api.post<CommunityEventCommentData>(
    `/api/community-events/${eventId}/comments`,
    { content },
    token
  );
}

export function editCommunityEventComment(
  eventId: number | string,
  commentId: number | string,
  content: string,
  token: string
): Promise<CommunityEventCommentData> {
  return api.put<CommunityEventCommentData>(
    `/api/community-events/${eventId}/comments/${commentId}`,
    { content },
    token
  );
}

export function deleteCommunityEventComment(
  eventId: number | string,
  commentId: number | string,
  token: string
): Promise<CommunityEventMessageResponse> {
  return api.delete<CommunityEventMessageResponse>(
    `/api/community-events/${eventId}/comments/${commentId}`,
    token
  );
}

export function pinCommunityEventComment(
  eventId: number | string,
  commentId: number | string,
  token: string
): Promise<CommunityEventMessageResponse> {
  return api.post<CommunityEventMessageResponse>(
    `/api/community-events/${eventId}/comments/${commentId}/pin`,
    undefined,
    token
  );
}

export function unpinCommunityEventComment(
  eventId: number | string,
  commentId: number | string,
  token: string
): Promise<CommunityEventMessageResponse> {
  return api.delete<CommunityEventMessageResponse>(
    `/api/community-events/${eventId}/comments/${commentId}/pin`,
    token
  );
}

// ─── Reactions ────────────────────────────────────────────────────────────────

export function setCommunityEventReaction(
  eventId: number | string,
  emoji: string,
  token: string
): Promise<SetReactionResponse> {
  return api.put<SetReactionResponse>(`/api/community-events/${eventId}/reaction`, { emoji }, token);
}

export function removeCommunityEventReaction(
  eventId: number | string,
  token: string
): Promise<CommunityEventMessageResponse> {
  return api.delete<CommunityEventMessageResponse>(`/api/community-events/${eventId}/reaction`, token);
}
