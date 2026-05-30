'use client';

import { useCallback, useEffect, useState } from 'react';
import { ChevronDown } from 'lucide-react';
import { useSummaryStore } from '@/stores/summaryStore';
import { FoxSleepSprite } from './FoxSleepSprite';

// Raining Hermes-face drops behind the fox — varied position/size/timing.
const HERMES_DROPS = [
  { left: '6%', size: 30, dur: '3.0s', delay: '0s' },
  { left: '20%', size: 42, dur: '4.2s', delay: '1.1s' },
  { left: '36%', size: 26, dur: '2.7s', delay: '0.5s' },
  { left: '52%', size: 38, dur: '3.6s', delay: '1.8s' },
  { left: '66%', size: 28, dur: '3.1s', delay: '0.2s' },
  { left: '80%', size: 46, dur: '4.4s', delay: '2.3s' },
  { left: '90%', size: 32, dur: '2.9s', delay: '0.9s' },
];

/**
 * Floating fox orb pinned bottom-right (X Grok-style). Clicking it pops a small
 * "coming soon" card (sleeping fox) that grows out of the orb — same open/close
 * choreography as the thread summary modal: the orb collapses away while the
 * card is open, the card closes via a down-chevron (or Esc / outside click)
 * playing the collapse-back animation, then the orb pops back.
 * Styling in globals.css `.floating-fox` + keyframes `fox-modal-in/out`.
 */
export function FloatingFoxButton() {
  const summaryOpen = useSummaryStore((s) => s.summaryThread !== null);
  const [showSoon, setShowSoon] = useState(false);
  const [closing, setClosing] = useState(false);

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
            aria-label="Coming soon"
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
              alignItems: 'center',
              gap: 11,
              padding: '38px 32px 28px',
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 18,
              width: 280,
              overflow: 'hidden',
              transformOrigin: 'bottom right',
              animation: closing ? 'fox-modal-out 0.2s ease forwards' : 'fox-modal-in 0.2s ease',
            }}
          >
            {/* Raining, glitching Hermes faces behind the fox + text. */}
            <div className="hermes-rain" aria-hidden="true">
              {HERMES_DROPS.map((d, i) => (
                <span
                  key={i}
                  className="hermes-drop"
                  style={{ left: d.left, width: d.size, height: d.size, animationDuration: d.dur, animationDelay: d.delay }}
                >
                  <span className="hermes-drop-img" />
                </span>
              ))}
            </div>

            <button
              type="button"
              aria-label="Close"
              onClick={requestClose}
              className="btn-ghost"
              style={{
                position: 'absolute',
                top: 6,
                right: 6,
                zIndex: 2,
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

            <div style={{ position: 'relative', zIndex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 10 }}>
              <FoxSleepSprite />
              <span style={{ fontWeight: 700, fontSize: 17, color: 'var(--text)', textAlign: 'center' }}>
                Your personal agentic AI
              </span>
              <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Coming soon</span>
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
