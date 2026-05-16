'use client';

import { useState, useCallback, useRef } from 'react';
import { Grid } from '@giphy/react-components';
import { GiphyFetch } from '@giphy/js-fetch-api';
import type { IGif } from '@giphy/js-types';
import { Search, X } from 'lucide-react';

const API_KEY = process.env.NEXT_PUBLIC_GIPHY_API_KEY ?? '';
const gf = API_KEY ? new GiphyFetch(API_KEY) : null;

interface GifPickerProps {
  onSelect: (gifUrl: string) => void;
  onClose: () => void;
}

export function GifPicker({ onSelect, onClose }: GifPickerProps) {
  const [query, setQuery] = useState('');
  const [activeQuery, setActiveQuery] = useState('');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const fetchGifs = useCallback(
    (offset: number) => {
      if (!gf) return Promise.reject(new Error('No API key'));
      return activeQuery
        ? gf.search(activeQuery, { offset, limit: 24 })
        : gf.trending({ offset, limit: 24 });
    },
    [activeQuery],
  );

  const handleQueryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setQuery(val);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => setActiveQuery(val), 350);
  };

  const handleGifClick = (gif: IGif, e: React.SyntheticEvent<HTMLElement>) => {
    e.preventDefault();
    onSelect(gif.images.original.url);
    onClose();
  };

  return (
    <div
      style={{
        position: 'absolute',
        bottom: 'calc(100% + 8px)',
        left: 0,
        right: 0,
        maxHeight: 380,
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
        <Search size={14} color="var(--text-muted)" />
        <input
          autoFocus
          value={query}
          onChange={handleQueryChange}
          placeholder="Search GIFs…"
          style={{ flex: 1, border: 'none', background: 'transparent', color: 'var(--text)', fontSize: 13, outline: 'none', fontFamily: 'inherit' }}
        />
        <button onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2, display: 'flex', color: 'var(--text-muted)' }}>
          <X size={16} />
        </button>
      </div>

      {/* No API key warning */}
      {!API_KEY && (
        <div style={{ padding: 20, textAlign: 'center', color: 'var(--text-muted)', fontSize: 12 }}>
          Set <code>NEXT_PUBLIC_GIPHY_API_KEY</code> to enable GIFs.
        </div>
      )}

      {/* GIF grid */}
      {gf && (
        <div style={{ flex: 1, overflowY: 'auto', padding: 6 }}>
          <Grid
            key={activeQuery}
            columns={3}
            width={330}
            fetchGifs={fetchGifs}
            onGifClick={handleGifClick}
            noLink
            hideAttribution
          />
        </div>
      )}

      {/* GIPHY attribution — required by their API terms */}
      <div style={{ textAlign: 'right', padding: '3px 8px', fontSize: 10, color: 'var(--text-muted)', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        Powered by GIPHY
      </div>
    </div>
  );
}
