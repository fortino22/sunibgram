'use client';

import { useQuery } from '@tanstack/react-query';
import {
  getCommunity,
  getMyCommunities,
  getJoinedCommunities,
  getDiscoverCommunities,
  getCommunityMembers,
  getCommunityJoinRequests,
  getCommunityThreadsLatest,
  getCommunityThreadsTop,
  getCommunityThreadsMedia,
  getCategories,
  getMyCommunitiesThreadsFeed,
  getExploreCommunitiesThreads,
  searchCommunities,
  searchCommunityThreadsTop,
  searchCommunityThreadsLatest,
  getCommunityAbout,
  getCommunityUsers,
  getPinnedCommunities,
  getCommunityThreadsBySort,
} from '@/lib/services/communities';
import { useAuthStore } from '@/stores/authStore';

export function useCommunity(id: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', id],
    queryFn: () => getCommunity(id!, token!),
    enabled: !!token && !!id,
  });
}

export function useDiscoverCommunities(categoryId?: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'discover', categoryId ?? null],
    queryFn: () => getDiscoverCommunities(token!, categoryId),
    enabled: !!token,
    staleTime: 5 * 60 * 1000,
  });
}

export function useMyCommunities() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'mine'],
    queryFn: () => getMyCommunities(token!, { limit: 50 }),
    enabled: !!token,
  });
}

export function useJoinedCommunities() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'joined'],
    queryFn: () => getJoinedCommunities(token!, { limit: 50 }),
    enabled: !!token,
  });
}

export function useCommunityMembers(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'members'],
    queryFn: () => getCommunityMembers(communityId!, token!, { limit: 50 }),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityRequests(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'requests'],
    queryFn: () => getCommunityJoinRequests(communityId!, token!, { limit: 50 }),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityThreadsLatest(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'threads', 'latest'],
    queryFn: () => getCommunityThreadsLatest(communityId!, token!, { limit: 20 }),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityThreadsTop(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'threads', 'top'],
    queryFn: () => getCommunityThreadsTop(communityId!, token!, { limit: 20 }),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityThreadsMedia(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'threads', 'media'],
    queryFn: () => getCommunityThreadsMedia(communityId!, token!, { limit: 20 }),
    enabled: !!token && !!communityId,
  });
}

export function useCategories() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['categories'],
    queryFn: () => getCategories(token ?? undefined),
    staleTime: 10 * 60 * 1000,
  });
}

export function useMyCommunitiesThreadsFeed() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'threads', 'feed'],
    queryFn: () => getMyCommunitiesThreadsFeed(token!, { limit: 20 }),
    enabled: !!token,
    staleTime: 2 * 60 * 1000,
  });
}

export function useExploreCommunitiesThreads(categoryId?: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'threads', 'explore', categoryId ?? null],
    queryFn: () => getExploreCommunitiesThreads(token!, categoryId, { limit: 20 }),
    enabled: !!token,
    staleTime: 2 * 60 * 1000,
  });
}

export function useSearchCommunities(query: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'search', query],
    queryFn: () => searchCommunities(query, token!),
    enabled: !!token && query.trim().length > 0,
    staleTime: 60 * 1000,
  });
}

export function useSearchCommunityThreadsTop(query: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'threads', 'search', 'top', query],
    queryFn: () => searchCommunityThreadsTop(token!, query, { limit: 20 }),
    enabled: !!token && query.trim().length > 0,
    staleTime: 60 * 1000,
  });
}

export function useSearchCommunityThreadsLatest(query: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'threads', 'search', 'latest', query],
    queryFn: () => searchCommunityThreadsLatest(token!, query, { limit: 20 }),
    enabled: !!token && query.trim().length > 0,
    staleTime: 60 * 1000,
  });
}

export function useCommunityAbout(communityId: number | undefined) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'about'],
    queryFn: () => getCommunityAbout(communityId!, token!),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityUsers(
  communityId: number | undefined,
  params: { role?: string; query?: string } = {}
) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'users', params.role ?? null, params.query ?? ''],
    queryFn: () => getCommunityUsers(communityId!, { ...params, limit: 30 }, token!),
    enabled: !!token && !!communityId,
  });
}

export function useCommunityThreadsBySort(
  communityId: number | undefined,
  sort: 'latest' | 'top' | 'likes'
) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['community', communityId, 'threads', 'sort', sort],
    queryFn: () => getCommunityThreadsBySort(communityId!, sort, token!, { limit: 20 }),
    enabled: !!token && !!communityId,
  });
}

export function usePinnedCommunities() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['communities', 'pins'],
    queryFn: () => getPinnedCommunities(token!, { limit: 50 }),
    enabled: !!token,
    staleTime: 2 * 60 * 1000,
  });
}
