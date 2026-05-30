import { api } from '../api';

export interface BlockMessageResponse {
  message: string;
}

export function blockUser(userId: number | string, token: string): Promise<BlockMessageResponse> {
  return api.post<BlockMessageResponse>(`/api/users/${userId}/block`, undefined, token);
}

export function unblockUser(userId: number | string, token: string): Promise<BlockMessageResponse> {
  return api.delete<BlockMessageResponse>(`/api/users/${userId}/unblock`, token);
}
