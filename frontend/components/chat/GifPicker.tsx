'use client';

import { useState, useRef, useEffect, useMemo } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { Search, X } from 'lucide-react';

const API_KEY = process.env.NEXT_PUBLIC_KLIPY_API_KEY ?? '';
const PER_PAGE = 24;
const KLIPY_BASE = 'https://api.klipy.com/api/v1';
const MEDIA_URL_RE = /^https?:\/\/[^\s"']+\.(gif|webp|mp4)(\?[^\s"']*)?(#[^\s"']*)?$/i;

interface KlipyGif {
  slug?: string;
  title?: string;
  // Klipy nests media under `files` keyed by format/tier; exact sub-shape varies
  // by tier. We walk the object recursively to find a usable media URL.
  [key: string]: unknown;
}

interface KlipyPage {
  result: boolean;
  data?: {
    data?: KlipyGif[];
    current_page?: number;
    has_next?: boolean;
  };
  errors?: unknown;
}

// Recursively walks an item and returns the first string value that looks like
// a .gif / .webp / .mp4 URL. Survives any field-name shape Klipy returns.
function findMediaUrl(node: unknown, depth = 0): string | null {
  if (depth > 6 || node == null) return null;
  if (typeof node === 'string') {
    return MEDIA_URL_RE.test(node) ? node : null;
  }
  if (Array.isArray(node)) {
    for (const v of node) {
      const r = findMediaUrl(v, depth + 1);
      if (r) return r;
    }
    return null;
  }
  if (typeof node === 'object') {
    for (const v of Object.values(node as Record<string, unknown>)) {
      const r = findMediaUrl(v, depth + 1);
      if (r) return r;
    }
  }
  return null;
}

interface GifPickerProps {
  onSelect: (gifUrl: string) => void;
  onClose: () => void;
  placement?: 'above' | 'inline';
}

export function GifPicker({ onSelect, onClose, placement = 'above' }: GifPickerProps) {
  const [query, setQuery] = useState('');
  const [activeQuery, setActiveQuery] = useState('');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const shapeWarnedRef = useRef(false);

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isError,
    error,
    isLoading,
  } = useInfiniteQuery<KlipyPage>({
    queryKey: ['klipy', activeQuery || 'trending'],
    queryFn: async ({ pageParam }) => {
      if (!API_KEY) throw new Error('Missing NEXT_PUBLIC_KLIPY_API_KEY');
      const page = typeof pageParam === 'number' ? pageParam : 1;
      const endpoint = activeQuery ? 'search' : 'trending';
      const params = new URLSearchParams({ per_page: String(PER_PAGE), page: String(page) });
      if (activeQuery) params.set('q', activeQuery);
      const url = `${KLIPY_BASE}/${API_KEY}/gifs/${endpoint}?${params.toString()}`;
      const res = await fetch(url);
      if (!res.ok) throw new Error(`Klipy request failed: ${res.status}`);
      const json = (await res.json()) as KlipyPage;
      // Klipy returns 200 OK with `result: false` for rate limits, bad keys, etc.
      // Throw so the query enters an error state and pagination stops — otherwise
      // an "empty success" would let the IntersectionObserver spam the endpoint.
      if (json.result === false) {
        const msg =
          (json.errors &&
            typeof json.errors === 'object' &&
            'message' in json.errors &&
            Array.isArray((json.errors as { message: unknown }).message)
            ? ((json.errors as { message: string[] }).message[0] ?? 'Klipy error')
            : 'Klipy error');
        throw new Error(msg);
      }
      return json;
    },
    initialPageParam: 1,
    getNextPageParam: (last) => {
      if (!last || last.result === false) return undefined;
      return last.data?.has_next ? (last.data?.current_page ?? 0) + 1 : undefined;
    },
    enabled: API_KEY.length > 0,
    refetchOnWindowFocus: false,
    retry: false,
  });

  const tiles = useMemo(() => {
    const items: KlipyGif[] = (data?.pages ?? []).flatMap((p) => p?.data?.data ?? []);
    const out: { item: KlipyGif; url: string; key: string }[] = [];
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      const url = findMediaUrl(item);
      if (url) out.push({ item, url, key: `${item.slug ?? 'k'}-${i}` });
    }
    // If the API returned items but we couldn't extract any URLs, log the first
    // item once so the developer can adjust MEDIA_URL_RE / findMediaUrl.
    if (items.length > 0 && out.length === 0 && !shapeWarnedRef.current) {
      shapeWarnedRef.current = true;
      console.warn('[GifPicker] Klipy returned items but no media URL was found. Sample:', items[0]);
    }
    return out;
  }, [data]);

  // Infinite-scroll trigger — gated on having at least one rendered tile, so an
  // empty/erroring response can never trigger a fetch loop.
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || tiles.length === 0 || !hasNextPage || isFetchingNextPage || isError) return;
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          fetchNextPage();
        }
      },
      { rootMargin: '200px' },
    );
    io.observe(el);
    return () => io.disconnect();
  }, [fetchNextPage, hasNextPage, isFetchingNextPage, isError, tiles.length]);

  function handleQueryChange(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setQuery(val);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => setActiveQuery(val), 350);
  }

  function handleGifClick(url: string, e: React.MouseEvent<HTMLButtonElement>) {
    e.preventDefault();
    onSelect(url);
    onClose();
  }

  return (
    <div
      style={{
        ...(placement === 'above'
          ? { position: 'absolute', bottom: 'calc(100% + 8px)', left: 0, right: 0 }
          : { position: 'relative', width: '100%' }),
        maxHeight: 'min(380px, 50vh)',
        maxWidth: 'calc(100vw - 32px)',
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 12,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        boxShadow: '0 4px 24px rgba(0,0,0,0.18)',
        zIndex: 50,
      }}
    >
      {/* Search bar */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 10px', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
        <Search size={14} style={{ color: 'var(--text-muted)' }} />
        <input
          autoFocus
          value={query}
          onChange={handleQueryChange}
          placeholder="Search GIFs…"
          style={{ flex: 1, minWidth: 0, border: 'none', background: 'transparent', color: 'var(--text)', fontSize: 13, outline: 'none', fontFamily: 'inherit' }}
        />
        <button className="hover-dim" onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2, display: 'flex', color: 'var(--text-muted)' }}>
          <X size={16} />
        </button>
      </div>

      {/* No API key warning */}
      {!API_KEY && (
        <div style={{ padding: 20, textAlign: 'center', color: 'var(--text-muted)', fontSize: 12 }}>
          Set <code>NEXT_PUBLIC_KLIPY_API_KEY</code> to enable GIFs.
        </div>
      )}

      {/* GIF grid */}
      {API_KEY && (
        <div style={{ flex: 1, minHeight: 0, overflowY: 'auto', overflowX: 'hidden', padding: 6 }}>
          {isError && (
            <div style={{ padding: 16, textAlign: 'center', color: 'var(--text-muted)', fontSize: 12 }}>
              {error instanceof Error ? error.message : 'Couldn’t load GIFs.'}
            </div>
          )}

          {!isError && tiles.length > 0 && (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 4 }}>
              {tiles.map(({ item, url, key }) => (
                <button
                  key={key}
                  type="button"
                  onClick={(e) => handleGifClick(url, e)}
                  style={{
                    padding: 0,
                    border: 'none',
                    background: 'var(--bg-input, rgba(255,255,255,0.04))',
                    borderRadius: 6,
                    overflow: 'hidden',
                    cursor: 'pointer',
                    display: 'block',
                    lineHeight: 0,
                  }}
                  title={item.title ?? ''}
                >
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={url}
                    alt={item.title ?? 'GIF'}
                    loading="lazy"
                    style={{ width: '100%', height: 'auto', display: 'block' }}
                  />
                </button>
              ))}
            </div>
          )}

          {!isError && !isLoading && tiles.length === 0 && (
            <div style={{ padding: 16, textAlign: 'center', color: 'var(--text-muted)', fontSize: 12 }}>
              No GIFs found.
            </div>
          )}

          {(isLoading || isFetchingNextPage) && (
            <div style={{ padding: 12, textAlign: 'center', color: 'var(--text-muted)', fontSize: 11 }}>
              Loading…
            </div>
          )}

          <div ref={sentinelRef} style={{ height: 1 }} />
        </div>
      )}

      {/* KLIPY attribution */}
      <div style={{ textAlign: 'right', padding: '3px 8px', fontSize: 10, color: 'var(--text-muted)', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        Powered by KLIPY
      </div>
    </div>
  );
}
