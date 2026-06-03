'use client';

import { Fragment, type ReactNode } from 'react';
import Link from 'next/link';

// Matches markdown links [label](target). Only internal app routes (target
// starting with "/") render as clickable Next links; everything else (external
// http links, mailto, etc.) renders as plain label text for safety.
const LINK_RE = /\[([^\]]+)\]\(([^)]+)\)/g;
// Bold (**x** / __x__) and italic (*x* / _x_) and inline code (`x`).
const EMPHASIS_RE = /(\*\*|__)(.+?)\1|(\*|_)(.+?)\3|`([^`]+)`/g;

/** Parse bold/italic/code in a plain (link-free) text segment → ReactNodes. */
function renderEmphasis(text: string, keyBase: string): ReactNode[] {
  const out: ReactNode[] = [];
  let last = 0;
  let k = 0;
  for (const m of text.matchAll(EMPHASIS_RE)) {
    const start = m.index ?? 0;
    if (start > last) out.push(<Fragment key={`${keyBase}-t${k++}`}>{text.slice(last, start)}</Fragment>);
    if (m[2] !== undefined) {
      out.push(<strong key={`${keyBase}-b${k++}`}>{m[2]}</strong>);
    } else if (m[4] !== undefined) {
      out.push(<em key={`${keyBase}-i${k++}`}>{m[4]}</em>);
    } else if (m[5] !== undefined) {
      // Inline code → plain text (no monospace box, keep it clean).
      out.push(<Fragment key={`${keyBase}-c${k++}`}>{m[5]}</Fragment>);
    }
    last = start + m[0].length;
  }
  if (last < text.length) out.push(<Fragment key={`${keyBase}-t${k++}`}>{text.slice(last)}</Fragment>);
  return out;
}

/** Render one line: links first, then bold/italic/code on the non-link parts. */
function renderInline(line: string, onNavigate?: () => void): ReactNode[] {
  const out: ReactNode[] = [];
  let last = 0;
  let key = 0;
  for (const m of line.matchAll(LINK_RE)) {
    const [full, label, target] = m;
    const start = m.index ?? 0;
    if (start > last) out.push(...renderEmphasis(line.slice(last, start), `pre${key++}`));
    if (target.startsWith('/')) {
      out.push(
        <Link
          key={`lnk${key++}`}
          href={target}
          onClick={onNavigate}
          style={{ color: 'var(--blue, #1d9bf0)', textDecoration: 'none', fontWeight: 600 }}
        >
          {label}
        </Link>,
      );
    } else {
      // External/unsafe target — render the label as plain emphasized text.
      out.push(...renderEmphasis(label, `lbl${key++}`));
    }
    last = start + full.length;
  }
  if (last < line.length) out.push(...renderEmphasis(line.slice(last), `post${key++}`));
  return out;
}

interface FoxMarkdownProps {
  text: string;
  /** Called when an internal route link is clicked (e.g. close the fox card). */
  onNavigate?: () => void;
}

interface Block {
  kind: 'text' | 'bullet' | 'number' | 'heading' | 'gap';
  content: string;
  /** For numbered lists, the rendered marker (e.g. "1."). */
  marker?: string;
}

/** Parse raw text into display blocks, stripping markdown noise. */
function toBlocks(text: string): Block[] {
  const rawLines = text.split('\n');
  const blocks: Block[] = [];
  for (const raw of rawLines) {
    const line = raw.replace(/\s+$/, '');
    if (line.trim() === '') {
      // Collapse runs of blank lines into a single gap.
      if (blocks.length > 0 && blocks[blocks.length - 1].kind !== 'gap') {
        blocks.push({ kind: 'gap', content: '' });
      }
      continue;
    }
    // Heading: leading # ... → bold line.
    const heading = line.match(/^\s{0,3}#{1,6}\s+(.*)$/);
    if (heading) {
      blocks.push({ kind: 'heading', content: heading[1] });
      continue;
    }
    // Numbered list: "1." or "2)".
    const numbered = line.match(/^\s*(\d+)[.)]\s+(.*)$/);
    if (numbered) {
      blocks.push({ kind: 'number', content: numbered[2], marker: `${numbered[1]}.` });
      continue;
    }
    // Bullet: -, *, or • (but not **bold** at line start).
    const bullet = line.match(/^\s*([-*•])\s+(.*)$/);
    if (bullet && !line.startsWith('**')) {
      blocks.push({ kind: 'bullet', content: bullet[2] });
      continue;
    }
    blocks.push({ kind: 'text', content: line });
  }
  // Drop trailing gap.
  while (blocks.length > 0 && blocks[blocks.length - 1].kind === 'gap') blocks.pop();
  return blocks;
}

/**
 * Minimal markdown renderer for Fox guide answers. Renders bold/italic, headings
 * (as bold lines), bullets and numbered lists, and clickable in-app links —
 * stripping raw markdown symbols so users never see literal `*`, `#`, or
 * backticks. No external markdown dependency.
 */
export function FoxMarkdown({ text, onNavigate }: FoxMarkdownProps) {
  const blocks = toBlocks(text);
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      {blocks.map((b, i) => {
        if (b.kind === 'gap') return <div key={i} style={{ height: 6 }} />;
        if (b.kind === 'heading') {
          return (
            <div key={i} style={{ fontWeight: 700, lineHeight: 1.5, marginTop: i > 0 ? 2 : 0 }}>
              {renderInline(b.content, onNavigate)}
            </div>
          );
        }
        if (b.kind === 'bullet' || b.kind === 'number') {
          return (
            <div key={i} style={{ display: 'flex', gap: 7, lineHeight: 1.5 }}>
              <span style={{ color: 'var(--text-tertiary)', flexShrink: 0 }}>
                {b.kind === 'number' ? b.marker : '•'}
              </span>
              <span style={{ wordBreak: 'break-word' }}>{renderInline(b.content, onNavigate)}</span>
            </div>
          );
        }
        return (
          <div key={i} style={{ lineHeight: 1.5, wordBreak: 'break-word' }}>
            {renderInline(b.content, onNavigate)}
          </div>
        );
      })}
    </div>
  );
}
