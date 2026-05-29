'use client';

import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { EditEventDialog } from '@/components/events/EditEventDialog';
import { DangerConfirmModal } from '@/components/shared/DangerConfirmModal';
import { useDeleteEvent } from '@/hooks/useDeleteEvent';
import type { EventDetail, MyEventItem } from '@/lib/services/events';

interface Props {
  event: MyEventItem;
}

export function MyEventCardActions({ event }: Props) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const deleteMutation = useDeleteEvent();

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    if (menuOpen) document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [menuOpen]);

  const canEdit = event.status === 'revision';

  const editEventForDialog: EventDetail = {
    ...event,
    documentUrl: undefined,
  };

  function handleDelete() {
    deleteMutation.mutate(event.id, {
      onSuccess: () => setDeleteOpen(false),
    });
  }

  return (
    <>
      <div ref={menuRef} style={{ position: 'absolute', top: 10, right: 10, zIndex: 5 }}>
        <button
          type="button"
          aria-label="Event actions"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            setMenuOpen((o) => !o);
          }}
          style={{
            width: 32,
            height: 32,
            borderRadius: '50%',
            border: '1px solid oklch(100% 0 0 / 0.2)',
            background: 'oklch(0% 0 0 / 0.55)',
            backdropFilter: 'blur(6px)',
            color: 'var(--text)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'background 0.15s ease, border-color 0.15s ease',
          }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'oklch(0% 0 0 / 0.75)'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'oklch(0% 0 0 / 0.55)'; }}
        >
          <MoreHorizontal size={16} />
        </button>

        {menuOpen && (
          <div
            onClick={(e) => e.stopPropagation()}
            style={{
              position: 'absolute',
              top: 38,
              right: 0,
              zIndex: 50,
              background: 'var(--bg-card, var(--bg))',
              border: '1px solid var(--border)',
              borderRadius: 12,
              minWidth: 170,
              overflow: 'hidden',
              boxShadow: '0 12px 32px oklch(0% 0 0 / 0.6)',
            }}
          >
            {canEdit && (
              <button
                type="button"
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  setMenuOpen(false);
                  setEditOpen(true);
                }}
                style={{
                  width: '100%',
                  padding: '11px 14px',
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer',
                  fontSize: 13,
                  fontWeight: 600,
                  color: 'var(--text)',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  fontFamily: 'inherit',
                  textAlign: 'left',
                }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
              >
                <Pencil size={14} />
                Edit event
              </button>
            )}
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setMenuOpen(false);
                setDeleteOpen(true);
              }}
              style={{
                width: '100%',
                padding: '11px 14px',
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                fontSize: 13,
                fontWeight: 600,
                color: 'var(--coral, #e53e3e)',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                fontFamily: 'inherit',
                textAlign: 'left',
              }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
            >
              <Trash2 size={14} />
              Delete event
            </button>
          </div>
        )}
      </div>

      {editOpen && (
        <EditEventDialog open={editOpen} onOpenChange={setEditOpen} event={editEventForDialog} />
      )}

      <DangerConfirmModal
        open={deleteOpen}
        title="Delete event?"
        description="This can't be undone and the event will be removed for all participants."
        confirmText={event.title}
        confirmLabel="Delete event"
        busy={deleteMutation.isPending}
        onConfirm={handleDelete}
        onClose={() => { if (!deleteMutation.isPending) setDeleteOpen(false); }}
      />
    </>
  );
}
