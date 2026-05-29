'use client';

import { useEffect, useState } from 'react';
import { Link2, MessageCircle, Share2 } from 'lucide-react';
import { toast } from 'sonner';
import { useRecordShare } from '@/hooks/useShareThread';
import { buildThreadShareUrl } from '@/lib/shareUrl';

interface SharePopoverProps {
  threadId: number;
  anchorRect: DOMRect;
  onClose: () => void;
  onOpenChatModal: () => void;
}

const POPOVER_WIDTH = 220;
const GAP = 8;

interface NavigatorShare {
  share?: (data: { url?: string; title?: string; text?: string }) => Promise<void>;
}

export function SharePopover({ threadId, anchorRect, onClose, onOpenChatModal }: SharePopoverProps) {
  const recordShare = useRecordShare(threadId);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const url = buildThreadShareUrl(threadId);

  const supportsNativeShare =
    typeof navigator !== 'undefined' && typeof (navigator as NavigatorShare).share === 'function';

  const anchorCenterX = anchorRect.left + anchorRect.width / 2;
  const viewportW = typeof window !== 'undefined' ? window.innerWidth : 1024;
  const viewportH = typeof window !== 'undefined' ? window.innerHeight : 768;
  const popoverLeft = Math.max(8, Math.min(viewportW - POPOVER_WIDTH - 8, anchorCenterX - POPOVER_WIDTH / 2));
  const spaceBelow = viewportH - anchorRect.bottom;
  const placeBelow = spaceBelow > 220;
  const positionStyle: React.CSSProperties = placeBelow
    ? { top: anchorRect.bottom + GAP }
    : { bottom: viewportH - anchorRect.top + GAP };

  async function handleCopy() {
    if (busy) return;
    setBusy(true);
    try {
      await navigator.clipboard.writeText(url);
      toast.success('Link copied to clipboard');
      recordShare.mutate();
    } catch {
      toast.error('Failed to copy link');
    } finally {
      setBusy(false);
      onClose();
    }
  }

  async function handleNativeShare() {
    if (busy) return;
    const nav = navigator as NavigatorShare;
    if (!nav.share) {
      await handleCopy();
      return;
    }
    setBusy(true);
    try {
      await nav.share({ url });
      recordShare.mutate();
    } catch {
      // user cancelled — no toast
    } finally {
      setBusy(false);
      onClose();
    }
  }

  function handleChat() {
    onOpenChatModal();
  }

  return (
    <>
      <div
        onClick={onClose}
        style={{ position: 'fixed', inset: 0, zIndex: 1500, background: 'transparent' }}
      />
      <div
        role="menu"
        aria-label="Share options"
        onClick={(e) => e.stopPropagation()}
        style={{
          position: 'fixed',
          left: popoverLeft,
          ...positionStyle,
          width: POPOVER_WIDTH,
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 14,
          boxShadow: '0 12px 32px rgba(0,0,0,0.5)',
          zIndex: 1501,
          padding: 6,
          display: 'flex',
          flexDirection: 'column',
          gap: 2,
        }}
      >
        <MenuRow icon={<Link2 size={18} strokeWidth={2} />} label="Copy link" onClick={handleCopy} disabled={busy} />
        {supportsNativeShare && (
          <MenuRow icon={<Share2 size={18} strokeWidth={2} />} label="Share post via …" onClick={handleNativeShare} disabled={busy} />
        )}
        <MenuRow icon={<MessageCircle size={18} strokeWidth={2} />} label="Send via Chat" onClick={handleChat} disabled={busy} />
      </div>
    </>
  );
}

interface MenuRowProps {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  disabled?: boolean;
}

function MenuRow({ icon, label, onClick, disabled }: MenuRowProps) {
  return (
    <button
      type="button"
      role="menuitem"
      onClick={onClick}
      disabled={disabled}
      onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--sidebar-hover)'; }}
      onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '10px 12px',
        background: 'transparent',
        border: 'none',
        borderRadius: 10,
        color: 'var(--text)',
        fontSize: 14,
        fontFamily: 'inherit',
        fontWeight: 500,
        cursor: disabled ? 'wait' : 'pointer',
        textAlign: 'left',
      }}
    >
      {icon}
      <span>{label}</span>
    </button>
  );
}
