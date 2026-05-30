'use client';

/**
 * Decorative fox brand icon shown beside the thread 3-dots menu.
 * Pure presentational, non-interactive (aria-hidden). Hover/blue styling
 * lives in globals.css under `.thread-fox`.
 */
export function ThreadFoxButton() {
  return (
    <span className="thread-fox" aria-hidden="true">
      <img src="/assets/fox.png" alt="" />
    </span>
  );
}
