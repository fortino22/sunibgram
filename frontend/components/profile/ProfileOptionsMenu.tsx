'use client';

import { MoreHorizontal, Flag, Ban, VolumeX, Link2, Info } from 'lucide-react';
import { type CSSProperties, type MouseEvent as ReactMouseEvent, useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { buildProfileShareUrl } from '@/lib/shareUrl';

interface ProfileOptionsMenuProps {
  username: string;
  userId: number;
  isBlocked: boolean;
  isMuted: boolean;
  onBlock?: () => void;
  onMute?: () => void;
  onReport: () => void;
}

const ITEM_BASE: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 12,
  width: '100%',
  padding: '10px 16px',
  background: 'transparent',
  border: 'none',
  fontSize: 14,
  fontWeight: 700,
  color: 'var(--text)',
  textAlign: 'left',
  cursor: 'pointer',
  minHeight: 40,
  lineHeight: 1.2,
  whiteSpace: 'nowrap',
};

const ICON_SIZE = 17;

export function ProfileOptionsMenu({ username, userId, isBlocked, isMuted, onBlock, onMute, onReport }: ProfileOptionsMenuProps) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const router = useRouter();

  useEffect(() => {
    if (!open) return undefined;

    function handlePointerDown(event: globalThis.MouseEvent | TouchEvent) {
      const target = event.target;
      if (!(target instanceof Node) || menuRef.current?.contains(target) === true) return;
      setOpen(false);
    }

    document.addEventListener('mousedown', handlePointerDown);
    document.addEventListener('touchstart', handlePointerDown);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      document.removeEventListener('touchstart', handlePointerDown);
    };
  }, [open]);

  function onItemEnter(event: ReactMouseEvent<HTMLButtonElement>) {
    event.currentTarget.style.background = 'var(--sidebar-hover)';
  }

  function onItemLeave(event: ReactMouseEvent<HTMLButtonElement>) {
    event.currentTarget.style.background = 'transparent';
  }

  function run(callback: () => void) {
    return () => {
      setOpen(false);
      callback();
    };
  }

  return (
    <div ref={menuRef} style={{ position: 'relative', flexShrink: 0 }}>
      <button
        type="button"
        aria-label="More options"
        aria-expanded={open}
        onClick={() => setOpen((current) => !current)}
        className="btn-ghost"
        style={{
          border: '1px solid var(--border-strong)',
          borderRadius: '50%',
          width: 36,
          height: 36,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
          color: 'var(--text)',
        }}
      >
        <MoreHorizontal size={17} strokeWidth={1.8} />
      </button>

      {open && (
        <div
          role="menu"
          style={{
            position: 'absolute',
            right: 0,
            top: '100%',
            zIndex: 50,
            marginTop: 6,
            borderRadius: 14,
            border: '1px solid var(--border)',
            background: 'var(--bg-card)',
            boxShadow: '0 10px 28px rgba(0, 0, 0, 0.35)',
            padding: '6px 0',
            overflow: 'hidden',
          }}
        >
          <button
            type="button"
            role="menuitem"
            onClick={run(() => router.push(`/about-profile/${userId}`))}
            onMouseEnter={onItemEnter}
            onMouseLeave={onItemLeave}
            style={ITEM_BASE}
          >
            <Info size={ICON_SIZE} strokeWidth={2} />
            About this account
          </button>
          <button
            type="button"
            role="menuitem"
            onClick={run(async () => {
              try {
                await navigator.clipboard.writeText(buildProfileShareUrl(userId));
                toast.success('Link copied to clipboard');
              } catch {
                toast.error('Failed to copy link');
              }
            })}
            onMouseEnter={onItemEnter}
            onMouseLeave={onItemLeave}
            style={ITEM_BASE}
          >
            <Link2 size={ICON_SIZE} strokeWidth={2} />
            Copy link to profile
          </button>
          {onMute && (
            <button
              type="button"
              role="menuitem"
              onClick={run(onMute)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              style={ITEM_BASE}
            >
              <VolumeX size={ICON_SIZE} strokeWidth={2} />
              {isMuted ? `Unmute @${username}` : `Mute @${username}`}
            </button>
          )}
          {onBlock && (
            <button
              type="button"
              role="menuitem"
              onClick={run(onBlock)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              style={{ ...ITEM_BASE, color: 'var(--coral)' }}
            >
              <Ban size={ICON_SIZE} strokeWidth={2} />
              {isBlocked ? `Unblock @${username}` : `Block @${username}`}
            </button>
          )}
          <button
            type="button"
            role="menuitem"
            onClick={run(onReport)}
            onMouseEnter={onItemEnter}
            onMouseLeave={onItemLeave}
            style={ITEM_BASE}
          >
            <Flag size={ICON_SIZE} strokeWidth={2} />
            Report user
          </button>
        </div>
      )}
    </div>
  );
}
