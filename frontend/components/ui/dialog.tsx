'use client';

import * as React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DialogContextValue {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const DialogContext = React.createContext<DialogContextValue | null>(null);

export function Dialog({
  open,
  onOpenChange,
  children,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}) {
  return (
    <DialogContext.Provider value={{ open, onOpenChange }}>
      {children}
    </DialogContext.Provider>
  );
}

function useDialogContext() {
  const value = React.useContext(DialogContext);
  if (value === null) {
    throw new Error('Dialog components must be used within <Dialog />');
  }
  return value;
}

export function DialogContent({
  className,
  children,
}: {
  className?: string;
  children: React.ReactNode;
}) {
  const { open, onOpenChange } = useDialogContext();

  React.useEffect(() => {
    if (!open) {
      return undefined;
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onOpenChange(false);
      }
    }

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [open, onOpenChange]);

  React.useEffect(() => {
    if (!open) return;
    const previous = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = previous; };
  }, [open]);

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        aria-label="Close dialog"
        className="absolute inset-0 admin-modal-backdrop"
        onClick={() => onOpenChange(false)}
        style={{ background: 'rgba(2, 6, 23, 0.7)', backdropFilter: 'blur(6px)' }}
        type="button"
      />
      <div
        role="dialog"
        aria-modal="true"
        className={cn(
          'admin-modal-panel relative z-10 w-full flex flex-col',
          className,
        )}
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border-strong)',
          borderRadius: 20,
          boxShadow: '0 12px 48px rgba(0, 0, 0, 0.55)',
          maxWidth: 480,
          maxHeight: '90vh',
          overflow: 'hidden',
        }}
      >
        {children}
      </div>
    </div>
  );
}

interface DialogHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  showClose?: boolean;
}

export function DialogHeader({ className, showClose = true, children, ...props }: DialogHeaderProps) {
  const { onOpenChange } = useDialogContext();
  return (
    <div className={cn('admin-modal-header', className)} {...props}>
      <div className="admin-modal-header-text">{children}</div>
      {showClose && (
        <button
          type="button"
          aria-label="Close dialog"
          onClick={() => onOpenChange(false)}
          className="admin-modal-close"
        >
          <X size={18} />
        </button>
      )}
    </div>
  );
}

export function DialogTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h2 className={cn('text-lg font-semibold tracking-tight', className)} {...props} />;
}

export function DialogDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn('mt-1 text-sm leading-6', className)}
      style={{ color: 'var(--text-secondary)' }}
      {...props}
    />
  );
}

export function DialogBody({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('admin-modal-body hide-scrollbar', className)} {...props} />;
}

export function DialogFooter({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('admin-modal-footer', className)} {...props} />;
}
