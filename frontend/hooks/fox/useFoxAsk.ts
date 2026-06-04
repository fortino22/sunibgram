'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { streamFoxAsk, FoxLimitError, type FoxQAPair, type FoxSessionTurn } from '@/lib/services/fox';
import { useAuthStore } from '@/stores/authStore';

export type FoxRole = 'user' | 'fox';
export type FoxMsgStatus = 'streaming' | 'done' | 'error' | 'limit';

export interface FoxMessage {
  id: number;
  role: FoxRole;
  /** Revealed text (typewriter); may trail the received text while streaming. */
  text: string;
  status?: FoxMsgStatus;
}

const REVEAL_STEP = 3; // chars revealed per tick
const REVEAL_MS = 16;

/**
 * Drives the Fox guide chat for one re-openable session. Holds the message list
 * and exposes `ask(question)` (streams the answer with a typewriter reveal),
 * `loadSession(id, turns)` (hydrate a past conversation to continue it), and
 * `startNew()` (begin a fresh session). The active session id is sent to the
 * backend so turns persist server-side; a brand-new session's id comes back on
 * the first answer.
 */
export function useFoxAsk() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();
  const [messages, setMessages] = useState<FoxMessage[]>([]);
  const [busy, setBusy] = useState(false);

  // Per-stream reveal target + the id of the fox message being filled.
  const targetRef = useRef('');
  const activeIdRef = useRef<number | null>(null);
  const streamDoneRef = useRef(false);
  const nextId = useRef(1);
  const abortRef = useRef<AbortController | null>(null);
  // Current chat session id (0 = not yet created server-side).
  const sessionIdRef = useRef(0);
  // Mirror of messages so `ask` can read the latest list without a stale closure.
  const messagesRef = useRef<FoxMessage[]>([]);
  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  // Reveal ticker: advance the active fox message toward its received text.
  useEffect(() => {
    const timer = setInterval(() => {
      const activeId = activeIdRef.current;
      if (activeId === null) return;
      setMessages((prev) => {
        let changed = false;
        const next = prev.map((m) => {
          if (m.id !== activeId) return m;
          const target = targetRef.current;
          if (m.text.length < target.length) {
            changed = true;
            return { ...m, text: target.slice(0, Math.min(m.text.length + REVEAL_STEP, target.length)) };
          }
          if (streamDoneRef.current && m.status === 'streaming') {
            changed = true;
            return { ...m, status: 'done' as FoxMsgStatus };
          }
          return m;
        });
        return changed ? next : prev;
      });
    }, REVEAL_MS);
    return () => clearInterval(timer);
  }, []);

  // startNew clears the panel for a fresh conversation.
  const startNew = useCallback(() => {
    abortRef.current?.abort();
    abortRef.current = null;
    activeIdRef.current = null;
    targetRef.current = '';
    streamDoneRef.current = false;
    sessionIdRef.current = 0;
    setBusy(false);
    setMessages([]);
  }, []);

  // loadSession hydrates a past chat session's turns so the user can continue it.
  const loadSession = useCallback((sessionId: number, turns: FoxSessionTurn[]) => {
    abortRef.current?.abort();
    abortRef.current = null;
    activeIdRef.current = null;
    targetRef.current = '';
    streamDoneRef.current = false;
    sessionIdRef.current = sessionId;
    setBusy(false);

    const hydrated: FoxMessage[] = [];
    for (const t of turns) {
      hydrated.push({ id: nextId.current++, role: 'user', text: t.question });
      hydrated.push({ id: nextId.current++, role: 'fox', text: t.answer, status: 'done' });
    }
    setMessages(hydrated);
  }, []);

  const ask = useCallback(
    (question: string) => {
      const q = question.trim();
      if (!q || busy || token === null) return;

      // Snapshot prior completed turns (user→fox pairs) from the live list.
      const prev = messagesRef.current;
      const history: FoxQAPair[] = [];
      for (let i = 0; i < prev.length - 1; i += 1) {
        if (prev[i].role === 'user' && prev[i + 1]?.role === 'fox') {
          history.push({ question: prev[i].text, answer: prev[i + 1].text });
        }
      }

      const userId = nextId.current++;
      const foxId = nextId.current++;

      // Reset stream state, then append the two messages with a PURE updater
      // (no side effects inside — StrictMode double-invokes updaters).
      targetRef.current = '';
      streamDoneRef.current = false;
      activeIdRef.current = foxId;
      setBusy(true);
      setMessages((cur) => [
        ...cur,
        { id: userId, role: 'user', text: q },
        { id: foxId, role: 'fox', text: '', status: 'streaming' },
      ]);

      const controller = new AbortController();
      abortRef.current = controller;

      streamFoxAsk(q, sessionIdRef.current, history, token, {
        signal: controller.signal,
        onChunk: (chunk) => {
          targetRef.current += chunk;
        },
      })
        .then((resolvedSessionId) => {
          streamDoneRef.current = true;
          if (resolvedSessionId !== null) {
            sessionIdRef.current = resolvedSessionId;
          }
          // Refresh the history list so the new/updated session appears, and the
          // turns detail for this session so re-opening it shows the new turn.
          void queryClient.invalidateQueries({ queryKey: ['fox', 'sessions'] });
          if (sessionIdRef.current !== 0) {
            void queryClient.invalidateQueries({ queryKey: ['fox', 'session', sessionIdRef.current] });
          }
        })
        .catch((err: unknown) => {
          if (controller.signal.aborted) return;
          if (err instanceof FoxLimitError) {
            targetRef.current = err.message;
            streamDoneRef.current = true;
            setMessages((cur) => cur.map((m) => (m.id === foxId ? { ...m, status: 'limit' } : m)));
            return;
          }
          targetRef.current = 'Something went wrong reaching Fox. Please try again.';
          streamDoneRef.current = true;
          setMessages((cur) => cur.map((m) => (m.id === foxId ? { ...m, status: 'error' } : m)));
        })
        .finally(() => {
          setBusy(false);
        });
    },
    [busy, token, queryClient],
  );

  useEffect(() => () => abortRef.current?.abort(), []);

  return { messages, ask, busy, loadSession, startNew };
}
