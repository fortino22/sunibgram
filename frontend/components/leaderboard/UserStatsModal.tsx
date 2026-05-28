'use client';

import { useEffect } from 'react';
import { X } from 'lucide-react';
import { UserStatsPanel } from './UserStatsPanel';
import type { TrendingUser } from '@/lib/services/trendingUsers';

interface UserStatsModalProps {
  user: TrendingUser;
  rank?: number;
  onClose: () => void;
}

export function UserStatsModal({ user, rank, onClose }: UserStatsModalProps) {
  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose]);

  return (
    <div
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 60,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
        animation: 'modal-backdrop-in 0.18s ease',
        padding: 16,
      }}
    >
      <div
        style={{
          position: 'relative',
          width: '100%',
          maxWidth: 420,
          maxHeight: '85vh',
          overflowY: 'auto',
          animation: 'modal-panel-in 0.22s cubic-bezier(0.16, 1, 0.3, 1)',
        }}
      >
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          style={{
            position: 'absolute',
            top: 12,
            right: 12,
            zIndex: 1,
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border)',
            borderRadius: '50%',
            width: 32,
            height: 32,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            color: 'var(--text)',
          }}
        >
          <X size={18} strokeWidth={2} />
        </button>
        <UserStatsPanel user={user} rank={rank} />
      </div>
    </div>
  );
}
