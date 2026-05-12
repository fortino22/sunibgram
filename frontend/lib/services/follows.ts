import { api } from '../api';

export interface FollowMessageResponse {
  message: string;
}

export interface FollowUser {
  user_id: number;
  name?: string;
  username: string;
  is_premium?: boolean;
  avatar_url?: string | null;
}

export interface FollowListResponse {
  data: FollowUser[];
  meta: { total: number; page: number; limit: number };
}

export interface FollowListParams {
  page?: number;
  limit?: number;
}

function query(params?: FollowListParams): string {
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

export function getFollowers(
  userId: number | string,
  token: string,
  params?: FollowListParams
): Promise<FollowListResponse> {
  return api.get<FollowListResponse>(`/api/follows/${userId}/followers${query(params)}`, token);
}

export function getFollowing(
  userId: number | string,
  token: string,
  params?: FollowListParams
): Promise<FollowListResponse> {
  return api.get<FollowListResponse>(`/api/follows/${userId}/following${query(params)}`, token);
}

export function followUser(userId: number | string, token: string): Promise<FollowMessageResponse> {
  return api.post<FollowMessageResponse>(`/api/users/${userId}/follow`, undefined, token);
}

export function unfollowUser(userId: number | string, token: string): Promise<FollowMessageResponse> {
  return api.delete<FollowMessageResponse>(`/api/users/${userId}/unfollow`, token);
}
