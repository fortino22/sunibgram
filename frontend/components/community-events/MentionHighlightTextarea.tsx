'use client';

import { useRef } from 'react';

interface Props {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  /** Returns true if a bare @word should render as a blue mention tag. */
  recognize: (word: string) => boolean;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  rows?: number;
  onKeyDown?: React.KeyboardEventHandler<HTMLTextAreaElement>;
  onSelect?: React.ReactEventHandler<HTMLTextAreaElement>;
  /** Reports the mounted textarea element (use for portal anchors — avoids reading a ref in render). */
  onTextareaEl?: (el: HTMLTextAreaElement | null) => void;
}

// Text-layout props shared by the backdrop and the textarea so the highlighted
// text behind the (transparent-text) textarea lines up exactly, char for char.
// Border/background/radius are owned by the parent .ce-composer card, not here.
const textLayout: React.CSSProperties = {
  margin: 0,
  padding: 12,
  fontSize: 14,
  fontFamily: 'inherit',
  lineHeight: 1.5,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
  boxSizing: 'border-box',
};

export function MentionHighlightTextarea({ value, onChange, placeholder, recognize, textareaRef, rows = 3, onKeyDown, onSelect, onTextareaEl }: Props) {
  const backdropRef = useRef<HTMLDivElement>(null);

  function buildNodes() {
    const nodes: React.ReactNode[] = [];
    let last = 0;
    let key = 0;
    for (const m of value.matchAll(/@([a-zA-Z0-9_.]+)/g)) {
      const token = m[0];
      const word = m[1];
      const idx = m.index ?? 0;
      if (!recognize(word)) continue;
      if (idx > last) nodes.push(<span key={key++}>{value.slice(last, idx)}</span>);
      // Highlight must keep the EXACT glyph width of the transparent textarea so
      // the caret stays aligned — only color/background (layout-neutral). Never
      // font-weight/padding/margin (those widen the token and drift the caret).
      nodes.push(
        <span
          key={key++}
          style={{
            color: 'var(--accent)',
            background: 'color-mix(in oklab, var(--accent) 12%, transparent)',
            borderRadius: 4,
          }}
        >
          {token}
        </span>
      );
      last = idx + token.length;
    }
    if (last < value.length) nodes.push(<span key={key++}>{value.slice(last)}</span>);
    // A trailing newline has no glyph; add a zero-width char so the backdrop
    // grows the same final line the textarea does.
    if (value.endsWith('\n')) nodes.push('​');
    return nodes;
  }

  function syncScroll() {
    if (backdropRef.current && textareaRef.current) {
      backdropRef.current.scrollTop = textareaRef.current.scrollTop;
      backdropRef.current.scrollLeft = textareaRef.current.scrollLeft;
    }
  }

  return (
    <div style={{ position: 'relative' }}>
      <div
        ref={backdropRef}
        aria-hidden
        style={{
          ...textLayout,
          position: 'absolute',
          inset: 0,
          overflow: 'hidden',
          color: 'var(--text)',
          pointerEvents: 'none',
        }}
      >
        {buildNodes()}
      </div>
      <textarea
        ref={(el) => {
          textareaRef.current = el;
          onTextareaEl?.(el);
        }}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onScroll={syncScroll}
        onKeyDown={onKeyDown}
        onSelect={onSelect}
        placeholder={placeholder}
        rows={rows}
        style={{
          ...textLayout,
          position: 'relative',
          display: 'block',
          width: '100%',
          border: 'none',
          background: 'transparent',
          color: 'transparent',
          caretColor: 'var(--text)',
          resize: 'vertical',
          outline: 'none',
        }}
      />
    </div>
  );
}
