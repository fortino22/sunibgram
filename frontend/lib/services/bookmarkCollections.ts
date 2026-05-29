import { api } from '../api';
import type { FeedItemResponse, FeedMeta, FormattedTime, PaginationParams } from './threads';

export interface BookmarkCollection {
  id: number;
  name: string;
  description?: string | null;
  item_count: number;
  created_at: FormattedTime;
  updated_at: FormattedTime;
}

export interface BookmarkCollectionListResponse {
  data: BookmarkCollection[];
}

export interface BookmarkCollectionThreadsResponse {
  data: FeedItemResponse[];
  meta: FeedMeta;
}

export interface CreateBookmarkCollectionInput {
  name: string;
  description?: string | null;
}

export interface UpdateBookmarkCollectionInput {
  name: string;
  description?: string | null;
}

export interface CreateBookmarkCollectionResponse {
  collection_id: number;
}

export interface BookmarkCollectionMessageResponse {
  message: string;
}

export const MAX_COLLECTIONS_PER_USER = 50;
export const MAX_COLLECTION_NAME_LENGTH = 50;

function query(params?: PaginationParams): string {
  const search = new URLSearchParams();
  if (params?.page !== undefined) search.set('page', String(params.page));
  if (params?.limit !== undefined) search.set('limit', String(params.limit));
  const value = search.toString();
  return value ? `?${value}` : '';
}

export function listBookmarkCollections(token: string): Promise<BookmarkCollectionListResponse> {
  return api.get<BookmarkCollectionListResponse>('/api/bookmark-collections', token);
}

export function getBookmarkCollection(id: number | string, token: string): Promise<BookmarkCollection> {
  return api.get<BookmarkCollection>(`/api/bookmark-collections/${id}`, token);
}

export function createBookmarkCollection(
  input: CreateBookmarkCollectionInput,
  token: string
): Promise<CreateBookmarkCollectionResponse> {
  return api.post<CreateBookmarkCollectionResponse>(
    '/api/bookmark-collections',
    {
      name: input.name,
      description: input.description ?? undefined,
    },
    token
  );
}

export function updateBookmarkCollection(
  id: number | string,
  input: UpdateBookmarkCollectionInput,
  token: string
): Promise<BookmarkCollectionMessageResponse> {
  return api.patch<BookmarkCollectionMessageResponse>(
    `/api/bookmark-collections/${id}`,
    {
      name: input.name,
      description: input.description ?? undefined,
    },
    token
  );
}

export function deleteBookmarkCollection(
  id: number | string,
  token: string
): Promise<BookmarkCollectionMessageResponse> {
  return api.delete<BookmarkCollectionMessageResponse>(`/api/bookmark-collections/${id}`, token);
}

export function listBookmarkCollectionThreads(
  id: number | string,
  params: PaginationParams = {},
  token: string
): Promise<BookmarkCollectionThreadsResponse> {
  return api.get<BookmarkCollectionThreadsResponse>(
    `/api/bookmark-collections/${id}/threads${query(params)}`,
    token
  );
}

export function addThreadToBookmarkCollection(
  collectionId: number | string,
  threadId: number,
  token: string
): Promise<BookmarkCollectionMessageResponse> {
  return api.post<BookmarkCollectionMessageResponse>(
    `/api/bookmark-collections/${collectionId}/threads`,
    { thread_id: threadId },
    token
  );
}

export function removeThreadFromBookmarkCollection(
  collectionId: number | string,
  threadId: number,
  token: string
): Promise<BookmarkCollectionMessageResponse> {
  return api.delete<BookmarkCollectionMessageResponse>(
    `/api/bookmark-collections/${collectionId}/threads/${threadId}`,
    token
  );
}
