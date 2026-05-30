'use client';

/**
 * Floating fox orb pinned bottom-right (X Grok-style).
 * UI only for now — no click action wired yet. Styling in globals.css `.floating-fox`.
 */
export function FloatingFoxButton() {
  return (
    <button type="button" className="floating-fox" aria-label="Ask fox">
      <img src="/assets/fox.png" alt="" />
    </button>
  );
}
