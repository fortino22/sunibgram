'use client';

import { FoxSleepSprite } from '@/components/layout/FoxSleepSprite';

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
 * "Your Agent" teaser: sleeping fox over a rain of glitching Hermes faces, with
 * the "coming soon" copy. Previously the fox card's footer; now fills the
 * Your-Agent tab on the /summaries page. Uses the global `.hermes-rain` CSS.
 */
export function FoxAgentTeaser() {
  return (
    <div
      style={{
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
        width: '100%',
        height: '100%',
        minHeight: 440,
        padding: '48px 24px',
        overflow: 'hidden',
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
      <div style={{ position: 'relative', zIndex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 10 }}>
        <FoxSleepSprite width={264} height={212} />
        <span style={{ fontWeight: 700, fontSize: 28, color: 'var(--text)', textAlign: 'center', letterSpacing: '-0.01em' }}>
          Your personal agentic AI
        </span>
        <span style={{ fontSize: 18, color: 'var(--text-secondary)' }}>Coming soon</span>
      </div>
    </div>
  );
}
