'use client';

import { useEffect, useRef, useState, type CSSProperties, type KeyboardEvent } from 'react';
import { useFoxAsk } from '@/hooks/fox/useFoxAsk';
import { FoxMarkdown } from '@/components/thread/foxMarkdown';

interface FoxAskPanelProps {
  /** Closes the fox card — passed to route links so navigating dismisses it. */
  onNavigate: () => void;
}

const SUGGESTIONS = [
  'How do I post something?',
  'Where are my bookmarks?',
  'How do I go premium?',
  'How do I make my account private?',
];

/**
 * Lightweight chat column for the Fox guide. The user asks how to use the app;
 * Fox streams an answer grounded in the LM wiki, with clickable in-app routes.
 * In-memory while the card is open; history persists server-side.
 */
export function FoxAskPanel({ onNavigate }: FoxAskPanelProps) {
  const { messages, ask, busy } = useFoxAsk();
  const [input, setInput] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  // Keep the newest message in view as it streams.
  useEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages]);

  const send = () => {
    const q = input.trim();
    if (!q || busy) return;
    ask(q);
    setInput('');
  };

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  };

  const empty = messages.length === 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', minHeight: 0 }}>
      <div ref={scrollRef} style={{ flex: 1, overflowY: 'auto', padding: '14px 14px 6px', minHeight: 0 }}>
        {empty ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12, paddingTop: 6 }}>
            <p style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5, margin: 0 }}>
              Hi! I&apos;m Fox 🦊 — ask me anything about how to use Sunibgram.
            </p>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {SUGGESTIONS.map((s) => (
                <button key={s} type="button" onClick={() => ask(s)} className="btn-ghost" style={chipStyle}>
                  {s}
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {messages.map((m) =>
              m.role === 'user' ? (
                <div key={m.id} style={{ alignSelf: 'flex-end', maxWidth: '85%', ...userBubble }}>
                  {m.text}
                </div>
              ) : (
                <div key={m.id} style={{ alignSelf: 'flex-start', maxWidth: '92%', ...foxBubble }}>
                  {m.text === '' && m.status === 'streaming' ? (
                    <ThinkingDots />
                  ) : (
                    <FoxMarkdown text={m.text} onNavigate={onNavigate} />
                  )}
                </div>
              ),
            )}
          </div>
        )}
      </div>

      <div style={{ borderTop: '1px solid var(--border)', padding: 10, display: 'flex', gap: 8, alignItems: 'flex-end' }}>
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={onKeyDown}
          rows={1}
          placeholder="Ask Fox how to use the app…"
          style={inputStyle}
        />
        <button
          type="button"
          onClick={send}
          disabled={busy || input.trim() === ''}
          style={{ ...sendStyle, opacity: busy || input.trim() === '' ? 0.5 : 1 }}
        >
          Send
        </button>
      </div>
    </div>
  );
}

const chipStyle: CSSProperties = {
  fontSize: 12,
  fontWeight: 500,
  color: 'var(--text)',
  background: 'var(--bg-hover)',
  border: '1px solid var(--border)',
  borderRadius: 9999,
  padding: '6px 12px',
  cursor: 'pointer',
};

const userBubble: CSSProperties = {
  fontSize: 13,
  lineHeight: 1.5,
  color: '#fff',
  background: 'var(--accent)',
  borderRadius: '14px 14px 4px 14px',
  padding: '8px 12px',
  wordBreak: 'break-word',
};

const foxBubble: CSSProperties = {
  fontSize: 13,
  color: 'var(--text)',
  background: 'var(--bg-secondary)',
  border: '1px solid var(--border)',
  borderRadius: '14px 14px 14px 4px',
  padding: '8px 12px',
};

/**
 * Claude-style "thinking" indicator — three dots that pulse in sequence while
 * Fox composes its answer. Uses the shared `pulse` keyframe (globals.css) with
 * staggered delays.
 */
function ThinkingDots() {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, padding: '2px 0' }} aria-label="Fox is thinking">
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          style={{
            width: 6,
            height: 6,
            borderRadius: '50%',
            background: 'var(--text-tertiary)',
            animation: 'pulse 1s ease-in-out infinite',
            animationDelay: `${i * 0.18}s`,
          }}
        />
      ))}
    </span>
  );
}

const inputStyle: CSSProperties = {
  flex: 1,
  resize: 'none',
  maxHeight: 120,
  boxSizing: 'border-box',
  fontSize: 13,
  fontFamily: 'inherit',
  color: 'var(--text)',
  background: 'var(--bg-hover)',
  border: '1px solid var(--border)',
  borderRadius: 14,
  padding: '9px 12px',
  outline: 'none',
};

const sendStyle: CSSProperties = {
  flexShrink: 0,
  fontSize: 13,
  fontWeight: 600,
  color: '#000',
  background: '#fff',
  border: '1px solid var(--border-strong)',
  borderRadius: 9999,
  padding: '9px 16px',
  cursor: 'pointer',
};
