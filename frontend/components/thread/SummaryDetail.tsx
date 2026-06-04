'use client';

import type { CSSProperties } from 'react';
import Link from 'next/link';
import { toBullets } from '@/components/thread/summaryUtils';
import { renderRichContent } from '@/components/thread/MentionText';
import { FoxMarkdown } from '@/components/thread/foxMarkdown';
import { ArrowRight } from 'lucide-react';
import { useFoxSession } from '@/hooks/fox/useFoxSession';
import type { FoxSessionItem } from '@/lib/services/fox';

interface SummaryDetailProps {
  item: FoxSessionItem | null;
  /** Called when navigating to the thread, so the card can close. */
  onNavigate: () => void;
  /**
   * Re-open a chat session in the Ask Fox tab to continue it. When omitted (e.g.
   * the /summaries page, which has no Ask tab) the chat view stays read-only.
   */
  onContinue?: (sessionId: number) => void;
}

/** Right-pane detail: chat transcript, or full thread content + summary result. */
export function SummaryDetail({ item, onNavigate, onContinue }: SummaryDetailProps) {
  // Chat sessions load their turns; the hook is disabled for summaries/empty.
  const chatSessionId = item && item.type === 'chat' ? item.id : null;
  const { data: detail, isLoading: turnsLoading } = useFoxSession(chatSessionId);

  if (item === null) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', padding: 24 }}>
        <p style={{ fontSize: 13, color: 'var(--text-secondary)', margin: 0, textAlign: 'center' }}>
          Select an item to view it.
        </p>
      </div>
    );
  }

  // Chat session: replay the conversation as chat bubbles (read-only), mirroring
  // the Ask Fox panel — user questions right, Fox answers left.
  if (item.type === 'chat') {
    const turns = detail?.turns ?? [];
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {turnsLoading && turns.length === 0 ? (
          <p style={{ fontSize: 13, color: 'var(--text-secondary)', margin: 0 }}>Loading conversation…</p>
        ) : turns.length === 0 ? (
          <p style={{ fontSize: 13, color: 'var(--text-secondary)', margin: 0 }}>No messages in this conversation.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            {turns.map((turn, i) => (
              <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <div style={{ alignSelf: 'flex-end', maxWidth: '88%', ...userBubble }}>
                  {turn.question.trim() || '—'}
                </div>
                <div style={{ alignSelf: 'flex-start', maxWidth: '92%', ...foxBubble }}>
                  <FoxMarkdown text={turn.answer} onNavigate={onNavigate} />
                </div>
              </div>
            ))}
          </div>
        )}
        <div style={{ fontSize: 11, color: 'var(--text-tertiary)', textAlign: 'center' }}>{item.updated_at}</div>
        {onContinue && (
          <button type="button" onClick={() => onContinue(item.id)} className="fox-continue" style={{ alignSelf: 'center' }}>
            Continue in Ask Fox
            <ArrowRight size={13} strokeWidth={2.4} />
          </button>
        )}
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
        <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{item.updated_at}</span>
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

// Chat-bubble styles mirror FoxAskPanel so the History transcript reads like the
// live chat: user questions right (accent), Fox answers left (secondary).
const userBubble: CSSProperties = {
  fontSize: 13.5,
  lineHeight: 1.5,
  color: '#fff',
  background: 'var(--accent)',
  borderRadius: '16px 16px 5px 16px',
  padding: '9px 13px',
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
};

const foxBubble: CSSProperties = {
  fontSize: 13.5,
  color: 'var(--text)',
  background: 'var(--bg-secondary)',
  border: '1px solid var(--border)',
  borderRadius: '16px 16px 16px 5px',
  padding: '9px 13px',
  wordBreak: 'break-word',
};

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
