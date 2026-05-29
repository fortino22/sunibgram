import { api } from '../api';

export interface ShareUser {
  user_id: number;
  name: string;
  username: string;
  is_premium: boolean;
  avatar_url?: string | null;
}

interface SearchShareUsersResponse {
  data: ShareUser[] | null;
}

export interface CopyShareResponse {
  message: string;
  share_count: number;
}

export interface ChatShareResponse {
  message: string;
  recipient_count: number;
}

export async function searchShareUsers(query: string, token: string): Promise<ShareUser[]> {
  const trimmed = query.trim();
  if (!trimmed) return [];
  const raw = await api.get<SearchShareUsersResponse>(
    `/api/users/search-share?q=${encodeURIComponent(trimmed)}`,
    token,
  );
  return raw.data ?? [];
}

export function recordShareCopy(threadId: number | string, token: string): Promise<CopyShareResponse> {
  return api.post<CopyShareResponse>(`/api/threads/${threadId}/share/copy`, undefined, token);
}

export interface ShareViaChatParams {
  threadId: number | string;
  userIds: number[];
  message?: string;
}

export function shareViaChat(params: ShareViaChatParams, token: string): Promise<ChatShareResponse> {
  const body: { user_ids: number[]; message?: string } = { user_ids: params.userIds };
  const trimmed = params.message?.trim();
  if (trimmed) body.message = trimmed;
  return api.post<ChatShareResponse>(`/api/threads/${params.threadId}/share/chat`, body, token);
}
