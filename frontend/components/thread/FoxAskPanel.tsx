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
              Hi! I&apos;m Fox — ask me anything about how to use Sunibgram.
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
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {messages.map((m) =>
              m.role === 'user' ? (
                <div key={m.id} style={{ alignSelf: 'flex-end', maxWidth: '88%', ...userBubble }}>
                  {m.text}
                </div>
              ) : (
                m.text === '' && m.status === 'streaming' ? (
                  <div key={m.id} style={{ alignSelf: 'flex-start', padding: '2px 2px' }}>
                    <ThinkingIndicator />
                  </div>
                ) : (
                  <div
                    key={m.id}
                    className={m.status === 'streaming' ? 'fox-stream-wave' : undefined}
                    style={{ alignSelf: 'flex-start', maxWidth: '92%', ...foxBubble }}
                  >
                    <FoxMarkdown text={m.text} onNavigate={onNavigate} />
                  </div>
                )
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
          className="fox-send"
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
  fontSize: 13.5,
  lineHeight: 1.5,
  color: '#fff',
  background: 'var(--accent)',
  borderRadius: '16px 16px 5px 16px',
  padding: '9px 13px',
  wordBreak: 'break-word',
};

const foxBubble: CSSProperties = {
  fontSize: 13.5,
  color: 'var(--text)',
  background: 'var(--bg-secondary)',
  border: '1px solid var(--border)',
  borderRadius: '16px 16px 16px 5px',
  padding: '9px 13px',
};

const THINKING_PHRASES = [
  'Fox is thinking...',
  'Fox is cooking...',
  'Fox is sniffing around the guide...',
  'Fox is fetching the answer...',
];

/**
 * "Thinking" indicator — a shimmering status phrase that rotates every ~2s,
 * beside three sequentially-pulsing dots. Uses the Fox-Agent blue palette
 * (`.fox-thinking-shimmer` / `.fox-thinking-dot` from globals.css) to match the
 * sidebar "Fox Agent" entry, with a soft cross-fade between phrases.
 */
function ThinkingIndicator() {
  const [i, setI] = useState(0);
  useEffect(() => {
    const t = setInterval(() => setI((n) => (n + 1) % THINKING_PHRASES.length), 2000);
    return () => clearInterval(t);
  }, []);
  const phrase = THINKING_PHRASES[i];
  return (
    <span
      style={{ display: 'inline-flex', alignItems: 'center', gap: 8, padding: '1px 0' }}
      aria-label={phrase}
    >
      <span
        key={i}
        className="fox-thinking-wrap"
        style={{ fontSize: 13, fontWeight: 500, lineHeight: 1, display: 'inline-flex', whiteSpace: 'pre' }}
        aria-hidden="true"
      >
        {Array.from(phrase).map((ch, idx) => (
          <span
            key={idx}
            className="fox-wave-letter"
            style={{ animationDelay: `${idx * 0.06}s` }}
          >
            {ch}
          </span>
        ))}
      </span>
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

