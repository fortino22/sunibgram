'use client';

import { useEffect } from 'react';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';
import type { ThreadMediaResponse } from '@/lib/services/threads';

interface ImageLightboxProps {
  media: ThreadMediaResponse[];
  currentIndex: number;
  onClose: () => void;
  onNavigate: (index: number) => void;
}

export function ImageLightbox({ media, currentIndex, onClose, onNavigate }: ImageLightboxProps) {
  const current = media[currentIndex];
  const hasPrev = currentIndex > 0;
  const hasNext = currentIndex < media.length - 1;

  useEffect(() => {
    const original = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = original; };
  }, []);

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowLeft' && hasPrev) onNavigate(currentIndex - 1);
      if (e.key === 'ArrowRight' && hasNext) onNavigate(currentIndex + 1);
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [currentIndex, hasPrev, hasNext, onClose, onNavigate]);

  if (!current) return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 1000,
        background: 'rgba(0,0,0,0.92)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
      onClick={onClose}
    >
      <button
        type="button"
        onClick={onClose}
        aria-label="Close"
        style={{
          position: 'absolute',
          top: 16,
          left: 16,
          width: 36,
          height: 36,
          borderRadius: '50%',
          background: 'rgba(255,255,255,0.14)',
          border: 'none',
          color: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
          zIndex: 10,
        }}
      >
        <X size={18} />
      </button>

      {hasPrev && (
        <button
          type="button"
          aria-label="Previous image"
          onClick={(e) => { e.stopPropagation(); onNavigate(currentIndex - 1); }}
          style={{
            position: 'absolute',
            left: 16,
            top: '50%',
            transform: 'translateY(-50%)',
            width: 44,
            height: 44,
            borderRadius: '50%',
            background: 'rgba(255,255,255,0.14)',
            border: 'none',
            color: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            zIndex: 10,
          }}
        >
          <ChevronLeft size={24} />
        </button>
      )}

      <img
        src={current.url}
        alt=""
        onClick={(e) => e.stopPropagation()}
        style={{
          maxWidth: '90vw',
          maxHeight: '90vh',
          objectFit: 'contain',
          borderRadius: 8,
          display: 'block',
          userSelect: 'none',
        }}
      />

      {hasNext && (
        <button
          type="button"
          aria-label="Next image"
          onClick={(e) => { e.stopPropagation(); onNavigate(currentIndex + 1); }}
          style={{
            position: 'absolute',
            right: 16,
            top: '50%',
            transform: 'translateY(-50%)',
            width: 44,
            height: 44,
            borderRadius: '50%',
            background: 'rgba(255,255,255,0.14)',
            border: 'none',
            color: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            zIndex: 10,
          }}
        >
          <ChevronRight size={24} />
        </button>
      )}

      {media.length > 1 && (
        <div
          style={{
            position: 'absolute',
            bottom: 24,
            left: '50%',
            transform: 'translateX(-50%)',
            display: 'flex',
            gap: 6,
          }}
        >
          {media.map((_, i) => (
            <button
              key={i}
              type="button"
              aria-label={`Go to image ${i + 1}`}
              onClick={(e) => { e.stopPropagation(); onNavigate(i); }}
              style={{
                width: i === currentIndex ? 20 : 8,
                height: 8,
                borderRadius: 4,
                background: i === currentIndex ? '#fff' : 'rgba(255,255,255,0.35)',
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
  );
}
