'use client';

import { AlertTriangle } from 'lucide-react';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';

interface ConfirmModalProps {
  open: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
  busy?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

export function ConfirmModal({
  open,
  title,
  description,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  destructive,
  busy,
  onConfirm,
  onClose,
}: ConfirmModalProps) {
  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
            {destructive && (
              <span
                aria-hidden
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 38,
                  height: 38,
                  borderRadius: 12,
                  background: 'var(--coral-soft)',
                  color: 'var(--coral)',
                  flexShrink: 0,
                }}
              >
                <AlertTriangle size={20} strokeWidth={2.2} />
              </span>
            )}
            <div style={{ minWidth: 0 }}>
              <DialogTitle>{title}</DialogTitle>
              {description && (
                <DialogDescription style={{ marginTop: 6 }}>{description}</DialogDescription>
              )}
            </div>
          </div>
        </DialogHeader>
        <DialogFooter>
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            className="admin-btn admin-btn-outline"
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={busy}
            className={`admin-btn ${destructive ? 'admin-btn-danger' : 'admin-btn-primary'}`}
          >
            {busy ? 'Working…' : confirmLabel}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
