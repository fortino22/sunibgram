'use client';

import type { CSSProperties } from 'react';
import Link from 'next/link';
import { toBullets } from '@/components/thread/summaryUtils';
import { renderRichContent } from '@/components/thread/MentionText';
import type { SummaryHistoryItem } from '@/lib/services/threads';

interface SummaryDetailProps {
  item: SummaryHistoryItem | null;
  /** Called when navigating to the thread, so the card can close. */
  onNavigate: () => void;
}

/** Right-pane detail: full thread content on top, summary result below. */
export function SummaryDetail({ item, onNavigate }: SummaryDetailProps) {
  if (item === null) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', padding: 24 }}>
        <p style={{ fontSize: 13, color: 'var(--text-secondary)', margin: 0, textAlign: 'center' }}>
          Select a summary to view it.
        </p>
      </div>
    );
  }

  const bullets = toBullets(item.summary);
  const hasContent = item.content.trim().length > 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      {/* Full thread content — on top */}
      <section>
        <div style={labelStyle}>Full thread</div>
        <div
          style={{
            fontSize: 14,
            lineHeight: 1.6,
            color: 'var(--text)',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          {hasContent ? renderRichContent(item.content) : 'Post no longer available'}
        </div>
        {item.media.length > 0 && (
          <div
            style={{
              display: 'grid',
              gap: 6,
              marginTop: hasContent ? 10 : 8,
              gridTemplateColumns:
                item.media.length === 1 ? '1fr' : item.media.length === 4 ? 'repeat(2, 1fr)' : 'repeat(3, 1fr)',
            }}
          >
            {item.media.map((m) => (
              <div
                key={m.media_id}
                style={{
                  borderRadius: 10,
                  overflow: 'hidden',
                  background: 'var(--bg-subtle)',
                  height: item.media.length === 1 ? 160 : item.media.length === 4 ? 110 : 80,
                }}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={m.url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
              </div>
            ))}
          </div>
        )}
      </section>

      <div style={{ height: 1, background: 'var(--border)' }} />

      {/* Summary result */}
      <section>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
          <span style={labelStyle}>Summary</span>
          {item.is_stale && <span style={staleBadgeStyle}>Outdated</span>}
        </div>
        <ul style={{ margin: 0, paddingLeft: 20, listStyleType: 'disc', listStylePosition: 'outside' }}>
          {bullets.map((line, i) => (
            <li
              key={i}
              style={{ fontSize: 14, lineHeight: 1.5, color: 'var(--text)', marginBottom: i === bullets.length - 1 ? 0 : 7 }}
            >
              {line}
            </li>
          ))}
        </ul>
      </section>

      {/* Footer */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 2 }}>
        <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{item.summarized_at}</span>
        <Link
          href={`/thread/${item.thread_id}`}
          onClick={onNavigate}
          style={{ marginLeft: 'auto', fontSize: 12, fontWeight: 600, color: 'var(--blue, #1d9bf0)', textDecoration: 'none' }}
        >
          Open thread
        </Link>
      </div>
    </div>
  );
}

const labelStyle: CSSProperties = {
  fontSize: 11,
  fontWeight: 700,
  color: 'var(--text-tertiary)',
  textTransform: 'uppercase',
  letterSpacing: '0.04em',
  marginBottom: 10,
};

const staleBadgeStyle: CSSProperties = {
  fontSize: 10,
  fontWeight: 700,
  color: 'var(--coral, #f4212e)',
  border: '1px solid var(--coral, #f4212e)',
  borderRadius: 9999,
  padding: '1px 7px',
};
