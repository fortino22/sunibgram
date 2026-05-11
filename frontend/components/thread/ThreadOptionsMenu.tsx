'use client';

import { MoreHorizontal } from 'lucide-react';
import { type MouseEvent as ReactMouseEvent, useEffect, useRef, useState } from 'react';

interface ThreadOptionsMenuProps {
  onEdit: () => void;
  onDelete: () => void;
  deleting: boolean;
  disabled?: boolean;
  compact?: boolean;
}

export function ThreadOptionsMenu({
  onEdit,
  onDelete,
  deleting,
  disabled = false,
  compact = false,
}: ThreadOptionsMenuProps) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const iconSize = compact ? 16 : 18;

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    function handlePointerDown(event: globalThis.MouseEvent | TouchEvent) {
      const target = event.target;
      if (!(target instanceof Node) || menuRef.current?.contains(target) === true) {
        return;
      }

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
    if (disabled) {
      return;
    }

    setOpen((current) => !current);
  }

  function handleEdit(event: ReactMouseEvent<HTMLButtonElement>) {
    stopEvent(event);
    setOpen(false);
    onEdit();
  }

  function handleDelete(event: ReactMouseEvent<HTMLButtonElement>) {
    stopEvent(event);
    setOpen(false);
    onDelete();
  }

  return (
    <div ref={menuRef} className="relative shrink-0" onClick={(event) => event.stopPropagation()}>
      <button
        type="button"
        aria-label="More options"
        aria-expanded={open}
        onClick={handleToggle}
        disabled={disabled}
        className="inline-flex h-8 w-8 items-center justify-center rounded-full transition-opacity"
        style={{
          color: 'var(--text-tertiary)',
          background: 'transparent',
          opacity: disabled ? 0.5 : 1,
          cursor: disabled ? 'not-allowed' : 'pointer',
        }}
      >
        <MoreHorizontal size={iconSize} strokeWidth={1.9} />
      </button>

      {open && (
        <div
          className="absolute right-0 top-full z-30 mt-2 min-w-[168px] rounded-xl border p-1.5 shadow-lg"
          style={{
            borderColor: 'var(--border)',
            background: 'var(--bg-card)',
            boxShadow: '0 18px 40px rgba(15, 20, 25, 0.16)',
          }}
        >
          <button
            type="button"
            onClick={handleEdit}
            className="flex min-h-10 w-full items-center rounded-lg px-4 py-2.5 text-left text-sm font-medium"
            style={{ color: 'var(--text)' }}
          >
            Edit
          </button>
          <button
            type="button"
            onClick={handleDelete}
            disabled={deleting}
            className="mt-1 flex min-h-10 w-full items-center rounded-lg px-4 py-2.5 text-left text-sm font-medium"
            style={{
              color: 'var(--coral)',
              opacity: deleting ? 0.5 : 1,
              cursor: deleting ? 'not-allowed' : 'pointer',
            }}
          >
            {deleting ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      )}
    </div>
  );
}
