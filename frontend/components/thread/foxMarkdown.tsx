'use client';

import { Fragment, type ReactNode } from 'react';
import Link from 'next/link';

// Matches markdown links [label](target). Only internal app routes (target
// starting with "/") render as clickable Next links; everything else (external
// http links, mailto, etc.) renders as plain label text for safety.
const LINK_RE = /\[([^\]]+)\]\(([^)]+)\)/g;

/** Render one line of Fox answer text, turning [label](/route) into links. */
function renderInline(line: string, onNavigate?: () => void): ReactNode[] {
  const out: ReactNode[] = [];
  let last = 0;
  let key = 0;
  for (const m of line.matchAll(LINK_RE)) {
    const [full, label, target] = m;
    const start = m.index ?? 0;
    if (start > last) out.push(<Fragment key={key++}>{line.slice(last, start)}</Fragment>);
    if (target.startsWith('/')) {
      out.push(
        <Link
          key={key++}
          href={target}
          onClick={onNavigate}
          style={{ color: 'var(--blue, #1d9bf0)', textDecoration: 'none', fontWeight: 600 }}
        >
          {label}
        </Link>,
      );
    } else {
      out.push(<Fragment key={key++}>{label}</Fragment>);
    }
    last = start + full.length;
  }
  if (last < line.length) out.push(<Fragment key={key++}>{line.slice(last)}</Fragment>);
  return out;
}

interface FoxMarkdownProps {
  text: string;
  /** Called when an internal route link is clicked (e.g. close the fox card). */
  onNavigate?: () => void;
}

/**
 * Minimal markdown renderer for Fox guide answers: preserves line breaks,
 * renders `- ` lines as bullets, and turns `[label](/route)` into clickable
 * in-app links. Intentionally tiny — no external markdown dependency.
 */
export function FoxMarkdown({ text, onNavigate }: FoxMarkdownProps) {
  const lines = text.split('\n');
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      {lines.map((raw, i) => {
        const line = raw.replace(/\s+$/, '');
        const bullet = /^\s*[-*]\s+/.test(line);
        const content = bullet ? line.replace(/^\s*[-*]\s+/, '') : line;
        if (line.trim() === '') return <div key={i} style={{ height: 6 }} />;
        return (
          <div key={i} style={{ display: 'flex', gap: bullet ? 6 : 0, lineHeight: 1.5 }}>
            {bullet && <span style={{ color: 'var(--text-tertiary)' }}>•</span>}
            <span style={{ wordBreak: 'break-word' }}>{renderInline(content, onNavigate)}</span>
          </div>
        );
      })}
    </div>
  );
}
