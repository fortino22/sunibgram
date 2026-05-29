'use client';

import { ReactNode } from 'react';
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';

interface DetailModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  hero?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
}

export function DetailModal({ open, onClose, title = 'Detail', hero, children, footer }: DetailModalProps) {
  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            {hero}
            <div className="admin-detail-grid">{children}</div>
          </div>
        </DialogBody>
        {footer && <DialogFooter>{footer}</DialogFooter>}
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
