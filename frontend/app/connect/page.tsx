'use client';

import { Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { ConnectList } from '@/components/connect/ConnectList';
import { useMutualSuggestions, useFacultySuggestions } from '@/hooks/useSuggestions';

type ConnectTab = 'mutual' | 'faculty';

const TABS: { id: ConnectTab; label: string }[] = [
  { id: 'mutual', label: 'Mutual' },
  { id: 'faculty', label: 'Faculty' },
];

function ConnectPageInner() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const activeTab: ConnectTab = searchParams.get('tab') === 'faculty' ? 'faculty' : 'mutual';

  const mutualQuery = useMutualSuggestions(activeTab === 'mutual');
  const facultyQuery = useFacultySuggestions(activeTab === 'faculty');

  function setTab(tab: ConnectTab) {
    const url = new URLSearchParams(searchParams.toString());
    url.set('tab', tab);
    router.replace(`/connect?${url.toString()}`);
  }

  return (
    <div>
      {/* Header */}
      <header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 20,
          background: 'var(--bg-blur)',
          backdropFilter: 'blur(12px)',
          padding: '12px 16px',
          display: 'flex',
          alignItems: 'center',
          gap: 14,
        }}
      >
        <button
          type="button"
          aria-label="Go back"
          onClick={() => router.back()}
          className="btn-ghost"
          style={{
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text)',
            padding: 6,
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          <ArrowLeft size={20} strokeWidth={1.8} />
        </button>
        <div style={{ fontWeight: 800, fontSize: 19, color: 'var(--text)', letterSpacing: '-0.02em' }}>
          Follow
        </div>
      </header>

      {/* Tabs */}
      <div
        role="tablist"
        style={{
          display: 'flex',
          borderBottom: '1px solid var(--border)',
          position: 'sticky',
          top: 65,
          zIndex: 10,
          background: 'var(--bg-blur)',
          backdropFilter: 'blur(12px)',
        }}
      >
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={activeTab === tab.id}
            onClick={() => setTab(tab.id)}
            className="btn-ghost"
            style={{
              flex: 1,
              padding: '16px 8px',
              border: 'none',
              cursor: 'pointer',
              fontSize: 15,
              fontWeight: activeTab === tab.id ? 700 : 400,
              color: activeTab === tab.id ? 'var(--text)' : 'var(--text-tertiary)',
              borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
              fontFamily: 'inherit',
              letterSpacing: '-0.01em',
              whiteSpace: 'nowrap',
              transition: 'color 0.15s ease, border-color 0.15s ease',
              marginBottom: -1,
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      {activeTab === 'mutual' ? (
        <ConnectList
          query={mutualQuery}
          emptyTitle="No mutual suggestions"
          emptySubtitle="People followed by accounts you follow will appear here."
        />
      ) : (
        <ConnectList
          query={facultyQuery}
          emptyTitle="No faculty suggestions"
          emptySubtitle="People from your faculty will appear here."
        />
      )}
    </div>
  );
}

export default function ConnectPage() {
  return (
    <Suspense>
      <ConnectPageInner />
    </Suspense>
  );
}
