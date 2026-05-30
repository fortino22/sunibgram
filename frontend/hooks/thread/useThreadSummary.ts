'use client';

import { useEffect, useRef, useState } from 'react';
import { streamThreadSummary } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export type SummaryStatus = 'streaming' | 'done' | 'error';

export interface ThreadSummaryResult {
  text: string;
  status: SummaryStatus;
}

const REVEAL_STEP = 3; // chars revealed per tick
const REVEAL_MS = 18;

/**
 * Streams a thread's LLM summary and reveals it with a typewriter effect.
 * The reveal ticker is independent of how the backend chunks the response, so
 * the text types out smoothly whether it arrives token-by-token (fresh) or as
 * one blob (cached). Mount via a component keyed by thread id (fresh state).
 */
export function useThreadSummaryStream(id: number | string): ThreadSummaryResult {
  const token = useAuthStore((state) => state.token);
  const [displayed, setDisplayed] = useState('');
  const [status, setStatus] = useState<SummaryStatus>(token === null ? 'error' : 'streaming');

  const targetRef = useRef('');
  const streamDoneRef = useRef(false);

  // Stream the raw text into targetRef.
  useEffect(() => {
    if (token === null) return undefined;
    const controller = new AbortController();

    streamThreadSummary(id, token, {
      signal: controller.signal,
      onChunk: (chunk) => { targetRef.current += chunk; },
    })
      .then(() => { streamDoneRef.current = true; })
      .catch((err: unknown) => {
        if (controller.signal.aborted) return;
        if (err instanceof DOMException && err.name === 'AbortError') return;
        setStatus('error');
      });

    return () => controller.abort();
  }, [id, token]);

  // Reveal ticker — advance displayed toward the received text.
  useEffect(() => {
    if (token === null) return undefined;
    const timer = setInterval(() => {
      setDisplayed((prev) => {
        const target = targetRef.current;
        if (prev.length < target.length) {
          return target.slice(0, Math.min(prev.length + REVEAL_STEP, target.length));
        }
        if (streamDoneRef.current) setStatus((s) => (s === 'streaming' ? 'done' : s));
        return prev;
      });
    }, REVEAL_MS);
    return () => clearInterval(timer);
  }, [id, token]);

  return { text: displayed, status };
}
