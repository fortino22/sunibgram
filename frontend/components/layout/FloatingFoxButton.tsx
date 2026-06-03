'use client';

import { useCallback, useEffect, useState, type CSSProperties } from 'react';
import { ChevronDown } from 'lucide-react';
import { useSummaryStore } from '@/stores/summaryStore';
import { FoxHistoryPanel } from '@/components/thread/FoxHistoryPanel';
import { FoxAskPanel } from '@/components/thread/FoxAskPanel';

type FoxTab = 'summaries' | 'ask';

/**
 * Floating fox orb pinned bottom-right (X Grok-style). Clicking it pops a small
 * card that grows out of the orb — same open/close choreography as the thread
 * summary modal: the orb collapses away while the card is open, the card closes
 * via a down-chevron (or Esc / outside click) playing the collapse-back
 * animation, then the orb pops back. The card shows recent summaries (list +
 * detail); the "Your Agent" teaser now lives on the /summaries page.
 * Styling in globals.css `.floating-fox` + keyframes `fox-modal-in/out`.
 */
export function FloatingFoxButton() {
  const summaryOpen = useSummaryStore((s) => s.summaryThread !== null);
  const [showSoon, setShowSoon] = useState(false);
  const [closing, setClosing] = useState(false);
  const [tab, setTab] = useState<FoxTab>('ask');

  const requestClose = useCallback(() => setClosing(true), []);
  const orbHidden = summaryOpen || showSoon;

  // Esc requests close (plays the collapse-into-orb animation first).
  useEffect(() => {
    if (!showSoon) return undefined;
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setClosing(true);
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [showSoon]);

  return (
    <>
      {showSoon && (
        <>
          {/* transparent click-catcher — outside click requests close */}
          <div
            onClick={requestClose}
            style={{ position: 'fixed', inset: 0, zIndex: 110, background: 'transparent' }}
          />
          <div
            role="dialog"
            aria-label="Fox"
            className="fox-soon-card"
            onClick={(e) => e.stopPropagation()}
            onAnimationEnd={(e) => {
              if (closing && e.target === e.currentTarget) {
                setShowSoon(false);
                setClosing(false);
              }
            }}
            style={{
              position: 'fixed',
              right: 22,
              bottom: 90,
              zIndex: 111,
              display: 'flex',
              flexDirection: 'column',
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 18,
              width: 500,
              maxWidth: 'calc(100vw - 16px)',
              height: 'min(78vh, 600px)',
              overflow: 'hidden',
              transformOrigin: 'bottom right',
              boxShadow: 'var(--shadow-dropdown)',
              animation: closing ? 'fox-modal-out 0.2s ease forwards' : 'fox-modal-in 0.2s ease',
            }}
          >
            <button
              type="button"
              aria-label="Close"
              onClick={requestClose}
              className="btn-ghost"
              style={{
                position: 'absolute',
                top: 8,
                right: 8,
                zIndex: 3,
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                height: 28,
                width: 28,
                borderRadius: '50%',
                border: 'none',
                background: 'transparent',
                color: 'var(--text)',
                cursor: 'pointer',
              }}
            >
              <ChevronDown size={18} strokeWidth={2.2} />
            </button>

            {/* Tab header — Summaries (history) and Ask (guide chat). */}
            <div role="tablist" style={{ display: 'flex', gap: 0, paddingRight: 40, borderBottom: '1px solid var(--border)' }}>
              <FoxTabButton label="Ask Fox" active={tab === 'ask'} onClick={() => setTab('ask')} />
              <FoxTabButton label="History" active={tab === 'summaries'} onClick={() => setTab('summaries')} />
            </div>

            {/* Body swaps by tab. minHeight:0 lets inner panels scroll. */}
            <div style={{ flex: 1, minHeight: 0 }}>
              {tab === 'summaries' ? (
                <FoxHistoryPanel active={showSoon && !closing} onNavigate={requestClose} />
              ) : (
                <FoxAskPanel onNavigate={requestClose} />
              )}
            </div>
          </div>
        </>
      )}

      <button
        type="button"
        className={`floating-fox${orbHidden ? ' floating-fox-hidden' : ''}`}
        aria-label="Ask fox"
        aria-hidden={orbHidden}
        onClick={() => setShowSoon(true)}
      >
        <img src="/assets/fox.png" alt="" />
      </button>
    </>
  );
}

function FoxTabButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button type="button" role="tab" aria-selected={active} onClick={onClick} style={tabButtonStyle(active)}>
      {label}
      {active && (
        <span
          aria-hidden
          style={{
            position: 'absolute',
            bottom: 0,
            left: '50%',
            transform: 'translateX(-50%)',
            width: 40,
            height: 3,
            borderRadius: 2,
            background: 'var(--accent)',
          }}
        />
      )}
    </button>
  );
}

function tabButtonStyle(active: boolean): CSSProperties {
  return {
    flex: 1,
    padding: '12px 0',
    background: 'transparent',
    border: 'none',
    cursor: 'pointer',
    color: active ? 'var(--text)' : 'var(--text-tertiary)',
    fontFamily: 'inherit',
    fontWeight: active ? 700 : 500,
    fontSize: 13,
    position: 'relative',
  };
}
