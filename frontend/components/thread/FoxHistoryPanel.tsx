'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useFoxSessions } from '@/hooks/fox/useFoxSessions';
import { SummaryHistoryList } from '@/components/thread/SummaryHistoryList';
import { SummaryDetail } from '@/components/thread/SummaryDetail';

interface FoxHistoryPanelProps {
  /** The card is open — drives the query's `enabled` flag. */
  active: boolean;
  /** Closes the fox card (used when navigating to a thread). */
  onNavigate: () => void;
  /** Re-open a chat session: load its turns into Ask Fox and switch tabs. */
  onContinueSession: (sessionId: number) => void;
}

/**
 * Two-pane history body for the fox card: a clickable session list on the left,
 * and the selected session's content on the right. Clicking an "Ask" session
 * re-opens it in the Ask Fox tab to continue; a "Summary" session shows its
 * summary in the detail pane (read-only).
 */
export function FoxHistoryPanel({ active, onNavigate, onContinueSession }: FoxHistoryPanelProps) {
  const router = useRouter();
  const { data, isLoading, isError, refetch } = useFoxSessions(active);
  const items = data?.data ?? [];
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [query, setQuery] = useState('');

  // "Show more" → full history page, History tab; close the card as we navigate.
  const handleShowMore = () => {
    router.push('/summaries?tab=history');
    onNavigate();
  };

  // Both kinds preview in the detail pane on click; a chat session is only
  // continued via the explicit "Continue in Ask Fox" link in that pane.
  const handleSelect = (id: number) => setSelectedId(id);

  // Client-side search over title + content + summary + preview (history is capped).
  const q = query.trim().toLowerCase();
  const filtered = q
    ? items.filter((i) => `${i.title} ${i.content} ${i.summary} ${i.preview}`.toLowerCase().includes(q))
    : items;

  // No auto-pick — the pane stays a placeholder until a row is clicked.
  const selected = selectedId !== null ? filtered.find((i) => i.id === selectedId) ?? null : null;
  const effectiveSelectedId = selected?.id ?? null;

  return (
    <div style={{ display: 'flex', flexDirection: 'row', height: '100%', minHeight: 0 }}>
      {/* Left sidebar — history list */}
      <div
        style={{
          width: 200,
          flexShrink: 0,
          borderRight: '1px solid var(--border)',
          display: 'flex',
          flexDirection: 'column',
          minHeight: 0,
        }}
      >
        <div style={sidebarHeaderStyle}>Recent activity</div>
        {items.length > 0 && (
          <div style={{ padding: '0 10px 14px' }}>
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search activity"
              style={searchInputStyle}
            />
          </div>
        )}
        <div style={{ flex: 1, overflowY: 'auto', padding: '0 8px 10px' }}>
          {isLoading ? (
            <SidebarSkeleton />
          ) : isError ? (
            <SidebarError onRetry={() => refetch()} />
          ) : items.length === 0 ? (
            <p style={emptyStyle}>Nothing yet — ask Fox a question or tap the fox on any thread to summarize it.</p>
          ) : filtered.length === 0 ? (
            <p style={emptyStyle}>No matches.</p>
          ) : (
            <SummaryHistoryList items={filtered} selectedId={effectiveSelectedId} onSelect={handleSelect} />
          )}
        </div>
        {items.length >= MODAL_HISTORY_LIMIT && (
          <button type="button" onClick={handleShowMore} className="btn-ghost" style={showMoreStyle}>
            Show more
          </button>
        )}
      </div>

      {/* Right column — selected summary detail */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, minHeight: 0 }}>
        <div style={{ flex: 1, overflowY: 'auto', padding: 14 }}>
          <SummaryDetail item={selected} onNavigate={onNavigate} onContinue={onContinueSession} />
        </div>
      </div>
    </div>
  );
}

const sidebarHeaderStyle = {
  fontSize: 11,
  fontWeight: 700,
  color: 'var(--text-tertiary)',
  textTransform: 'uppercase',
  letterSpacing: '0.04em',
  padding: '10px 12px 8px',
} as const;

// Modal fetches the 6 most recent; a full page means more likely exist, so
// only then is "Show more" worth showing. Keep in sync with useFoxSessions.
const MODAL_HISTORY_LIMIT = 6;

// No inline `background` — that would beat the `.btn-ghost:hover` rule on
// specificity and kill the hover tint.
const showMoreStyle = {
  flexShrink: 0,
  margin: '4px 10px 10px',
  padding: '8px 0',
  fontSize: 12,
  fontWeight: 600,
  color: 'var(--text)',
  border: '1px solid var(--border-strong)',
  borderRadius: 9999,
  cursor: 'pointer',
} as const;

const searchInputStyle = {
  width: '100%',
  boxSizing: 'border-box',
  fontSize: 13,
  color: 'var(--text)',
  background: 'var(--bg-hover)',
  border: '1px solid var(--border)',
  borderRadius: 8,
  padding: '6px 10px',
  outline: 'none',
} as const;

const emptyStyle = {
  fontSize: 12,
  color: 'var(--text-secondary)',
  lineHeight: 1.5,
  margin: 0,
  padding: '8px 6px',
} as const;

function SidebarSkeleton() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, padding: '2px 2px' }}>
      {[0, 1, 2, 3].map((i) => (
        <div
          key={i}
          style={{
            height: 46,
            borderRadius: 8,
            background: 'var(--bg-hover)',
            animation: 'pulse 1.2s ease-in-out infinite',
          }}
        />
      ))}
    </div>
  );
}

function SidebarError({ onRetry }: { onRetry: () => void }) {
  return (
    <div style={{ textAlign: 'center', padding: '16px 8px' }}>
      <p style={{ fontSize: 12, color: 'var(--text-secondary)', margin: '0 0 10px' }}>Couldn&apos;t load your history.</p>
      <button
        type="button"
        className="btn-ghost"
        onClick={onRetry}
        style={{
          fontSize: 12,
          fontWeight: 600,
          color: 'var(--text)',
          background: 'transparent',
          border: '1px solid var(--border-strong)',
          borderRadius: 9999,
          padding: '5px 12px',
          cursor: 'pointer',
        }}
      >
        Try again
      </button>
    </div>
  );
}
