import { api } from '../api';
import type { PaginatedFeedResponse } from './threads';

// Raw backend shapes
interface RawExploreUser {
  userId: number;
  profile: string;
  username: string;
  name: string;
  is_premium: boolean;
  is_following: boolean;
}

interface RawExplorePerson {
  userId: number;
  user_profile_pic: string;
  username: string;
  name: string;
  bio?: string;
  is_premium: boolean;
  is_following: boolean;
}

interface RawExploreCommunity {
  community_id: number;
  banner_url: string;
  community_name: string;
  member_count: number;
  is_join: boolean;
}

// Frontend types
export interface ExploreUser {
  userId: number;
  profile: string;
  username: string;
  name: string;
  isPremium: boolean;
  isFollowing: boolean;
}

export interface ExplorePerson {
  userId: number;
  userProfilePic: string;
  username: string;
  name: string;
  bio?: string;
  isPremium: boolean;
  isFollowing: boolean;
}

export interface ExploreSearchResult {
  users: ExploreUser[];
  threads: { content: string }[];
}

export interface ExploreCommunity {
  communityId: number;
  banner: string;
  communityName: string;
  memberCount: number;
  isJoin: boolean;
}

function mapPerson(raw: RawExplorePerson): ExplorePerson {
  return {
    userId: raw.userId,
    userProfilePic: raw.user_profile_pic,
    username: raw.username,
    name: raw.name,
    bio: raw.bio ?? '',
    isPremium: raw.is_premium,
    isFollowing: raw.is_following,
  };
}

function mapCommunity(raw: RawExploreCommunity): ExploreCommunity {
  return {
    communityId: raw.community_id,
    banner: raw.banner_url ?? '',
    communityName: raw.community_name,
    memberCount: raw.member_count,
    isJoin: raw.is_join,
  };
}

export async function searchExplore(q: string, token: string): Promise<ExploreSearchResult> {
  const raw = await api.get<{ users: RawExploreUser[]; threads: { content: string }[] }>(
    `/api/explore/search?q=${encodeURIComponent(q)}`,
    token
  );
  return {
    users: raw.users.map((u) => ({
      userId: u.userId,
      profile: u.profile,
      username: u.username,
      name: u.name,
      isPremium: u.is_premium,
      isFollowing: u.is_following,
    })),
    threads: raw.threads,
  };
}

export function getTopPosts(q: string, token: string): Promise<PaginatedFeedResponse> {
  return api.get<PaginatedFeedResponse>(`/api/explore/top?q=${encodeURIComponent(q)}`, token);
}

export function getLatestPosts(q: string, token: string): Promise<PaginatedFeedResponse> {
  return api.get<PaginatedFeedResponse>(`/api/explore/latest?q=${encodeURIComponent(q)}`, token);
}

export async function getPeople(q: string, token: string): Promise<ExplorePerson[]> {
  const raw = await api.get<{ data: RawExplorePerson[] }>(
    `/api/explore/people?q=${encodeURIComponent(q)}`,
    token
  );
  return raw.data.map(mapPerson);
}

export function getMediaPosts(q: string, token: string): Promise<PaginatedFeedResponse> {
  return api.get<PaginatedFeedResponse>(`/api/explore/media?q=${encodeURIComponent(q)}`, token);
}

export async function getCommunitySearch(q: string, token: string): Promise<ExploreCommunity[]> {
  const raw = await api.get<{ data: RawExploreCommunity[] }>(
    `/api/explore/community?q=${encodeURIComponent(q)}`,
    token
  );
  return raw.data.map(mapCommunity);
}

export async function getSuggestedPeople(token: string, limit = 10): Promise<ExplorePerson[]> {
  const raw = await api.get<{ data: RawExplorePerson[] }>(
    `/api/users/suggested?limit=${limit}`,
    token
  );
  return raw.data.map(mapPerson);
}

// Onboarding follow gate. Empty q → ranked suggestions; q → people search.
// Either way ≤10, self- and admin-excluded (server-enforced).
export async function getOnboardingPeople(token: string, q = '', limit = 10): Promise<ExplorePerson[]> {
  const raw = await api.get<{ data: RawExplorePerson[] }>(
    `/api/users/onboarding/suggestions?q=${encodeURIComponent(q)}&limit=${limit}`,
    token
  );
  return (raw.data ?? []).map(mapPerson);
}

export async function getSuggestedPeoplePreview(token: string): Promise<ExplorePerson[]> {
  const raw = await api.get<{ data: RawExplorePerson[] }>('/api/users/suggested/preview', token);
  return (raw.data ?? []).map(mapPerson);
}

export async function getMutualSuggestions(token: string, page: number, limit = 10): Promise<ExplorePerson[]> {
  const raw = await api.get<{ data: RawExplorePerson[] }>(
    `/api/users/suggestions/mutual?page=${page}&limit=${limit}`,
    token
  );
  return (raw.data ?? []).map(mapPerson);
}

export async function getFacultySuggestions(token: string, page: number, limit = 10): Promise<ExplorePerson[]> {
  const raw = await api.get<{ data: RawExplorePerson[] }>(
    `/api/users/suggestions/faculty?page=${page}&limit=${limit}`,
    token
  );
  return (raw.data ?? []).map(mapPerson);
}
