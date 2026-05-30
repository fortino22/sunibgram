'use client';

import { AlertTriangle } from 'lucide-react';
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';

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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div style={{ display: 'flex', alignItems: 'flex-start', gap: 14 }}>
            {destructive && (
              <span
                aria-hidden
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 40,
                  height: 40,
                  borderRadius: 12,
                  background: 'var(--bg-subtle)',
                  color: 'var(--text)',
                  flexShrink: 0,
                  border: '1px solid var(--border)',
                }}
              >
                <AlertTriangle size={20} strokeWidth={2.2} />
              </span>
            )}
            <div style={{ minWidth: 0, flex: 1 }}>
              {description && <DialogDescription style={{ marginTop: 0 }}>{description}</DialogDescription>}
            </div>
          </div>
        </DialogBody>
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
            className="admin-btn admin-btn-primary"
          >
            {destructive && !busy && (
              <AlertTriangle size={14} strokeWidth={2.4} style={{ color: 'var(--coral)' }} />
            )}
            {busy ? 'Working…' : confirmLabel}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
