'use client';

import { useEffect } from 'react';
import { createPortal } from 'react-dom';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';

interface ImageZoomLightboxProps {
  images: string[];
  index: number;
  title?: string;
  onClose: () => void;
  onNavigate: (index: number) => void;
}

export function ImageZoomLightbox({ images, index, title, onClose, onNavigate }: ImageZoomLightboxProps) {
  const current = images[index];
  const hasPrev = index > 0;
  const hasNext = index < images.length - 1;
  const multi = images.length > 1;

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowLeft' && hasPrev) onNavigate(index - 1);
      if (e.key === 'ArrowRight' && hasNext) onNavigate(index + 1);
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [index, hasPrev, hasNext, onClose, onNavigate]);

  useEffect(() => {
    const previous = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = previous; };
  }, []);

  if (!current || typeof document === 'undefined') return null;

  const navBtn: React.CSSProperties = {
    position: 'absolute',
    top: '50%',
    transform: 'translateY(-50%)',
    width: 40,
    height: 40,
    borderRadius: '50%',
    background: 'rgba(0,0,0,0.55)',
    border: 'none',
    color: '#fff',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    zIndex: 2,
  };

  return createPortal(
    <div
      className="fixed inset-0 flex items-center justify-center p-4"
      style={{ zIndex: 1000 }}
    >
      {/* Backdrop — click closes */}
      <button
        type="button"
        aria-label="Close image"
        onClick={onClose}
        style={{
          position: 'absolute',
          inset: 0,
          border: 'none',
          cursor: 'default',
          background: 'rgba(0,0,0,0.8)',
          backdropFilter: 'blur(8px)',
        }}
      />

      {/* Card */}
      <div
        className="modal-panel-in"
        onClick={(e) => e.stopPropagation()}
        style={{
          position: 'relative',
          zIndex: 1,
          display: 'flex',
          flexDirection: 'column',
          width: 'auto',
          maxWidth: '92vw',
          maxHeight: '92vh',
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          overflow: 'hidden',
          boxShadow: '0 12px 48px rgba(0,0,0,0.55)',
        }}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 12,
            padding: '12px 14px 12px 18px',
            borderBottom: '1px solid var(--border)',
            flexShrink: 0,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
            {title && (
              <span
                style={{
                  fontSize: 14,
                  fontWeight: 700,
                  color: 'var(--text)',
                  letterSpacing: '-0.01em',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {title}
              </span>
            )}
            {multi && (
              <span style={{ fontSize: 12, color: 'var(--text-tertiary)', flexShrink: 0 }}>
                {index + 1}/{images.length}
              </span>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="btn-ghost"
            style={{
              width: 32,
              height: 32,
              borderRadius: '50%',
              border: 'none',
              background: 'transparent',
              color: 'var(--text)',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              flexShrink: 0,
            }}
          >
            <X size={18} />
          </button>
        </div>

        {/* Body */}
        <div
          style={{
            position: 'relative',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'var(--bg)',
            padding: 12,
            minHeight: 0,
            flex: 1,
          }}
        >
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={current}
            alt={title ?? ''}
            style={{
              maxWidth: '88vw',
              maxHeight: '78vh',
              width: 'auto',
              height: 'auto',
              objectFit: 'contain',
              borderRadius: 8,
              display: 'block',
              userSelect: 'none',
            }}
          />

          {multi && hasPrev && (
            <button
              type="button"
              aria-label="Previous image"
              onClick={() => onNavigate(index - 1)}
              style={{ ...navBtn, left: 12 }}
            >
              <ChevronLeft size={22} />
            </button>
          )}
          {multi && hasNext && (
            <button
              type="button"
              aria-label="Next image"
              onClick={() => onNavigate(index + 1)}
              style={{ ...navBtn, right: 12 }}
            >
              <ChevronRight size={22} />
            </button>
          )}
        </div>

        {/* Footer dots */}
        {multi && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              gap: 6,
              padding: '10px 0 12px',
              borderTop: '1px solid var(--border)',
              flexShrink: 0,
            }}
          >
            {images.map((_, i) => (
              <button
                key={i}
                type="button"
                aria-label={`Go to image ${i + 1}`}
                onClick={() => onNavigate(i)}
                style={{
                  width: i === index ? 20 : 8,
                  height: 8,
                  borderRadius: 4,
                  background: i === index ? 'var(--text)' : 'var(--border)',
                  border: 'none',
                  cursor: 'pointer',
                  padding: 0,
                  transition: 'width 0.15s, background 0.15s',
                }}
              />
            ))}
          </div>
        )}
      </div>
    </div>,
    document.body,
  );
}
