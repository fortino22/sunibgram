'use client';

import { X } from 'lucide-react';
import { useComposeStore } from '@/stores/composeStore';
import { PostComposer } from './PostComposer';

export function NewPostModal() {
  const { open, closeCompose } = useComposeStore();
  if (!open) return null;

  return (
    <div
      onClick={closeCompose}
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.55)',
        zIndex: 200,
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'center',
        paddingTop: 80,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: '100%',
          maxWidth: 600,
          background: 'var(--bg-card)',
          borderRadius: 16,
          overflow: 'hidden',
          boxShadow: '0 8px 40px rgba(0,0,0,0.4)',
          margin: '0 16px',
        }}
      >
        <div style={{
          display: 'flex',
          alignItems: 'center',
          padding: '12px 16px 0',
        }}>
          <button
            onClick={closeCompose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 6,
              borderRadius: 8,
            }}
          >
            <X size={20} />
          </button>
        </div>
        <PostComposer placeholder="What is happening?!" onSuccess={closeCompose} />
      </div>
    </div>
  );
}
