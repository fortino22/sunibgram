'use client';

import { useEffect } from 'react';
import { X } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';

interface AvatarZoomModalProps {
  open: boolean;
  onClose: () => void;
  username: string;
  avatarUrl?: string | null;
}

export function AvatarZoomModal({ open, onClose, username, avatarUrl }: AvatarZoomModalProps) {
  useEffect(() => {
    if (!open) return undefined;
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [open, onClose]);

  useEffect(() => {
    if (!open) return undefined;
    const previous = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = previous; };
  }, [open]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 flex items-center justify-center p-4"
      style={{ zIndex: 100 }}
    >
      <button
        type="button"
        aria-label="Close profile photo"
        className="absolute inset-0 admin-modal-backdrop"
        onClick={onClose}
        style={{ background: 'rgba(0, 0, 0, 0.8)', backdropFilter: 'blur(8px)' }}
      />

      <button
        type="button"
        aria-label="Close profile photo"
        onClick={onClose}
        className="admin-modal-close"
        style={{ position: 'absolute', top: 20, left: 20, zIndex: 10 }}
      >
        <X size={20} />
      </button>

      <div className="modal-panel-in relative" style={{ zIndex: 1 }}>
        <Avatar
          username={username}
          avatarUrl={avatarUrl}
          size={360}
          style={{
            border: '4px solid var(--bg-card)',
            boxShadow: '0 12px 48px rgba(0, 0, 0, 0.55)',
            maxWidth: '90vw',
            maxHeight: '90vw',
          }}
        />
      </div>
    </div>
  );
}
