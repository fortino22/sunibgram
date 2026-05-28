import type { ReactNode } from 'react';
import Link from 'next/link';

const TOKEN_RE = /(@\[[a-zA-Z0-9_.]+\|\d+\])|(#[a-zA-Z0-9_]+)/g;
const RICH_MENTION_RE = /^@\[([a-zA-Z0-9_.]+)\|(\d+)\]$/;
const HASHTAG_RE = /^#[a-zA-Z0-9_]+$/;

export function renderRichContent(content: string): ReactNode[] {
  const out: ReactNode[] = [];
  let lastIdx = 0;
  let key = 0;
  let m: RegExpExecArray | null;
  TOKEN_RE.lastIndex = 0;

  while ((m = TOKEN_RE.exec(content)) !== null) {
    if (m.index > lastIdx) {
      out.push(<span key={key++}>{content.slice(lastIdx, m.index)}</span>);
    }
    const token = m[0];
    const mention = token.match(RICH_MENTION_RE);
    if (mention) {
      const [, username, userId] = mention;
      out.push(
        <Link
          key={key++}
          href={`/profile/${userId}`}
          onClick={(e) => e.stopPropagation()}
          style={{ color: 'var(--accent)' }}
          className="mention-link"
        >
          @{username}
        </Link>
      );
    } else if (HASHTAG_RE.test(token)) {
      out.push(
        <Link
          key={key++}
          href={`/explore?q=${encodeURIComponent(token)}`}
          onClick={(e) => e.stopPropagation()}
          style={{ color: 'var(--accent)' }}
          className="mention-link"
        >
          {token}
        </Link>
      );
    } else {
      out.push(<span key={key++}>{token}</span>);
    }
    lastIdx = m.index + token.length;
  }

  if (lastIdx < content.length) {
    out.push(<span key={key++}>{content.slice(lastIdx)}</span>);
  }

  return out;
}
