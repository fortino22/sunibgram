import type { InfiniteData } from '@tanstack/react-query';
import type { ThreadListResponse } from '@/lib/services/threads';

// Thread lists live in the query cache in two shapes: a flat `ThreadListResponse`
// (`{ data, meta }`) for single-page lists, and `InfiniteData<ThreadListResponse>`
// (`{ pages: [...] }`) for the infinite feed/following/reply queries. Because
// `setQueriesData` matches by key prefix, a single updater can receive either
// shape, so cache mutators must detect which one they got before touching it.

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export function isThreadListResponse(value: unknown): value is ThreadListResponse {
  return (
    isRecord(value) &&
    Array.isArray(value.data) &&
    isRecord(value.meta) &&
    typeof value.meta.total === 'number'
  );
}

export function isInfiniteThreadListData(value: unknown): value is InfiniteData<ThreadListResponse> {
  return isRecord(value) && Array.isArray(value.pages) && (value.pages as unknown[]).every(isThreadListResponse);
}

function removeThreadFromList(current: ThreadListResponse, threadId: number): ThreadListResponse {
  const data = current.data.filter((item) => item.id !== threadId);
  if (data.length === current.data.length) {
    return current;
  }
  return {
    ...current,
    data,
    meta: { ...current.meta, total: Math.max(0, current.meta.total - 1) },
  };
}

/**
 * Remove a thread from whatever thread-list shape lives at a cache key. Returns
 * unexpected shapes (including `undefined`) unchanged, so it is safe to pass to
 * `setQueriesData` for prefix-matched keys that mix flat and infinite queries.
 */
export function removeThreadFromCache(current: unknown, threadId: number): unknown {
  if (isInfiniteThreadListData(current)) {
    return {
      ...current,
      pages: current.pages.map((page) => removeThreadFromList(page, threadId)),
    };
  }
  if (isThreadListResponse(current)) {
    return removeThreadFromList(current, threadId);
  }
  return current;
}
