'use client';

import { ReactNode } from 'react';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';

interface CrudFormModalProps {
  open: boolean;
  title: string;
  description?: string;
  submitLabel?: string;
  busy?: boolean;
  onClose: () => void;
  onSubmit: () => void;
  children: ReactNode;
}

export function CrudFormModal({
  open,
  title,
  description,
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
          {description && (
            <p style={{ marginTop: 4, fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5 }}>
              {description}
            </p>
          )}
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (!busy) onSubmit();
          }}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>{children}</div>
          <DialogFooter>
            <button
              type="button"
              onClick={onClose}
              disabled={busy}
              className="admin-btn admin-btn-outline"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={busy}
              className="admin-btn admin-btn-primary"
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
  hint?: string;
}

export function Field({ label, value, onChange, placeholder, required, textarea, hint }: FieldProps) {
  const common = {
    value,
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => onChange(e.target.value),
    placeholder,
    required,
  };
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <span style={{ fontSize: 12.5, fontWeight: 600, color: 'var(--text-secondary)', letterSpacing: '0.01em' }}>
        {label}
        {required && <span style={{ color: 'var(--coral)', marginLeft: 4 }}>*</span>}
      </span>
      {textarea
        ? <textarea {...common} className="admin-textarea" />
        : <input type="text" {...common} className="admin-input" />}
      {hint && <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{hint}</span>}
    </label>
  );
}
