'use client';

import { ReactNode } from 'react';
import { Dialog, DialogContent } from '@/components/ui/dialog';

interface DetailModalProps {
  open: boolean;
  onClose: () => void;
  hero?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
  maxWidth?: string;
}

export function DetailModal({ open, onClose, hero, children, footer, maxWidth = 'max-w-xl' }: DetailModalProps) {
  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className={maxWidth}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
          {hero}
          <div className="admin-detail-grid">{children}</div>
          {footer && <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>{footer}</div>}
        </div>
      </DialogContent>
    </Dialog>
  );
}

interface DetailRowProps {
  label: string;
  children: ReactNode;
}

export function DetailRow({ label, children }: DetailRowProps) {
  return (
    <div className="admin-detail-row">
      <div className="admin-detail-label">{label}</div>
      <div className="admin-detail-value">{children}</div>
    </div>
  );
}
