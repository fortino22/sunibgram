'use client';

import { useEffect, useRef, useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { renderAsync } from 'docx-preview';

interface EventDocViewerProps {
  urls: string[];
}

type DocStatus = 'idle' | 'loading' | 'ok' | 'error';

function DocViewer({ url }: { url: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [status, setStatus] = useState<DocStatus>('idle');
  const [message, setMessage] = useState('');

  useEffect(() => {
    let cancelled = false;
    const container = containerRef.current;
    if (!container) return;

    setStatus('loading');
    container.innerHTML = '';

    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`Failed to fetch document (${r.status})`);
        return r.blob();
      })
      .then((blob) => {
        if (cancelled) return;
        return renderAsync(blob, container, undefined, {
          className: 'docx',
          inWrapper: true,
          ignoreWidth: false,
          ignoreHeight: false,
          breakPages: true,
          experimental: true,
          useBase64URL: true,
        });
      })
      .then(() => {
        if (!cancelled) setStatus('ok');
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setMessage(err instanceof Error ? err.message : 'Unknown error');
        setStatus('error');
      });

    return () => {
      cancelled = true;
    };
  }, [url]);

  return (
    <div
      style={{
        position: 'relative',
        width: '100%',
        height: '70vh',
        border: '1px solid var(--border)',
        borderRadius: 12,
        overflow: 'auto',
        background: 'var(--bg-subtle)',
        boxSizing: 'border-box',
      }}
    >
      {(status === 'loading' || status === 'idle') && (
        <div style={overlayStyle('var(--text-tertiary)')}>Loading document…</div>
      )}
      {status === 'error' && (
        <div style={overlayStyle('var(--coral, #e53e3e)')}>
          Could not render document: {message}
        </div>
      )}
      <div
        ref={containerRef}
        style={{ display: status === 'ok' ? 'block' : 'none', padding: '16px 0' }}
      />
    </div>
  );
}

function overlayStyle(color: string): React.CSSProperties {
  return {
    position: 'absolute',
    inset: 0,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color,
    fontSize: 14,
    textAlign: 'center',
    padding: 24,
  };
}

export function EventDocViewer({ urls }: EventDocViewerProps) {
  const [idx, setIdx] = useState(0);

  if (urls.length === 0) {
    return (
      <div
        style={{
          padding: 24,
          textAlign: 'center',
          color: 'var(--text-tertiary)',
          border: '1px dashed var(--border)',
          borderRadius: 12,
        }}
      >
        No documents attached.
      </div>
    );
  }

  const url = urls[idx];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
        <span style={{ fontSize: 13, color: 'var(--text-secondary)', fontWeight: 500 }}>
          Document {idx + 1} of {urls.length}
        </span>
        <div style={{ display: 'flex', gap: 6 }}>
          <button
            type="button"
            onClick={() => setIdx((i) => Math.max(0, i - 1))}
            disabled={idx === 0}
            style={{
              width: 36, height: 36, borderRadius: 9999,
              border: '1px solid var(--border)', background: 'transparent',
              color: 'var(--text)', display: 'flex', alignItems: 'center',
              justifyContent: 'center', cursor: idx === 0 ? 'not-allowed' : 'pointer',
              opacity: idx === 0 ? 0.4 : 1,
            }}
            aria-label="Previous document"
          >
            <ChevronLeft size={18} />
          </button>
          <button
            type="button"
            onClick={() => setIdx((i) => Math.min(urls.length - 1, i + 1))}
            disabled={idx === urls.length - 1}
            style={{
              width: 36, height: 36, borderRadius: 9999,
              border: '1px solid var(--border)', background: 'transparent',
              color: 'var(--text)', display: 'flex', alignItems: 'center',
              justifyContent: 'center',
              cursor: idx === urls.length - 1 ? 'not-allowed' : 'pointer',
              opacity: idx === urls.length - 1 ? 0.4 : 1,
            }}
            aria-label="Next document"
          >
            <ChevronRight size={18} />
          </button>
          <a
            href={url}
            download
            rel="noopener noreferrer"
            style={{
              padding: '0 14px', height: 36, borderRadius: 9999,
              border: '1px solid var(--border)', background: 'transparent',
              color: 'var(--text)', fontSize: 13, fontWeight: 500,
              display: 'inline-flex', alignItems: 'center', textDecoration: 'none',
            }}
          >
            Download
          </a>
        </div>
      </div>

      <DocViewer key={url} url={url} />
    </div>
  );
}
