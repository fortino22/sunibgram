import { api } from '../api';
import type { FeedMeta } from './threads';

export interface MuteMessageResponse {
  message: string;
}

export interface MutedUser {
  id: number;
  name: string;
  username: string;
  avatar: string | null;
  is_premium: boolean;
}

interface BackendMutedUser {
  user_id: number;
  name: string;
  username: string;
  is_premium: boolean;
  avatar_url: string;
}

interface PaginatedMutedUsersResponse {
  data: BackendMutedUser[];
  meta: FeedMeta;
}

function mapMutedUser(raw: BackendMutedUser): MutedUser {
  return {
    id: raw.user_id,
    name: raw.name,
    username: raw.username,
    avatar: raw.avatar_url || null,
    is_premium: raw.is_premium,
  };
}

export async function getMutedUsers(
  token: string,
  page = 1,
  limit = 20
): Promise<MutedUser[]> {
  const res = await api.get<PaginatedMutedUsersResponse>(
    `/api/users/me/mutes?page=${page}&limit=${limit}`,
    token
  );
  return (res.data ?? []).map(mapMutedUser);
}

export function muteUser(userId: number | string, token: string): Promise<MuteMessageResponse> {
  return api.post<MuteMessageResponse>(`/api/users/${userId}/mute`, undefined, token);
}

export function unmuteUser(userId: number | string, token: string): Promise<MuteMessageResponse> {
  return api.delete<MuteMessageResponse>(`/api/users/${userId}/unmute`, token);
}
