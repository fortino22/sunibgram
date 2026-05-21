import { api } from '../api';

export interface NotificationFaculty {
  faculty_id: number;
  faculty_name: string;
}

export interface NotificationActor {
  user_id: number;
  username: string;
  name: string;
  avatar_url: string;
  is_premium: boolean;
  faculty?: NotificationFaculty | null;
}

export interface NotificationItem {
  id: number;
  type: string;
  message: string;
  reference_id: number | null;
  created_at: string;
  actor: NotificationActor;
}

export interface NotificationMeta {
  limit: number;
  offset: number;
}

export interface PaginatedNotificationResponse {
  data: NotificationItem[];
  meta: NotificationMeta;
}

export function getNotifications(
  token: string,
  limit = 20,
  offset = 0,
): Promise<PaginatedNotificationResponse> {
  return api.get<PaginatedNotificationResponse>(
    `/api/notifications?limit=${limit}&offset=${offset}`,
    token,
  );
}

export function deleteNotification(id: number, token: string): Promise<{ message: string }> {
  return api.delete<{ message: string }>(`/api/notifications/${id}`, token);
}
