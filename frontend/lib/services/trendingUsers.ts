import { api } from '../api';

interface RawTrendingUser {
  userId: number;
  username: string;
  name: string;
  profile: string;
  is_premium: boolean;
  is_following: boolean;
  score: number;
  detail: {
    received: {
      likes: number;
      comments: number;
      bookmarks: number;
      reposts: number;
      thread_count: number;
    };
    given: {
      likes: number;
      bookmarks: number;
      comments: number;
      posts: number;
    };
  };
}

interface RawTrendingUsersMeta {
  total: number;
  page: number;
  limit: number;
  cached_at: string;
  refresh_at: string;
  refresh_in_seconds: number;
}

interface RawTrendingUsersResponse {
  data: RawTrendingUser[];
  meta: RawTrendingUsersMeta;
}

export interface TrendingUserReceived {
  likes: number;
  comments: number;
  bookmarks: number;
  reposts: number;
  threadCount: number;
}

export interface TrendingUserGiven {
  likes: number;
  bookmarks: number;
  comments: number;
  posts: number;
}

export interface TrendingUser {
  userId: number;
  username: string;
  name: string;
  profile: string;
  isPremium: boolean;
  isFollowing: boolean;
  score: number;
  received: TrendingUserReceived;
  given: TrendingUserGiven;
}

export interface TrendingUsersMeta {
  total: number;
  page: number;
  limit: number;
  cachedAt: string;
  refreshAt: string;
  refreshInSeconds: number;
}

export interface TrendingUsersResponse {
  data: TrendingUser[];
  meta: TrendingUsersMeta;
}

function mapTrendingUser(raw: RawTrendingUser): TrendingUser {
  return {
    userId: raw.userId,
    username: raw.username,
    name: raw.name,
    profile: raw.profile,
    isPremium: raw.is_premium,
    isFollowing: raw.is_following,
    score: raw.score,
    received: {
      likes: raw.detail.received.likes,
      comments: raw.detail.received.comments,
      bookmarks: raw.detail.received.bookmarks,
      reposts: raw.detail.received.reposts,
      threadCount: raw.detail.received.thread_count,
    },
    given: {
      likes: raw.detail.given.likes,
      bookmarks: raw.detail.given.bookmarks,
      comments: raw.detail.given.comments,
      posts: raw.detail.given.posts,
    },
  };
}

export async function getTrendingUsers(
  token: string,
  page = 1,
  limit = 20,
): Promise<TrendingUsersResponse> {
  const raw = await api.get<RawTrendingUsersResponse>(
    `/api/trending-users?page=${page}&limit=${limit}`,
    token,
  );
  return {
    data: (raw.data ?? []).map(mapTrendingUser),
    meta: {
      total: raw.meta.total,
      page: raw.meta.page,
      limit: raw.meta.limit,
      cachedAt: raw.meta.cached_at,
      refreshAt: raw.meta.refresh_at,
      refreshInSeconds: raw.meta.refresh_in_seconds,
    },
  };
}
