'use client';

import type { MouseEvent } from 'react';
import type { ThreadItem } from '@/lib/services/threads';
import { useSummaryStore } from '@/stores/summaryStore';

interface ThreadFoxButtonProps {
  thread: ThreadItem;
}

/**
 * Fox icon beside the thread 3-dots. Click opens the right-side summary panel
 * (X Grok-style, summarize-only). Styling in globals.css `.thread-fox`.
 */
export function ThreadFoxButton({ thread }: ThreadFoxButtonProps) {
  const openSummary = useSummaryStore((s) => s.openSummary);

  function handleClick(e: MouseEvent<HTMLButtonElement>) {
    e.preventDefault();
    e.stopPropagation();
    const r = e.currentTarget.getBoundingClientRect();
    openSummary(thread, { top: r.top, left: r.left, bottom: r.bottom, right: r.right, width: r.width });
  }

  return (
    <button type="button" className="thread-fox" aria-label="Summarize thread" onClick={handleClick}>
      <img src="/assets/fox.png" alt="" />
    </button>
  );
}
