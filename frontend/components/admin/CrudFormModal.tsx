'use client';

import { ReactNode } from 'react';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';

interface CrudFormModalProps {
  open: boolean;
  title: string;
  submitLabel?: string;
  busy?: boolean;
  onClose: () => void;
  onSubmit: () => void;
  children: ReactNode;
}

export function CrudFormModal({
  open,
  title,
  submitLabel = 'Save',
  busy,
  onClose,
  onSubmit,
  children,
}: CrudFormModalProps) {
  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (!busy) onSubmit();
          }}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {children}
          </div>
          <DialogFooter>
            <button
              type="button"
              onClick={onClose}
              disabled={busy}
              className="btn-ghost"
              style={{
                padding: '8px 16px',
                borderRadius: 9999,
                background: 'transparent',
                border: '1px solid var(--border)',
                color: 'var(--text)',
                fontWeight: 500,
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={busy}
              className="btn-solid"
              style={{
                padding: '8px 18px',
                borderRadius: 9999,
                background: 'var(--accent)',
                border: 'none',
                color: '#fff',
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              {busy ? 'Saving…' : submitLabel}
            </button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

interface FieldProps {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  required?: boolean;
  textarea?: boolean;
}

export function Field({ label, value, onChange, placeholder, required, textarea }: FieldProps) {
  const common = {
    value,
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => onChange(e.target.value),
    placeholder,
    required,
    style: {
      width: '100%',
      padding: '10px 14px',
      borderRadius: 12,
      border: '1px solid var(--border)',
      background: 'var(--bg-card)',
      color: 'var(--text)',
      fontSize: 14,
      fontFamily: 'inherit',
      outline: 'none',
      resize: textarea ? ('vertical' as const) : undefined,
      minHeight: textarea ? 90 : undefined,
    },
  };
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <span style={{ fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)' }}>{label}</span>
      {textarea ? <textarea {...common} /> : <input type="text" {...common} />}
    </label>
  );
}
