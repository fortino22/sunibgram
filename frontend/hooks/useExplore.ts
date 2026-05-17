'use client';

import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import {
  getTrendingHashtags,
  getSuggestedPeople,
  searchExplore,
  getTopPosts,
  getLatestPosts,
  getPeople,
  getMediaPosts,
  getCommunitySearch,
} from '@/lib/services/explore';

export function useTrendingHashtags(limit = 10) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['hashtags', 'trending', limit],
    queryFn: () => getTrendingHashtags(token, limit),
    staleTime: 5 * 60 * 1000,
  });
}

export function useSuggestedPeople(limit = 10) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'suggested-people', limit],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getSuggestedPeople(token, limit);
    },
    enabled: token !== null,
  });
}

export function useExploreSearch(q: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'search', q],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return searchExplore(q, token);
    },
    enabled: token !== null && q.trim().length > 0,
  });
}

export function useExploreTop(q: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'top', q],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getTopPosts(q, token);
    },
    enabled: token !== null && q.trim().length > 0,
  });
}

export function useExploreLatest(q: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'latest', q],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getLatestPosts(q, token);
    },
    enabled: token !== null && q.trim().length > 0,
  });
}

export function useExplorePeople(q: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'people', q],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getPeople(q, token);
    },
    enabled: token !== null && q.trim().length > 0,
    placeholderData: keepPreviousData,
  });
}

export function useExploreMedia(q: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'media', q],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getMediaPosts(q, token);
    },
    enabled: token !== null && q.trim().length > 0,
  });
}

export function useExploreCommunities(q: string) {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['explore', 'communities', q],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getCommunitySearch(q, token);
    },
    enabled: token !== null && q.trim().length > 0,
  });
}
