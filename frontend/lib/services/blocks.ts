import { api } from '../api';
import type { FeedMeta } from './threads';

export interface BlockMessageResponse {
  message: string;
}

export interface BlockedUser {
  id: number;
  name: string;
  username: string;
  avatar: string | null;
  is_premium: boolean;
}

interface BackendBlockedUser {
  user_id: number;
  name: string;
  username: string;
  is_premium: boolean;
  avatar_url: string;
}

interface PaginatedBlockedUsersResponse {
  data: BackendBlockedUser[];
  meta: FeedMeta;
}

function mapBlockedUser(raw: BackendBlockedUser): BlockedUser {
  return {
    id: raw.user_id,
    name: raw.name,
    username: raw.username,
    avatar: raw.avatar_url || null,
    is_premium: raw.is_premium,
  };
}

export async function getBlockedUsers(
  token: string,
  page = 1,
  limit = 20
): Promise<BlockedUser[]> {
  const res = await api.get<PaginatedBlockedUsersResponse>(
    `/api/users/me/blocks?page=${page}&limit=${limit}`,
    token
  );
  return (res.data ?? []).map(mapBlockedUser);
}

export function blockUser(userId: number | string, token: string): Promise<BlockMessageResponse> {
  return api.post<BlockMessageResponse>(`/api/users/${userId}/block`, undefined, token);
}

export function unblockUser(userId: number | string, token: string): Promise<BlockMessageResponse> {
  return api.delete<BlockMessageResponse>(`/api/users/${userId}/unblock`, token);
}
