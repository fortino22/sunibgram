'use client';

import { useEffect, useRef, useState } from 'react';
import { AlertTriangle } from 'lucide-react';
import { Dialog, DialogContent } from '@/components/ui/dialog';

interface DangerConfirmModalProps {
  open: boolean;
  title: string;
  description?: string;
  /** When set, user must type this exact value to enable confirm button. */
  confirmText?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  busy?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

export function DangerConfirmModal({
  open,
  title,
  description,
  confirmText,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  busy,
  onConfirm,
  onClose,
}: DangerConfirmModalProps) {
  const [input, setInput] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!open) {
      setInput('');
    } else if (confirmText) {
      const t = setTimeout(() => inputRef.current?.focus(), 50);
      return () => clearTimeout(t);
    }
  }, [open, confirmText]);

  const requiresInput = !!confirmText;
  const matched = !requiresInput || input.trim() === confirmText;
  const disabled = busy || !matched;

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v && !busy) onClose(); }}>
      <DialogContent className="max-w-[420px] !p-0 !rounded-2xl overflow-hidden">
        <div style={{ padding: '22px 22px 18px' }}>
          <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
            <div
              style={{
                flexShrink: 0,
                width: 38,
                height: 38,
                borderRadius: 12,
                background: 'rgba(244, 33, 46, 0.12)',
                color: '#f4212e',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <AlertTriangle size={18} strokeWidth={2.5} />
            </div>
            <div style={{ flex: 1, minWidth: 0, paddingTop: 2 }}>
              <h2
                style={{
                  fontSize: 17,
                  fontWeight: 700,
                  color: 'var(--text)',
                  margin: 0,
                  marginBottom: 6,
                  letterSpacing: '-0.01em',
                }}
              >
                {title}
              </h2>
              {description && (
                <p
                  style={{
                    fontSize: 13.5,
                    color: 'var(--text-secondary)',
                    margin: 0,
                    lineHeight: 1.5,
                  }}
                >
                  {description}
                </p>
              )}
            </div>
          </div>

          {requiresInput && (
            <div style={{ marginTop: 18 }}>
              <div
                style={{
                  fontSize: 12.5,
                  color: 'var(--text-tertiary)',
                  marginBottom: 8,
                }}
              >
                Type{' '}
                <code
                  style={{
                    padding: '2px 6px',
                    borderRadius: 5,
                    background: 'var(--bg-subtle)',
                    border: '1px solid var(--border)',
                    color: 'var(--text)',
                    fontSize: 12,
                    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                    fontWeight: 600,
                  }}
                >
                  {confirmText}
                </code>{' '}
                to confirm
              </div>
              <input
                ref={inputRef}
                type="text"
                autoComplete="off"
                spellCheck={false}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !disabled) onConfirm();
                }}
                disabled={busy}
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  borderRadius: 10,
                  border: `1px solid ${
                    input.length === 0
                      ? 'var(--border)'
                      : matched
                      ? 'rgba(244, 33, 46, 0.55)'
                      : 'var(--border)'
                  }`,
                  background: 'var(--bg-subtle, rgba(255,255,255,0.03))',
                  color: 'var(--text)',
                  fontSize: 14,
                  fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                  outline: 'none',
                  boxSizing: 'border-box',
                  transition: 'border-color 0.15s, box-shadow 0.15s',
                  boxShadow: matched && input.length > 0 ? '0 0 0 3px rgba(244, 33, 46, 0.15)' : 'none',
                }}
              />
            </div>
          )}
        </div>

        <div
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            gap: 8,
            padding: '14px 22px',
            borderTop: '1px solid var(--border)',
            background: 'rgba(255,255,255,0.015)',
          }}
        >
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            style={{
              padding: '8px 16px',
              borderRadius: 9999,
              background: 'transparent',
              border: 'none',
              color: 'var(--text-secondary)',
              fontWeight: 600,
              fontSize: 14,
              cursor: busy ? 'not-allowed' : 'pointer',
              opacity: busy ? 0.5 : 1,
            }}
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={disabled}
            style={{
              padding: '8px 18px',
              borderRadius: 9999,
              background: disabled ? 'rgba(244, 33, 46, 0.35)' : '#f4212e',
              border: 'none',
              color: '#fff',
              fontWeight: 700,
              fontSize: 14,
              cursor: disabled ? 'not-allowed' : 'pointer',
              transition: 'background 0.15s, transform 0.05s',
              minWidth: 110,
            }}
            onMouseDown={(e) => {
              if (!disabled) (e.currentTarget as HTMLButtonElement).style.transform = 'scale(0.97)';
            }}
            onMouseUp={(e) => {
              (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1)';
            }}
          >
            {busy ? 'Working…' : confirmLabel}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
