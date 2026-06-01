'use client';

import { Suspense, useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { Search } from 'lucide-react';
import { useSummaryHistoryInfinite } from '@/hooks/thread/useSummaryHistoryInfinite';
import { SummaryHistoryList } from '@/components/thread/SummaryHistoryList';
import { SummaryDetail } from '@/components/thread/SummaryDetail';
import { FoxAgentTeaser } from '@/components/thread/FoxAgentTeaser';

type Tab = 'summaries' | 'agent';

/**
 * Full summary-history page reached via the sidebar fox tab and the fox card's
 * "Show more" link. Two-pane (list + detail) like the fox card, but with
 * server-side search and infinite scroll over the user's entire history.
 */
export default function SummariesPage() {
  return (
    <Suspense>
      <SummariesPageInner />
    </Suspense>
  );
}

function SummariesPageInner() {
  // "Show more" deep-links to ?tab=summaries; otherwise default to Your Agent.
  const initialTab: Tab = useSearchParams().get('tab') === 'summaries' ? 'summaries' : 'agent';
  const [activeTab, setActiveTab] = useState<Tab>(initialTab);
  const [input, setInput] = useState('');
  const [q, setQ] = useState('');

  // Debounce the search box so each keystroke doesn't fire a request.
  useEffect(() => {
    const t = setTimeout(() => setQ(input.trim()), 300);
    return () => clearTimeout(t);
  }, [input]);

  const query = useSummaryHistoryInfinite(q);
  const items = query.data?.pages.flatMap((p) => p.data) ?? [];
  const [selectedId, setSelectedId] = useState<number | null>(null);

  // Derive selection at render: clicked item if still present, else top match.
  const selected = items.find((i) => i.thread_id === selectedId) ?? items[0] ?? null;
  const effectiveSelectedId = selected?.thread_id ?? null;

  const { hasNextPage, isFetchingNextPage, fetchNextPage } = query;
  const sentinelRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && hasNextPage && !isFetchingNextPage) {
          void fetchNextPage();
        }
      },
      { rootMargin: '300px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const searching = q.length > 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      {/* Sticky header — title + tabs + (Summaries) search */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'var(--bg-blur)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '18px 20px 0',
        }}
      >
        <div
          className="fox-blue-title"
          style={{ display: 'inline-block', fontWeight: 700, fontSize: 20, letterSpacing: '-0.02em', marginBottom: 12 }}
        >
          Fox Agent
        </div>
        <div role="tablist" style={{ display: 'flex', gap: 0 }}>
          <TabButton label="Your Agent" active={activeTab === 'agent'} onClick={() => setActiveTab('agent')} />
          <TabButton label="Summaries" active={activeTab === 'summaries'} onClick={() => setActiveTab('summaries')} />
        </div>
      </div>

      {activeTab === 'agent' ? (
        <div style={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
          <FoxAgentTeaser />
        </div>
      ) : (
        /* Two-pane body — list (left) + detail (right) */
        <div style={{ display: 'flex', flex: 1, minHeight: 0 }}>
          <div style={{ width: 300, flexShrink: 0, borderRight: '1px solid var(--border)', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            <div style={{ position: 'relative', padding: '16px 10px 12px' }}>
              <Search
                size={16}
                style={{ position: 'absolute', left: 22, top: 'calc(50% + 2px)', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }}
              />
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Search summaries"
                style={{
                  width: '100%',
                  boxSizing: 'border-box',
                  padding: '9px 12px 9px 36px',
                  fontSize: 13,
                  color: 'var(--text)',
                  background: 'var(--bg-hover)',
                  border: '1px solid var(--border)',
                  borderRadius: 9999,
                  outline: 'none',
                  fontFamily: 'inherit',
                }}
              />
            </div>
            <div style={{ flex: 1, overflowY: 'auto', padding: '0 8px 10px', minHeight: 0 }}>
            {query.isLoading ? (
              <ListSkeleton />
            ) : query.isError ? (
              <ListMessage text="Couldn't load your summaries." />
            ) : items.length === 0 ? (
              <ListMessage
                text={searching ? 'No matches.' : 'No summaries yet — tap the fox on any thread to summarize it.'}
              />
            ) : (
              <>
                <SummaryHistoryList items={items} selectedId={effectiveSelectedId} onSelect={setSelectedId} />
                <div ref={sentinelRef} style={{ height: 1 }} />
                {isFetchingNextPage && (
                  <div style={{ textAlign: 'center', padding: '10px 0', fontSize: 12, color: 'var(--text-tertiary)' }}>
                    Loading…
                  </div>
                )}
              </>
            )}
            </div>
          </div>

          <div style={{ flex: 1, minWidth: 0, overflowY: 'auto', padding: 20 }}>
            <SummaryDetail item={selected} onNavigate={() => undefined} />
          </div>
        </div>
      )}
    </div>
  );
}

interface TabButtonProps {
  label: string;
  active: boolean;
  onClick: () => void;
}

function TabButton({ label, active, onClick }: TabButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      style={{
        flex: 1,
        padding: '14px 0',
        background: 'transparent',
        border: 'none',
        cursor: 'pointer',
        color: active ? 'var(--text)' : 'var(--text-tertiary)',
        fontFamily: 'inherit',
        fontWeight: active ? 700 : 500,
        fontSize: 14,
        position: 'relative',
      }}
    >
      {label}
      {active && (
        <span
          aria-hidden
          style={{
            position: 'absolute',
            bottom: 0,
            left: '50%',
            transform: 'translateX(-50%)',
            width: 48,
            height: 3,
            borderRadius: 2,
            background: 'var(--accent)',
          }}
        />
      )}
    </button>
  );
}

function ListSkeleton() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, padding: '2px' }}>
      {[0, 1, 2, 3, 4, 5].map((i) => (
        <div
          key={i}
          style={{ height: 52, borderRadius: 8, background: 'var(--bg-hover)', animation: 'pulse 1.2s ease-in-out infinite' }}
        />
      ))}
    </div>
  );
}

function ListMessage({ text }: { text: string }) {
  return (
    <p style={{ fontSize: 12, color: 'var(--text-secondary)', lineHeight: 1.5, margin: 0, padding: '12px 8px' }}>{text}</p>
  );
}
