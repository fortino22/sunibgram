'use client';

import { useRouter } from 'next/navigation';
import { X, ArrowRight } from 'lucide-react';

interface AboutIntroModalProps {
  /** Controlled by the onboarding gate in MainWrapper. */
  open: boolean;
  /** Dismiss — reveals the app underneath. */
  onClose: () => void;
}

// First-time welcome: a compact intro to Sunibgram that points users to the
// full /about page. Unlike the follow gate this is closable (X, outside click,
// "Got it"), and it is shown only once.
export function AboutIntroModal({ open, onClose }: AboutIntroModalProps) {
  const router = useRouter();

  if (!open) return null;

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 70,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
        padding: 16,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          position: 'relative',
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 18,
          width: '100%',
          maxWidth: 360,
          padding: '22px 22px 18px',
          boxShadow: '0 24px 60px -12px rgba(0,0,0,0.5)',
          textAlign: 'center',
        }}
      >
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="absolute flex items-center justify-center rounded-full transition-opacity"
          style={{
            top: 12,
            right: 12,
            width: 28,
            height: 28,
            border: 'none',
            background: 'transparent',
            color: 'var(--text-tertiary)',
            cursor: 'pointer',
            opacity: 0.7,
          }}
        >
          <X size={18} />
        </button>

        <h2 className="text-[19px] font-extrabold" style={{ marginTop: 4, color: 'var(--text)' }}>
          Welcome to Sunibgram
        </h2>
        <p className="text-[13.5px]" style={{ marginTop: 6, lineHeight: 1.5, color: 'var(--text-tertiary)' }}>
          The university exclusive space for Binus students. Share threads, join
          communities, and connect across campus.
        </p>

        <button
          type="button"
          onClick={() => {
            // Completing onboarding here too means returning from /about does
            // not reshow this intro (the gate is keyed off onboarding_completed).
            onClose();
            router.push('/about');
          }}
          className="flex w-full items-center justify-center gap-1.5 rounded-full text-[14px] font-bold transition-opacity hover:opacity-90"
          style={{
            marginTop: 18,
            padding: '11px 16px',
            background: 'var(--text)',
            color: 'var(--bg)',
            border: 'none',
            cursor: 'pointer',
          }}
        >
          Learn more
          <ArrowRight size={16} />
        </button>
        <button
          type="button"
          onClick={onClose}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = 'var(--text)';
            e.currentTarget.style.opacity = '1';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = 'var(--text-tertiary)';
            e.currentTarget.style.opacity = '0.85';
          }}
          className="text-[13.5px] font-semibold transition-colors hover:underline"
          style={{
            marginTop: 10,
            background: 'transparent',
            color: 'var(--text-tertiary)',
            border: 'none',
            cursor: 'pointer',
            opacity: 0.85,
          }}
        >
          Skip for now
        </button>
      </div>
    </div>
  );
}
