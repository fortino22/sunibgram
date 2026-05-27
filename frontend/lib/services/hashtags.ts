import { api } from '../api';

interface RawTrendingHashtag {
  name: string;
  thread_count: number;
}

export interface TrendingHashtag {
  name: string;
  threadCount: number;
}

function mapTrendingHashtag(raw: RawTrendingHashtag): TrendingHashtag {
  return {
    name: raw.name,
    threadCount: raw.thread_count,
  };
}

export async function getTrendingHashtags(token: string): Promise<TrendingHashtag[]> {
  const raw = await api.get<{ data: RawTrendingHashtag[] }>('/api/hashtags/trending', token);
  return (raw.data ?? []).map(mapTrendingHashtag);
}

/** Top-N trending hashtags (compact widget). Backend caps at 4. */
export async function getTopTrendingHashtags(token: string): Promise<TrendingHashtag[]> {
  const raw = await api.get<{ data: RawTrendingHashtag[] }>('/api/hashtags/trending/top', token);
  return (raw.data ?? []).map(mapTrendingHashtag);
}
