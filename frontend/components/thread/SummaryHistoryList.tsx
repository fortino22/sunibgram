'use client';

import type { CSSProperties } from 'react';
import type { SummaryHistoryItem } from '@/lib/services/threads';

interface SummaryHistoryListProps {
  items: SummaryHistoryItem[];
  selectedId: number | null;
  onSelect: (threadId: number) => void;
}

/** Compact, clickable history sidebar — one row per summarized thread. */
export function SummaryHistoryList({ items, selectedId, onSelect }: SummaryHistoryListProps) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      {items.map((item) => {
        const selected = item.thread_id === selectedId;
        const title = item.content.trim().length > 0 ? item.content : 'Post no longer available';
        return (
          <button
            key={item.thread_id}
            type="button"
            onClick={() => onSelect(item.thread_id)}
            className="summary-block"
            style={{
              ...rowStyle,
              background: selected ? 'var(--bg-hover)' : 'transparent',
              borderLeft: `2px solid ${selected ? 'var(--text)' : 'transparent'}`,
            }}
          >
            <span
              style={{
                fontSize: 13,
                lineHeight: 1.4,
                color: 'var(--text)',
                display: '-webkit-box',
                WebkitBoxOrient: 'vertical',
                WebkitLineClamp: 2,
                overflow: 'hidden',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}
            >
              {title}
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{item.summarized_at}</span>
              {item.is_stale && <span style={staleDotStyle} title="Outdated" aria-label="Outdated" />}
            </span>
          </button>
        );
      })}
    </div>
  );
}

const rowStyle: CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 5,
  alignItems: 'flex-start',
  textAlign: 'left',
  width: '100%',
  padding: '9px 10px',
  border: 'none',
  borderRadius: 8,
  cursor: 'pointer',
};

const staleDotStyle: CSSProperties = {
  width: 7,
  height: 7,
  borderRadius: '50%',
  background: 'var(--coral, #f4212e)',
  flexShrink: 0,
};
