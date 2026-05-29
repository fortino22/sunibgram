'use client';

import { useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PdfViewerProps {
  urls: string[];
}

export function PdfViewer({ urls }: PdfViewerProps) {
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
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 12,
        }}
      >
        <span style={{ fontSize: 13, color: 'var(--text-secondary)', fontWeight: 500 }}>
          Document {idx + 1} of {urls.length}
        </span>
        <div style={{ display: 'flex', gap: 6 }}>
          <button
            type="button"
            onClick={() => setIdx((i) => Math.max(0, i - 1))}
            disabled={idx === 0}
            className="btn-ghost"
            style={{
              width: 36,
              height: 36,
              borderRadius: 9999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: idx === 0 ? 'not-allowed' : 'pointer',
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
            className="btn-ghost"
            style={{
              width: 36,
              height: 36,
              borderRadius: 9999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              display: 'flex',
              alignItems: 'center',
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
            target="_blank"
            rel="noopener noreferrer"
            className="btn-ghost"
            style={{
              padding: '0 14px',
              height: 36,
              borderRadius: 9999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              fontSize: 13,
              fontWeight: 500,
              display: 'inline-flex',
              alignItems: 'center',
              textDecoration: 'none',
            }}
          >
            Open in new tab
          </a>
        </div>
      </div>
      <div
        style={{
          width: '100%',
          height: '70vh',
          border: '1px solid var(--border)',
          borderRadius: 12,
          overflow: 'hidden',
          background: 'var(--bg-card)',
        }}
      >
        <iframe
          key={url}
          src={url}
          title={`Document ${idx + 1}`}
          style={{ width: '100%', height: '100%', border: 'none' }}
        />
      </div>
    </div>
  );
}
