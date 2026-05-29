'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addThreadToBookmarkCollection,
  createBookmarkCollection,
  deleteBookmarkCollection,
  getBookmarkCollection,
  listBookmarkCollectionThreads,
  listBookmarkCollections,
  removeThreadFromBookmarkCollection,
  updateBookmarkCollection,
  type BookmarkCollection,
  type BookmarkCollectionListResponse,
  type BookmarkCollectionThreadsResponse,
  type CreateBookmarkCollectionInput,
  type UpdateBookmarkCollectionInput,
} from '@/lib/services/bookmarkCollections';
import { useAuthStore } from '@/stores/authStore';

const COLLECTIONS_KEY = ['bookmark-collections'] as const;

export function bookmarkCollectionsKey() {
  return COLLECTIONS_KEY;
}

export function bookmarkCollectionKey(id: number | string) {
  return ['bookmark-collections', String(id)] as const;
}

export function bookmarkCollectionThreadsKey(id: number | string) {
  return ['bookmark-collections', String(id), 'threads'] as const;
}

export function useBookmarkCollections() {
  const token = useAuthStore((s) => s.token);

  return useQuery<BookmarkCollectionListResponse>({
    queryKey: COLLECTIONS_KEY,
    queryFn: () => listBookmarkCollections(token!),
    enabled: !!token,
  });
}

export function useBookmarkCollection(id: number | string | null | undefined) {
  const token = useAuthStore((s) => s.token);

  return useQuery<BookmarkCollection>({
    queryKey: bookmarkCollectionKey(id ?? ''),
    queryFn: () => getBookmarkCollection(id!, token!),
    enabled: !!token && id !== null && id !== undefined && id !== '',
  });
}

export function useBookmarkCollectionThreads(id: number | string | null | undefined) {
  const token = useAuthStore((s) => s.token);

  return useQuery<BookmarkCollectionThreadsResponse>({
    queryKey: bookmarkCollectionThreadsKey(id ?? ''),
    queryFn: () => listBookmarkCollectionThreads(id!, { page: 1, limit: 20 }, token!),
    enabled: !!token && id !== null && id !== undefined && id !== '',
  });
}

export function useCreateBookmarkCollection() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateBookmarkCollectionInput) => {
      if (!token) throw new Error('Authentication required to create a collection.');
      return createBookmarkCollection(input, token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: COLLECTIONS_KEY });
    },
  });
}

export interface UpdateBookmarkCollectionVars {
  id: number | string;
  input: UpdateBookmarkCollectionInput;
}

export function useUpdateBookmarkCollection() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: UpdateBookmarkCollectionVars) => {
      if (!token) throw new Error('Authentication required to update this collection.');
      return updateBookmarkCollection(id, input, token);
    },
    onSuccess: (_data, vars) => {
      queryClient.invalidateQueries({ queryKey: COLLECTIONS_KEY });
      queryClient.invalidateQueries({ queryKey: bookmarkCollectionKey(vars.id) });
    },
  });
}

export function useDeleteBookmarkCollection() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number | string) => {
      if (!token) throw new Error('Authentication required to delete this collection.');
      return deleteBookmarkCollection(id, token);
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: COLLECTIONS_KEY });
      queryClient.removeQueries({ queryKey: bookmarkCollectionKey(id) });
      queryClient.removeQueries({ queryKey: bookmarkCollectionThreadsKey(id) });
    },
  });
}

export interface CollectionThreadVars {
  collectionId: number | string;
  threadId: number;
}

export function useAddThreadToBookmarkCollection() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ collectionId, threadId }: CollectionThreadVars) => {
      if (!token) throw new Error('Authentication required to update this collection.');
      return addThreadToBookmarkCollection(collectionId, threadId, token);
    },
    onSuccess: (_data, { collectionId }) => {
      queryClient.invalidateQueries({ queryKey: COLLECTIONS_KEY });
      queryClient.invalidateQueries({ queryKey: bookmarkCollectionKey(collectionId) });
      queryClient.invalidateQueries({ queryKey: bookmarkCollectionThreadsKey(collectionId) });
    },
  });
}

export function useRemoveThreadFromBookmarkCollection() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ collectionId, threadId }: CollectionThreadVars) => {
      if (!token) throw new Error('Authentication required to update this collection.');
      return removeThreadFromBookmarkCollection(collectionId, threadId, token);
    },
    onSuccess: (_data, { collectionId }) => {
      queryClient.invalidateQueries({ queryKey: COLLECTIONS_KEY });
      queryClient.invalidateQueries({ queryKey: bookmarkCollectionKey(collectionId) });
      queryClient.invalidateQueries({ queryKey: bookmarkCollectionThreadsKey(collectionId) });
    },
  });
}
