'use client';

import { useState } from 'react';
import { X } from 'lucide-react';
import { MAX_COLLECTION_NAME_LENGTH } from '@/lib/services/bookmarkCollections';

interface CollectionFormValues {
  name: string;
  description: string;
}

interface CollectionFormModalProps {
  mode: 'create' | 'edit';
  initialName?: string;
  initialDescription?: string | null;
  busy?: boolean;
  errorMessage?: string | null;
  onSubmit: (values: CollectionFormValues) => void;
  onClose: () => void;
}

export function CollectionFormModal({
  mode,
  initialName = '',
  initialDescription = '',
  busy,
  errorMessage,
  onSubmit,
  onClose,
}: CollectionFormModalProps) {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription ?? '');

  const title = mode === 'create' ? 'New collection' : 'Edit collection';
  const submitLabel = mode === 'create' ? 'Create' : 'Save changes';
  const trimmedName = name.trim();
  const canSubmit = trimmedName.length > 0 && trimmedName.length <= MAX_COLLECTION_NAME_LENGTH && !busy;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 100,
        padding: 16,
      }}
      onClick={(e) => { if (e.target === e.currentTarget && !busy) onClose(); }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          width: '100%',
          maxWidth: 460,
          maxHeight: '90vh',
          overflowY: 'auto',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
            position: 'sticky',
            top: 0,
            background: 'var(--bg-card)',
            zIndex: 1,
          }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
            {title}
          </span>
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            style={{
              background: 'none',
              border: 'none',
              cursor: busy ? 'not-allowed' : 'pointer',
              color: 'var(--text)',
              padding: 4,
              borderRadius: '50%',
              display: 'flex',
            }}
          >
            <X size={20} strokeWidth={2} />
          </button>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (!canSubmit) return;
            onSubmit({
              name: trimmedName,
              description: description.trim(),
            });
          }}
          style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 16 }}
        >
          <div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 6,
              }}
            >
              <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>
                Name
              </label>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                {trimmedName.length}/{MAX_COLLECTION_NAME_LENGTH}
              </span>
            </div>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Study resources"
              maxLength={MAX_COLLECTION_NAME_LENGTH}
              autoFocus
              disabled={busy}
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: 8,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                fontSize: 14,
                boxSizing: 'border-box',
                fontFamily: 'inherit',
                outline: 'none',
              }}
            />
          </div>

          <div>
            <label
              style={{
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--text-secondary)',
                display: 'block',
                marginBottom: 6,
              }}
            >
              Description (optional)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What goes in this collection?"
              disabled={busy}
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: 8,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                fontSize: 14,
                boxSizing: 'border-box',
                minHeight: 96,
                resize: 'vertical',
                fontFamily: 'inherit',
                outline: 'none',
              }}
            />
          </div>

          {errorMessage && (
            <div style={{ fontSize: 13, color: 'var(--coral)' }}>{errorMessage}</div>
          )}

          <button
            type="submit"
            disabled={!canSubmit}
            style={{
              width: '100%',
              padding: '12px',
              borderRadius: 999,
              border: 'none',
              background: 'var(--text)',
              color: 'var(--bg)',
              fontSize: 15,
              fontWeight: 700,
              cursor: !canSubmit ? 'default' : 'pointer',
              opacity: !canSubmit ? 0.5 : 1,
              fontFamily: 'inherit',
            }}
          >
            {busy ? 'Saving…' : submitLabel}
          </button>
        </form>
      </div>
    </div>
  );
}
