'use client';

import { useSummaryStore } from '@/stores/summaryStore';

/**
 * Floating fox orb pinned bottom-right (X Grok-style). Decorative trigger only.
 * Collapses away while the summary modal is open (the modal grows out of it),
 * then pops back when the modal closes. Styling in globals.css `.floating-fox`.
 */
export function FloatingFoxButton() {
  const open = useSummaryStore((s) => s.summaryThread !== null);

  return (
    <button
      type="button"
      className={`floating-fox${open ? ' floating-fox-hidden' : ''}`}
      aria-label="Ask fox"
      aria-hidden={open}
    >
      <img src="/assets/fox.png" alt="" />
    </button>
  );
}
