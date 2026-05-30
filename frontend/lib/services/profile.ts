import { api } from '../api';

export interface ProfileMessageResponse {
  message: string;
}

// Exact shape the backend returns for profile endpoints
interface BackendProfileResponse {
  user_id: number;
  name?: string;
  username: string;
  email?: string;
  is_premium?: boolean;
  bio?: string;
  avatar_url?: string | null;
  banner_url?: string | null;
  follower_count?: number;
  following_count?: number;
  faculty?: { faculty_id: number; faculty_name: string } | null;
  birthday_date?: string;
  created_at?: string;
  is_following?: boolean;
  is_blocked?: boolean;
  is_blocked_by?: boolean;
}

// Frontend model — used throughout components
export interface UserProfile {
  id: number;
  name?: string;
  username: string;
  email?: string;
  is_premium?: boolean;
  avatar?: string | null;
  banner?: string | null;
  bio?: string | null;
  faculty_id?: number | null;
  faculty?: { faculty_id: number; faculty_name: string } | null;
  birthday_date?: string | null;
  followers_count?: number;
  following_count?: number;
  is_following?: boolean;
  is_blocked?: boolean;
  is_blocked_by?: boolean;
  created_at?: string;
}

function mapProfile(raw: BackendProfileResponse): UserProfile {
  return {
    id: raw.user_id,
    name: raw.name,
    username: raw.username,
    email: raw.email,
    is_premium: raw.is_premium,
    avatar: raw.avatar_url ?? null,
    banner: raw.banner_url ?? null,
    bio: raw.bio ?? null,
    faculty_id: raw.faculty?.faculty_id ?? null,
    faculty: raw.faculty ?? null,
    birthday_date: raw.birthday_date ?? null,
    followers_count: raw.follower_count ?? 0,
    following_count: raw.following_count ?? 0,
    is_following: raw.is_following,
    is_blocked: raw.is_blocked,
    is_blocked_by: raw.is_blocked_by,
    created_at: raw.created_at,
  };
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
  birthdayDate?: string;
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

export async function getMyProfile(token: string): Promise<UserProfile> {
  const raw = await api.get<BackendProfileResponse>('/api/profile/me', token);
  return mapProfile(raw);
}

export async function getProfileByUserId(userId: number | string, token: string): Promise<UserProfile> {
  const raw = await api.get<BackendProfileResponse>(`/api/profile/${userId}`, token);
  return mapProfile(raw);
}

export const getProfile = getProfileByUserId;

export async function updateMyProfile(
  params: UpdateProfileParams,
  token: string
): Promise<UserProfile> {
  const formData = new FormData();
  appendOptionalString(formData, 'name', params.name);
  appendOptionalString(formData, 'username', params.username);
  appendOptionalString(formData, 'bio', params.bio);
  appendOptionalString(formData, 'faculty_id', params.facultyId);
  appendOptionalString(formData, 'clear_avatar', params.clearAvatar);
  appendOptionalString(formData, 'clear_banner', params.clearBanner);
  appendOptionalString(formData, 'birthday_date', params.birthdayDate);

  if (params.avatar) {
    formData.append('avatar', params.avatar);
  }
  if (params.banner) {
    formData.append('banner', params.banner);
  }

  const raw = await putFormData<BackendProfileResponse>('/api/profile/me', formData, token);
  return mapProfile(raw);
}
