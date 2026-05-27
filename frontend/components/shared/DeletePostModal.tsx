'use client';

import { useEffect } from 'react';
import { createPortal } from 'react-dom';
import { LoaderCircle } from 'lucide-react';

interface DeletePostModalProps {
  open: boolean;
  isPending?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  title?: string;
  description?: string;
}

export function DeletePostModal({ open, isPending = false, onConfirm, onCancel, title = 'Delete post?', description = "This can't be undone and it will be removed from your profile, the timeline of any accounts that follow you, and from search results." }: DeletePostModalProps) {
  useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = prev; };
  }, [open]);

  if (!open) return null;

  return createPortal(
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="delete-modal-title"
      onClick={isPending ? undefined : onCancel}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 1000,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(100,100,100,0.55)',
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: '#000',
          borderRadius: 20,
          padding: '32px 28px 24px',
          width: 300,
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}
      >
        <p
          id="delete-modal-title"
          style={{ fontSize: 20, fontWeight: 800, color: '#fff', letterSpacing: '-0.02em', marginBottom: 4 }}
        >
          {title}
        </p>
        <p style={{ fontSize: 14, color: 'rgb(113,118,123)', lineHeight: 1.5, marginBottom: 16 }}>
          {description}
        </p>
        <button
          type="button"
          className="btn-solid"
          disabled={isPending}
          onClick={onConfirm}
          style={{
            width: '100%',
            padding: '14px 0',
            borderRadius: 999,
            border: 'none',
            background: '#f4212e',
            color: '#fff',
            fontSize: 15,
            fontWeight: 700,
            cursor: isPending ? 'not-allowed' : 'pointer',
            opacity: isPending ? 0.7 : 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 8,
            fontFamily: 'inherit',
          }}
        >
          {isPending && <LoaderCircle size={15} className="animate-spin" />}
          Delete
        </button>
        <button
          type="button"
          disabled={isPending}
          onClick={onCancel}
          onMouseEnter={(e) => { if (!isPending) e.currentTarget.style.background = 'var(--bg-hover)'; }}
          onMouseLeave={(e) => { if (!isPending) e.currentTarget.style.background = 'transparent'; }}
          style={{
            width: '100%',
            padding: '14px 0',
            borderRadius: 999,
            border: '1px solid rgb(51,54,57)',
            background: 'transparent',
            color: '#fff',
            fontSize: 15,
            fontWeight: 700,
            cursor: isPending ? 'not-allowed' : 'pointer',
            opacity: isPending ? 0.7 : 1,
            transition: 'background 0.12s ease',
            fontFamily: 'inherit',
          }}
        >
          Cancel
        </button>
      </div>
    </div>,
    document.body
  );
}
