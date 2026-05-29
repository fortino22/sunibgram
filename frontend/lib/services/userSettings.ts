import { api } from '../api';

export interface UserSettings {
  user_id: number;
  is_private: boolean;
}

export interface MessageResponse {
  message: string;
}

export function getMyUserSettings(token: string): Promise<UserSettings> {
  return api.get<UserSettings>('/api/user-settings/me', token);
}

export function updateMyUserSettings(
  params: { is_private: boolean },
  token: string
): Promise<MessageResponse> {
  return api.put<MessageResponse>('/api/user-settings/me', params, token);
}
