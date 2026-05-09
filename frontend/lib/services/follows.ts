import { api } from '../api';

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

export function getFollowers(userId: number | string, token: string): Promise<FollowListResponse> {
  return api.get<FollowListResponse>(`/api/follows/${userId}/followers`, token);
}

export function getFollowing(userId: number | string, token: string): Promise<FollowListResponse> {
  return api.get<FollowListResponse>(`/api/follows/${userId}/following`, token);
}
