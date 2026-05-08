import { api } from '../api';

export interface ProfileMessageResponse {
  message: string;
}

export interface UserProfile {
  id: number;
  name?: string;
  username: string;
  email?: string;
  avatar?: string | null;
  banner?: string | null;
  bio?: string | null;
  faculty_id?: number | null;
  faculty?: { faculty_id: number; faculty_name: string } | null;
  followers_count?: number;
  following_count?: number;
  is_following?: boolean;
  created_at?: string;
}

export interface ProfileResponse {
  data: UserProfile;
  message?: string;
  [key: string]: unknown;
}

export interface UpdateProfileParams {
  name?: string;
  username?: string;
  avatar?: File | null;
  banner?: File | null;
  bio?: string;
  facultyId?: number | string;
  clearAvatar?: boolean;
  clearBanner?: boolean;
}

function appendOptionalString(formData: FormData, key: string, value?: string | number | boolean): void {
  if (value !== undefined) {
    formData.append(key, String(value));
  }
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
      typeof error === 'object' &&
      error !== null &&
      'message' in error &&
      typeof error.message === 'string'
        ? error.message
        : 'Request failed';
    throw new Error(message);
  }

  return res.json() as Promise<T>;
}

export function getMyProfile(token: string): Promise<ProfileResponse> {
  return api.get<ProfileResponse>('/api/profile/me', token);
}

export function getProfile(userId: number | string, token: string): Promise<ProfileResponse> {
  return api.get<ProfileResponse>(`/api/profile/${userId}`, token);
}

export function updateMyProfile(
  params: UpdateProfileParams,
  token: string
): Promise<ProfileResponse> {
  const formData = new FormData();
  appendOptionalString(formData, 'name', params.name);
  appendOptionalString(formData, 'username', params.username);
  appendOptionalString(formData, 'bio', params.bio);
  appendOptionalString(formData, 'faculty_id', params.facultyId);
  appendOptionalString(formData, 'clear_avatar', params.clearAvatar);
  appendOptionalString(formData, 'clear_banner', params.clearBanner);

  if (params.avatar) {
    formData.append('avatar', params.avatar);
  }
  if (params.banner) {
    formData.append('banner', params.banner);
  }

  return putFormData<ProfileResponse>('/api/profile/me', formData, token);
}

export function followUser(userId: number | string, token: string): Promise<ProfileMessageResponse> {
  return api.post<ProfileMessageResponse>(`/api/users/${userId}/follow`, undefined, token);
}

export function unfollowUser(userId: number | string, token: string): Promise<ProfileMessageResponse> {
  return api.delete<ProfileMessageResponse>(`/api/users/${userId}/unfollow`, token);
}
