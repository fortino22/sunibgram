'use client';

import { MoreHorizontal, Pencil, Trash2, Pin, PinOff, Flag } from 'lucide-react';
import { type CSSProperties, type MouseEvent as ReactMouseEvent, useEffect, useRef, useState } from 'react';
import type { CommunityRole } from '@/lib/services/communities';

interface ThreadOptionsMenuProps {
  onEdit?: () => void;
  onDelete?: () => void;
  deleting?: boolean;
  disabled?: boolean;
  compact?: boolean;
  communityRole?: CommunityRole | null;
  isPinned?: boolean;
  pinning?: boolean;
  onPin?: () => void;
  onUnpin?: () => void;
  onReport?: () => void;
}

const ITEM_BASE: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 12,
  width: '100%',
  padding: '9px 14px',
  background: 'transparent',
  border: 'none',
  fontSize: 14,
  fontWeight: 600,
  color: 'var(--text)',
  textAlign: 'left',
  cursor: 'pointer',
  minHeight: 38,
  lineHeight: 1.2,
};

const ICON_SIZE = 16;

export function ThreadOptionsMenu({
  onEdit,
  onDelete,
  deleting = false,
  disabled = false,
  compact = false,
  communityRole,
  isPinned = false,
  pinning = false,
  onPin,
  onUnpin,
  onReport,
}: ThreadOptionsMenuProps) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const triggerIconSize = compact ? 16 : 18;
  const canCommunityPin = communityRole === 'ADMIN';

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

  function stopEvent(event: ReactMouseEvent<HTMLElement>) {
    event.preventDefault();
    event.stopPropagation();
  }

  function handleToggle(event: ReactMouseEvent<HTMLButtonElement>) {
    stopEvent(event);
    if (disabled) return;
    setOpen((current) => !current);
  }

  function wrap(callback?: () => void) {
    return (event: ReactMouseEvent<HTMLButtonElement>) => {
      stopEvent(event);
      setOpen(false);
      callback?.();
    };
  }

  function onItemEnter(event: ReactMouseEvent<HTMLButtonElement>) {
    if (event.currentTarget.disabled) return;
    event.currentTarget.style.background = 'var(--sidebar-hover)';
  }

  function onItemLeave(event: ReactMouseEvent<HTMLButtonElement>) {
    event.currentTarget.style.background = 'transparent';
  }

  return (
    <div
      ref={menuRef}
      style={{ position: 'relative', flexShrink: 0 }}
      onClick={(event) => event.stopPropagation()}
    >
      <button
        type="button"
        aria-label="More options"
        aria-expanded={open}
        onClick={handleToggle}
        disabled={disabled}
        onMouseEnter={(e) => { if (!disabled) e.currentTarget.style.background = 'var(--sidebar-hover)'; }}
        onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: 32,
          width: 32,
          borderRadius: '50%',
          border: 'none',
          color: 'var(--text-tertiary)',
          background: 'transparent',
          opacity: disabled ? 0.5 : 1,
          cursor: disabled ? 'not-allowed' : 'pointer',
          transition: 'background 0.15s ease',
        }}
      >
        <MoreHorizontal size={triggerIconSize} strokeWidth={1.9} />
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
            minWidth: 220,
            borderRadius: 14,
            border: '1px solid var(--border)',
            background: 'var(--bg-card)',
            boxShadow: '0 10px 28px rgba(0, 0, 0, 0.35)',
            padding: '6px 0',
            overflow: 'hidden',
          }}
        >
          {onEdit && (
            <button
              type="button"
              role="menuitem"
              onClick={wrap(onEdit)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              style={ITEM_BASE}
            >
              <Pencil size={ICON_SIZE} strokeWidth={2} />
              Edit post
            </button>
          )}

          {onDelete && (
            <button
              type="button"
              role="menuitem"
              onClick={wrap(onDelete)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              disabled={deleting}
              style={{
                ...ITEM_BASE,
                color: 'var(--coral)',
                opacity: deleting ? 0.5 : 1,
                cursor: deleting ? 'not-allowed' : 'pointer',
              }}
            >
              <Trash2 size={ICON_SIZE} strokeWidth={2} />
              {deleting ? 'Deleting...' : 'Delete post'}
            </button>
          )}

          {canCommunityPin && !isPinned && onPin && (
            <button
              type="button"
              role="menuitem"
              onClick={wrap(onPin)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              disabled={pinning}
              style={{
                ...ITEM_BASE,
                opacity: pinning ? 0.5 : 1,
                cursor: pinning ? 'not-allowed' : 'pointer',
              }}
            >
              <Pin size={ICON_SIZE} strokeWidth={2} />
              {pinning ? 'Pinning...' : 'Pin post'}
            </button>
          )}

          {canCommunityPin && isPinned && onUnpin && (
            <button
              type="button"
              role="menuitem"
              onClick={wrap(onUnpin)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              disabled={pinning}
              style={{
                ...ITEM_BASE,
                opacity: pinning ? 0.5 : 1,
                cursor: pinning ? 'not-allowed' : 'pointer',
              }}
            >
              <PinOff size={ICON_SIZE} strokeWidth={2} />
              {pinning ? 'Unpinning...' : 'Unpin post'}
            </button>
          )}

          {onReport && (
            <button
              type="button"
              role="menuitem"
              onClick={wrap(onReport)}
              onMouseEnter={onItemEnter}
              onMouseLeave={onItemLeave}
              style={ITEM_BASE}
            >
              <Flag size={ICON_SIZE} strokeWidth={2} />
              Report post
            </button>
          )}
        </div>
      )}
    </div>
  );
}
