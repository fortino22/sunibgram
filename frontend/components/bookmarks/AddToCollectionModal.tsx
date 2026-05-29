'use client';

import { useState } from 'react';
import { Check, FolderPlus, Loader2, X } from 'lucide-react';
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

interface AddToCollectionModalProps {
  threadId: number;
  /** Collections this thread is already part of (used to render the selected state). */
  selectedCollectionIds?: ReadonlyArray<number>;
  onClose: () => void;
}

export function AddToCollectionModal({
  threadId,
  selectedCollectionIds = [],
  onClose,
}: AddToCollectionModalProps) {
  const collectionsQuery = useBookmarkCollections();
  const addMutation = useAddThreadToBookmarkCollection();
  const removeMutation = useRemoveThreadFromBookmarkCollection();
  const createMutation = useCreateBookmarkCollection();

  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [pendingCollectionId, setPendingCollectionId] = useState<number | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const collections = collectionsQuery.data?.data ?? [];
  const selectedSet = new Set(selectedCollectionIds);
  const busy = addMutation.isPending || removeMutation.isPending || createMutation.isPending;

  function handleClose() {
    if (busy) return;
    onClose();
  }

  async function handleToggle(collection: BookmarkCollection) {
    setErrorMessage(null);
    setPendingCollectionId(collection.id);
    try {
      const alreadyIn = selectedSet.has(collection.id);
      if (alreadyIn) {
        await removeMutation.mutateAsync({ collectionId: collection.id, threadId });
      } else {
        await addMutation.mutateAsync({ collectionId: collection.id, threadId });
      }
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : 'Something went wrong.');
    } finally {
      setPendingCollectionId(null);
    }
  }

  async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const trimmed = newName.trim();
    if (!trimmed) return;
    setErrorMessage(null);
    try {
      const created = await createMutation.mutateAsync({ name: trimmed });
      await addMutation.mutateAsync({ collectionId: created.collection_id, threadId });
      setCreating(false);
      setNewName('');
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : 'Could not create collection.');
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 100,
        padding: 16,
      }}
      onClick={(e) => { if (e.target === e.currentTarget) handleClose(); }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          width: '100%',
          maxWidth: 460,
          maxHeight: '90vh',
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
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
            background: 'var(--bg-card)',
          }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
            Save to collection
          </span>
          <button
            type="button"
            onClick={handleClose}
            disabled={busy}
            style={{
              background: 'none',
              border: 'none',
              cursor: busy ? 'not-allowed' : 'pointer',
              color: 'var(--text)',
              padding: 4,
              borderRadius: '50%',
              display: 'flex',
            }}
          >
            <X size={20} strokeWidth={2} />
          </button>
        </div>

        <div style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12, overflowY: 'auto' }}>
          {collectionsQuery.isLoading ? (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          ) : collections.length === 0 ? (
            <div style={{ padding: '12px 0', color: 'var(--text-tertiary)', fontSize: 14, textAlign: 'center' }}>
              You don&apos;t have any collections yet.
            </div>
          ) : (
            <ul style={{ display: 'flex', flexDirection: 'column', gap: 6, listStyle: 'none', margin: 0, padding: 0 }}>
              {collections.map((c) => {
                const checked = selectedSet.has(c.id);
                const pending = pendingCollectionId === c.id;
                return (
                  <li key={c.id}>
                    <button
                      type="button"
                      onClick={() => handleToggle(c)}
                      disabled={pending}
                      style={{
                        width: '100%',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        gap: 12,
                        padding: '10px 12px',
                        borderRadius: 8,
                        border: `1px solid ${checked ? 'var(--text)' : 'var(--border)'}`,
                        background: checked ? 'var(--bg-subtle)' : 'transparent',
                        color: 'var(--text)',
                        cursor: pending ? 'not-allowed' : 'pointer',
                        fontFamily: 'inherit',
                        textAlign: 'left',
                      }}
                    >
                      <div style={{ minWidth: 0 }}>
                        <div style={{ fontSize: 14, fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {c.name}
                        </div>
                        <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>
                          {c.item_count === 1 ? '1 post' : `${c.item_count} posts`}
                        </div>
                      </div>
                      <span style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: 22, height: 22, color: 'var(--text)' }}>
                        {pending ? <Loader2 size={16} className="animate-spin" /> : checked ? <Check size={18} strokeWidth={2.5} /> : null}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}

          {creating ? (
            <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
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
                  padding: '10px 12px',
                  borderRadius: 8,
                  border: '1px solid var(--border)',
                  background: 'var(--bg-subtle)',
                  color: 'var(--text)',
                  fontSize: 14,
                  fontFamily: 'inherit',
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
                <button
                  type="button"
                  onClick={() => { setCreating(false); setNewName(''); setErrorMessage(null); }}
                  disabled={createMutation.isPending}
                  style={{
                    padding: '8px 14px',
                    borderRadius: 999,
                    background: 'transparent',
                    border: '1px solid var(--border)',
                    color: 'var(--text)',
                    fontWeight: 600,
                    cursor: 'pointer',
                    fontFamily: 'inherit',
                    fontSize: 13,
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!newName.trim() || createMutation.isPending}
                  style={{
                    padding: '8px 16px',
                    borderRadius: 999,
                    background: 'var(--text)',
                    color: 'var(--bg)',
                    border: 'none',
                    fontWeight: 700,
                    cursor: 'pointer',
                    opacity: !newName.trim() || createMutation.isPending ? 0.5 : 1,
                    fontFamily: 'inherit',
                    fontSize: 13,
                  }}
                >
                  {createMutation.isPending ? 'Creating…' : 'Create & add'}
                </button>
              </div>
            </form>
          ) : (
            <button
              type="button"
              onClick={() => setCreating(true)}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 8,
                padding: '10px 14px',
                borderRadius: 8,
                border: '1px dashed var(--border)',
                background: 'transparent',
                color: 'var(--text)',
                fontWeight: 600,
                fontSize: 14,
                fontFamily: 'inherit',
                cursor: 'pointer',
              }}
            >
              <FolderPlus size={16} strokeWidth={2} />
              New collection
            </button>
          )}

          {errorMessage && (
            <div style={{ fontSize: 13, color: 'var(--coral)' }}>{errorMessage}</div>
          )}
        </div>
      </div>
    </div>
  );
}
