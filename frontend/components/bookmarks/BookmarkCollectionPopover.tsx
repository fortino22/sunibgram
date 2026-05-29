'use client';

import { useEffect, useState } from 'react';
import { Bookmark, Check, Loader2, Plus, X } from 'lucide-react';
import {
  useAddThreadToBookmarkCollection,
  useBookmarkCollections,
  useCreateBookmarkCollection,
  useRemoveThreadFromBookmarkCollection,
} from '@/hooks/useBookmarkCollections';
import {
  MAX_COLLECTION_NAME_LENGTH,
  type BookmarkCollection,
} from '@/lib/services/bookmarkCollections';

interface BookmarkCollectionPopoverProps {
  threadId: number;
  anchorRect: DOMRect;
  onClose: () => void;
}

const POPOVER_WIDTH = 280;
const ARROW_SIZE = 8;
const GAP = 10;

export function BookmarkCollectionPopover({ threadId, anchorRect, onClose }: BookmarkCollectionPopoverProps) {
  const collectionsQuery = useBookmarkCollections();
  const addMutation = useAddThreadToBookmarkCollection();
  const removeMutation = useRemoveThreadFromBookmarkCollection();
  const createMutation = useCreateBookmarkCollection();

  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [pendingId, setPendingId] = useState<number | null>(null);
  const [memberIds, setMemberIds] = useState<Set<number>>(new Set());
  const [error, setError] = useState<string | null>(null);

  const collections = collectionsQuery.data?.data ?? [];
  const busy = addMutation.isPending || removeMutation.isPending || createMutation.isPending;

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const anchorCenterX = anchorRect.left + anchorRect.width / 2;
  const popoverLeft = Math.max(8, Math.min(window.innerWidth - POPOVER_WIDTH - 8, anchorCenterX - POPOVER_WIDTH / 2));
  const arrowLeft = anchorCenterX - popoverLeft;
  const popoverBottom = window.innerHeight - anchorRect.top + GAP;

  async function handleToggle(c: BookmarkCollection) {
    if (busy) return;
    setError(null);
    setPendingId(c.id);
    const isMember = memberIds.has(c.id);
    try {
      if (isMember) {
        await removeMutation.mutateAsync({ collectionId: c.id, threadId });
        setMemberIds((prev) => {
          const next = new Set(prev);
          next.delete(c.id);
          return next;
        });
      } else {
        await addMutation.mutateAsync({ collectionId: c.id, threadId });
        setMemberIds((prev) => new Set(prev).add(c.id));
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not update collection.';
      if (message.includes('already in collection')) {
        setMemberIds((prev) => new Set(prev).add(c.id));
      } else if (message.includes('not in collection')) {
        setMemberIds((prev) => {
          const next = new Set(prev);
          next.delete(c.id);
          return next;
        });
      } else {
        setError(message);
      }
    } finally {
      setPendingId(null);
    }
  }

  async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const trimmed = newName.trim();
    if (!trimmed) return;
    setError(null);
    try {
      const created = await createMutation.mutateAsync({ name: trimmed });
      await addMutation.mutateAsync({ collectionId: created.collection_id, threadId });
      setMemberIds((prev) => new Set(prev).add(created.collection_id));
      setCreating(false);
      setNewName('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create collection.');
    }
  }

  return (
    <>
      <div
        onClick={onClose}
        style={{ position: 'fixed', inset: 0, zIndex: 1500, background: 'transparent' }}
      />
      <div
        role="dialog"
        aria-label="Save to collection"
        onClick={(e) => e.stopPropagation()}
        style={{
          position: 'fixed',
          left: popoverLeft,
          bottom: popoverBottom,
          width: POPOVER_WIDTH,
          maxHeight: 360,
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 14,
          boxShadow: '0 12px 32px rgba(0,0,0,0.5)',
          zIndex: 1501,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '10px 14px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.01em' }}>
            Collections
          </span>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            {!creating && (
              <button
                type="button"
                aria-label="New collection"
                onClick={() => { setCreating(true); setError(null); }}
                disabled={busy}
                style={iconBtnStyle(busy)}
              >
                <Plus size={16} strokeWidth={2.2} />
              </button>
            )}
            <button
              type="button"
              aria-label="Close"
              onClick={onClose}
              style={iconBtnStyle(false)}
            >
              <X size={16} strokeWidth={2.2} />
            </button>
          </div>
        </div>

        <div style={{ overflowY: 'auto', padding: 6 }}>
          {creating && (
            <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: 6, padding: '6px 8px 10px' }}>
              <input
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="Collection name"
                maxLength={MAX_COLLECTION_NAME_LENGTH}
                autoFocus
                disabled={createMutation.isPending}
                style={{
                  width: '100%',
                  padding: '8px 10px',
                  borderRadius: 8,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-subtle)',
                  color: 'var(--text)',
                  fontSize: 13,
                  fontFamily: 'inherit',
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 6 }}>
                <button
                  type="button"
                  onClick={() => { setCreating(false); setNewName(''); setError(null); }}
                  disabled={createMutation.isPending}
                  style={pillBtnStyle('ghost')}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!newName.trim() || createMutation.isPending}
                  style={{ ...pillBtnStyle('solid'), opacity: !newName.trim() || createMutation.isPending ? 0.5 : 1 }}
                >
                  {createMutation.isPending ? 'Saving…' : 'Save'}
                </button>
              </div>
            </form>
          )}

          {collectionsQuery.isLoading ? (
            <div style={{ padding: '16px', textAlign: 'center', fontSize: 13, color: 'var(--text-tertiary)' }}>
              Loading…
            </div>
          ) : collections.length === 0 && !creating ? (
            <div style={{ padding: '16px 12px', textAlign: 'center', fontSize: 13, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
              No collections yet.
            </div>
          ) : (
            <ul style={{ listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
              {collections.map((c) => {
                const pending = pendingId === c.id;
                const isMember = memberIds.has(c.id);
                const rowBusy = busy && !pending;
                return (
                  <li key={c.id}>
                    <button
                      type="button"
                      onClick={() => handleToggle(c)}
                      disabled={pending || rowBusy}
                      aria-pressed={isMember}
                      onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
                      onMouseLeave={(e) => { e.currentTarget.style.background = isMember ? 'var(--bg-subtle)' : 'transparent'; }}
                      style={{
                        width: '100%',
                        display: 'flex',
                        alignItems: 'center',
                        gap: 10,
                        padding: '8px 10px',
                        borderRadius: 8,
                        border: 'none',
                        background: isMember ? 'var(--bg-subtle)' : 'transparent',
                        color: 'var(--text)',
                        cursor: pending || rowBusy ? 'not-allowed' : 'pointer',
                        fontFamily: 'inherit',
                        textAlign: 'left',
                      }}
                    >
                      <span
                        aria-hidden
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: 8,
                          background: 'var(--accent-soft)',
                          display: 'inline-flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          color: 'var(--accent)',
                          flexShrink: 0,
                        }}
                      >
                        <Bookmark size={15} strokeWidth={2} fill={isMember ? 'var(--accent)' : 'none'} />
                      </span>
                      <span
                        style={{
                          flex: 1,
                          minWidth: 0,
                          fontSize: 13,
                          fontWeight: 600,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {c.name}
                      </span>
                      <span
                        aria-hidden
                        style={{
                          width: 20,
                          height: 20,
                          borderRadius: 6,
                          border: `1.5px solid ${isMember ? 'var(--accent)' : 'var(--border)'}`,
                          background: isMember ? 'var(--accent)' : 'transparent',
                          color: '#fff',
                          display: 'inline-flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          flexShrink: 0,
                        }}
                      >
                        {pending ? <Loader2 size={12} className="animate-spin" color="var(--accent)" /> : isMember ? <Check size={14} strokeWidth={3} /> : null}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}

          {error && (
            <div style={{ padding: '6px 10px', fontSize: 12, color: 'var(--coral)' }}>{error}</div>
          )}
        </div>

        <span
          aria-hidden
          style={{
            position: 'absolute',
            bottom: -ARROW_SIZE,
            left: arrowLeft - ARROW_SIZE,
            width: 0,
            height: 0,
            borderLeft: `${ARROW_SIZE}px solid transparent`,
            borderRight: `${ARROW_SIZE}px solid transparent`,
            borderTop: `${ARROW_SIZE}px solid var(--bg-card)`,
            filter: 'drop-shadow(0 1px 0 var(--border))',
          }}
        />
      </div>
    </>
  );
}

function iconBtnStyle(disabled: boolean): React.CSSProperties {
  return {
    width: 26,
    height: 26,
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: 6,
    background: 'transparent',
    border: 'none',
    color: 'var(--text)',
    cursor: disabled ? 'not-allowed' : 'pointer',
    opacity: disabled ? 0.5 : 1,
  };
}

function pillBtnStyle(variant: 'solid' | 'ghost'): React.CSSProperties {
  return {
    padding: '6px 12px',
    borderRadius: 999,
    border: variant === 'ghost' ? '1px solid var(--border)' : 'none',
    background: variant === 'ghost' ? 'transparent' : 'var(--text)',
    color: variant === 'ghost' ? 'var(--text)' : 'var(--bg)',
    fontWeight: 600,
    fontSize: 12,
    fontFamily: 'inherit',
    cursor: 'pointer',
  };
}
