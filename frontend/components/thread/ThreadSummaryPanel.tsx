'use client';

import { useCallback, useEffect, useState, type CSSProperties } from 'react';
import { createPortal } from 'react-dom';
import { ChevronDown } from 'lucide-react';
import Link from 'next/link';
import { Avatar } from '@/components/shared/Avatar';
import { ThreadBody } from '@/components/thread/ThreadBody';
import { getAuthorAvatarUrl } from '@/components/thread/threadUtils';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useThreadSummaryStream } from '@/hooks/thread/useThreadSummary';
import { useSummaryStore } from '@/stores/summaryStore';
import type { ThreadItem } from '@/lib/services/threads';

function toBullets(text: string): string[] {
  const lines = text
    .split('\n')
    .map((l) => l.replace(/^\s*(?:[-*•]|\d+[.)])\s*/, '').trim())
    .filter((l) => l.length > 0);
  return lines.length > 0 ? lines : [''];
}

const CARD_WIDTH = 360;

function cardPosition(): CSSProperties {
  const vw = typeof window === 'undefined' ? 1280 : window.innerWidth;
  return {
    position: 'fixed',
    right: 22,
    bottom: 90,
    width: Math.min(CARD_WIDTH, vw - 16),
  };
}

export function ThreadSummaryPanel() {
  const thread = useSummaryStore((s) => s.summaryThread);
  const closeSummary = useSummaryStore((s) => s.closeSummary);

  if (thread === null || typeof document === 'undefined') return null;

  // Keyed by thread id → fresh mount (and fresh stream state) per thread.
  return <SummaryCard key={thread.id} thread={thread} onClose={closeSummary} />;
}

interface SummaryCardProps {
  thread: ThreadItem;
  onClose: () => void;
}

function SummaryCard({ thread, onClose }: SummaryCardProps) {
  const { text, status, message } = useThreadSummaryStream(thread.id);
  const [closing, setClosing] = useState(false);
  const author = thread.author;
  const showSpinner = status === 'streaming' && text.length === 0;

  const requestClose = useCallback(() => setClosing(true), []);

  // Esc requests close (plays the collapse-into-orb animation first).
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setClosing(true);
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  return createPortal(
    <>
      {/* transparent click-catcher (no dimming — X-style popover) */}
      <div
        onClick={requestClose}
        style={{ position: 'fixed', inset: 0, zIndex: 110, background: 'transparent' }}
      />
      <div
        role="dialog"
        onClick={(e) => e.stopPropagation()}
        onAnimationEnd={(e) => { if (closing && e.target === e.currentTarget) onClose(); }}
        style={{
          ...cardPosition(),
          zIndex: 111,
          maxHeight: '70vh',
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          boxShadow: 'var(--shadow-dropdown)',
          transformOrigin: 'bottom right',
          animation: closing ? 'fox-modal-out 0.2s ease forwards' : 'fox-modal-in 0.2s ease',
        }}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '12px 14px',
            borderBottom: '1px solid var(--border)',
            position: 'sticky',
            top: 0,
            background: 'var(--bg-card)',
            zIndex: 1,
          }}
        >
          <span className="summary-fox" aria-hidden="true">
            <img src="/assets/fox.png" alt="" />
          </span>
          <span
            className={status === 'streaming' ? 'summary-shimmer' : undefined}
            style={{ fontWeight: 700, fontSize: 15, color: 'var(--text)', display: 'inline-block' }}
          >
            {status === 'streaming' ? 'Wait Fox Cooking…' : 'Summary'}
          </span>
          <button
            type="button"
            aria-label="Close"
            onClick={requestClose}
            className="btn-ghost"
            style={{
              marginLeft: 'auto',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: 30,
              width: 30,
              borderRadius: '50%',
              border: 'none',
              background: 'transparent',
              color: 'var(--text)',
              cursor: 'pointer',
            }}
          >
            <ChevronDown size={19} strokeWidth={2.2} />
          </button>
        </div>

        <div style={{ padding: 14, display: 'flex', flexDirection: 'column', gap: 14 }}>
          {/* Full thread content — on top */}
          <section>
            <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 10 }}>
              Full thread
            </div>
            <div className="summary-block" style={{ display: 'flex', gap: 10, padding: 8, margin: -8 }}>
              <Link href={`/profile/${author.id}`} style={{ flexShrink: 0 }}>
                <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
              </Link>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                  <Link href={`/profile/${author.id}`} style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', textDecoration: 'none' }}>
                    {author.name || author.username}
                  </Link>
                  {author.is_premium && <VerifiedBadge size={15} />}
                  <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{author.username}</span>
                  <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
                </div>
                <div style={{ marginTop: 6 }}>
                  <ThreadBody content={thread.content} media={thread.media} compact />
                </div>
              </div>
            </div>
          </section>

          {/* Summary — below the thread */}
          <section>
            <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 10 }}>
              Summary
            </div>
            <div
              className="summary-block-subtle"
              style={{
                borderRadius: 12,
                border: '1px solid var(--border)',
                padding: 12,
              }}
            >
              {status === 'limit' ? (
                <p style={{ fontSize: 14, color: 'var(--text-secondary)', margin: 0, lineHeight: 1.5 }}>
                  {message ?? "You've reached your daily summary limit. It resets tomorrow."}
                </p>
              ) : status === 'error' ? (
                <p style={{ fontSize: 14, color: 'var(--coral)', margin: 0 }}>
                  Couldn&apos;t generate a summary right now. Close and try again.
                </p>
              ) : showSpinner ? (
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: 'var(--text-secondary)' }}>
                  <span
                    style={{
                      width: 16,
                      height: 16,
                      borderRadius: '50%',
                      border: '2px solid var(--border-strong)',
                      borderTopColor: 'var(--text)',
                      animation: 'spin 0.7s linear infinite',
                      flexShrink: 0,
                    }}
                  />
                  <span style={{ fontSize: 14 }}>Summarizing…</span>
                </div>
              ) : (
                <ul style={{ margin: 0, paddingLeft: 20, listStyleType: 'disc', listStylePosition: 'outside' }}>
                  {toBullets(text).map((line, i, arr) => (
                    <li
                      key={i}
                      style={{ display: 'list-item', fontSize: 14, lineHeight: 1.5, color: 'var(--text)', marginBottom: i === arr.length - 1 ? 0 : 7 }}
                    >
                      {line}
                      {status === 'streaming' && i === arr.length - 1 && (
                        <span style={{ animation: 'pulse 1s ease-in-out infinite', fontWeight: 700 }}>▍</span>
                      )}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </section>
        </div>
      </div>
    </>,
    document.body,
  );
}
