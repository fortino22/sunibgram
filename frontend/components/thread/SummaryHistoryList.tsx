'use client';

import type { CSSProperties } from 'react';
import type { SummaryHistoryItem } from '@/lib/services/threads';

interface SummaryHistoryListProps {
  items: SummaryHistoryItem[];
  selectedId: number | null;
  onSelect: (id: number) => void;
}

/** Compact, clickable history sidebar — one row per summary or chat ask. */
export function SummaryHistoryList({ items, selectedId, onSelect }: SummaryHistoryListProps) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      {items.map((item) => {
        const selected = item.id === selectedId;
        const isChat = item.type === 'chat';
        const title = isChat
          ? (item.question?.trim() || 'Ask')
          : item.content.trim().length > 0
            ? item.content
            : 'Post no longer available';
        return (
          <button
            key={`${item.type}-${item.id}`}
            type="button"
            onClick={() => onSelect(item.id)}
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
              <span style={isChat ? askTagStyle : sumTagStyle}>{isChat ? 'Ask' : 'Summary'}</span>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{item.summarized_at}</span>
              {item.is_stale && <span style={staleDotStyle} title="Outdated" aria-label="Outdated" />}
            </span>
          </button>
        );
      })}
    </div>
  );
}

const tagBase: CSSProperties = {
  fontSize: 9,
  fontWeight: 700,
  textTransform: 'uppercase',
  letterSpacing: '0.04em',
  borderRadius: 9999,
  padding: '1px 6px',
};
const askTagStyle: CSSProperties = { ...tagBase, color: 'var(--accent)', border: '1px solid var(--accent)' };
const sumTagStyle: CSSProperties = { ...tagBase, color: 'var(--text-tertiary)', border: '1px solid var(--border-strong)' };

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
