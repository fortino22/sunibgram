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

interface RawMyTrendingCurrent {
  userId: number;
  username: string;
  name: string;
  profile: string;
  is_premium: boolean;
  rank: number | null;
  score: number;
  detail: RawTrendingUser['detail'];
  meta: RawTrendingUsersMeta;
}

export interface MyTrendingCurrent {
  userId: number;
  username: string;
  name: string;
  profile: string;
  isPremium: boolean;
  rank: number | null;
  score: number;
  received: TrendingUserReceived;
  given: TrendingUserGiven;
  meta: TrendingUsersMeta;
}

export async function getMyTrendingCurrent(token: string): Promise<MyTrendingCurrent> {
  const raw = await api.get<RawMyTrendingCurrent>('/api/trending-users/my-current', token);
  return {
    userId: raw.userId,
    username: raw.username,
    name: raw.name,
    profile: raw.profile,
    isPremium: raw.is_premium,
    rank: raw.rank,
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

interface RawMyTrendingHistoryItem {
  snapshot_id: number;
  week_start: string;
  rank: number;
  score: number;
  detail: RawTrendingUser['detail'];
}

interface RawMyTrendingHistoryMonth {
  year: number;
  month: number;
}

interface RawMyTrendingHistoryMeta {
  year: number;
  month: number;
  available_months: RawMyTrendingHistoryMonth[];
}

interface RawMyTrendingHistoryResponse {
  data: RawMyTrendingHistoryItem[];
  meta: RawMyTrendingHistoryMeta;
}

export interface MyTrendingHistoryItem {
  snapshotId: number;
  weekStart: string;
  rank: number;
  score: number;
  received: TrendingUserReceived;
  given: TrendingUserGiven;
}

export interface MyTrendingHistoryMonth {
  year: number;
  month: number;
}

export interface MyTrendingHistoryResult {
  data: MyTrendingHistoryItem[];
  year: number;
  month: number;
  availableMonths: MyTrendingHistoryMonth[];
}

export async function getMyTrendingHistory(
  token: string,
  year?: number,
  month?: number,
): Promise<MyTrendingHistoryResult> {
  const qs = new URLSearchParams();
  if (year !== undefined && month !== undefined) {
    qs.set('year', String(year));
    qs.set('month', String(month));
  }
  const path = `/api/trending-users/my-history${qs.toString() ? `?${qs}` : ''}`;
  const raw = await api.get<RawMyTrendingHistoryResponse>(path, token);
  const items = (raw.data ?? []).map<MyTrendingHistoryItem>((r) => ({
    snapshotId: r.snapshot_id,
    weekStart: r.week_start,
    rank: r.rank,
    score: r.score,
    received: {
      likes: r.detail.received.likes,
      comments: r.detail.received.comments,
      bookmarks: r.detail.received.bookmarks,
      reposts: r.detail.received.reposts,
      threadCount: r.detail.received.thread_count,
    },
    given: {
      likes: r.detail.given.likes,
      bookmarks: r.detail.given.bookmarks,
      comments: r.detail.given.comments,
      posts: r.detail.given.posts,
    },
  }));
  return {
    data: items,
    year: raw.meta.year,
    month: raw.meta.month,
    availableMonths: (raw.meta.available_months ?? []).map((m) => ({
      year: m.year,
      month: m.month,
    })),
  };
}

interface RawMyTrendingSnapshot {
  snapshot_id: number;
  week_start: string;
  rank: number;
  score: number;
  snapshot_at: string;
  detail: RawTrendingUser['detail'];
}

export interface MyTrendingSnapshot {
  snapshotId: number;
  weekStart: string;
  rank: number;
  score: number;
  snapshotAt: string;
  received: TrendingUserReceived;
  given: TrendingUserGiven;
}

export async function getMyTrendingSnapshot(
  token: string,
  snapshotId: number,
): Promise<MyTrendingSnapshot> {
  const raw = await api.get<RawMyTrendingSnapshot>(
    `/api/trending-users/snapshots/${snapshotId}`,
    token,
  );
  return {
    snapshotId: raw.snapshot_id,
    weekStart: raw.week_start,
    rank: raw.rank,
    score: raw.score,
    snapshotAt: raw.snapshot_at,
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
