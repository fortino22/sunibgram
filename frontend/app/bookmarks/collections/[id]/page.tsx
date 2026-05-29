'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, Bookmark, MoreHorizontal, Pencil, Trash2 } from 'lucide-react';
import { EmptyState } from '@/components/shared/EmptyState';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { CollectionFormModal } from '@/components/bookmarks/CollectionFormModal';
import { DeletePostModal } from '@/components/shared/DeletePostModal';
import {
  useBookmarkCollection,
  useBookmarkCollectionThreads,
  useDeleteBookmarkCollection,
  useRemoveThreadFromBookmarkCollection,
  useUpdateBookmarkCollection,
} from '@/hooks/useBookmarkCollections';
import type { ThreadItem } from '@/lib/services/threads';

export default function CollectionDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const routeId = Number(params.id);
  const isValidId = Number.isInteger(routeId) && routeId > 0;

  const collectionId = isValidId ? routeId : null;

  const collectionQuery = useBookmarkCollection(collectionId);
  const threadsQuery = useBookmarkCollectionThreads(collectionId);
  const updateMutation = useUpdateBookmarkCollection();
  const deleteMutation = useDeleteBookmarkCollection();
  const removeThreadMutation = useRemoveThreadFromBookmarkCollection();

  const [menuOpen, setMenuOpen] = useState(false);
  const [editing, setEditing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [removingThreadId, setRemovingThreadId] = useState<number | null>(null);

  const collection = collectionQuery.data;
  const threads = threadsQuery.data?.data ?? [];

  async function submitEdit({ name, description }: { name: string; description: string }) {
    if (!collection) return;
    setFormError(null);
    try {
      await updateMutation.mutateAsync({
        id: collection.id,
        input: { name, description: description ? description : undefined },
      });
      setEditing(false);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Could not save changes.');
    }
  }

  async function handleDelete() {
    if (!collection) return;
    try {
      await deleteMutation.mutateAsync(collection.id);
      router.push('/bookmarks');
    } catch {
      setConfirmDelete(false);
    }
  }

  async function handleRemoveThread(threadId: number) {
    if (!collection) return;
    setRemovingThreadId(threadId);
    try {
      await removeThreadMutation.mutateAsync({ collectionId: collection.id, threadId });
    } finally {
      setRemovingThreadId(null);
    }
  }

  if (!isValidId) {
    return <EmptyState title="Collection not found" subtitle="That collection does not exist." />;
  }

  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '12px 20px',
          }}
        >
          <button
            type="button"
            onClick={() => router.push('/bookmarks')}
            aria-label="Back to bookmarks"
            style={{
              width: 36,
              height: 36,
              borderRadius: 9999,
              background: 'transparent',
              border: 'none',
              color: 'var(--text)',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
          >
            <ArrowLeft size={20} strokeWidth={2} />
          </button>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div
              style={{
                fontWeight: 700,
                fontSize: 18,
                color: 'var(--text)',
                letterSpacing: '-0.02em',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {collectionQuery.isLoading ? 'Loading…' : collection?.name ?? 'Collection'}
            </div>
            {collection && (
              <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>
                {collection.item_count === 1 ? '1 post' : `${collection.item_count} posts`}
              </div>
            )}
          </div>
          {collection && (
            <div style={{ position: 'relative' }}>
              <button
                type="button"
                aria-label="Collection actions"
                onClick={() => setMenuOpen((v) => !v)}
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: 9999,
                  background: menuOpen ? 'var(--bg-hover)' : 'transparent',
                  border: 'none',
                  color: 'var(--text-tertiary)',
                  cursor: 'pointer',
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <MoreHorizontal size={18} />
              </button>
              {menuOpen && (
                <div
                  role="menu"
                  onMouseLeave={() => setMenuOpen(false)}
                  style={{
                    position: 'absolute',
                    top: 42,
                    right: 0,
                    minWidth: 180,
                    background: 'var(--bg-card)',
                    border: '1px solid var(--border)',
                    borderRadius: 12,
                    padding: 6,
                    boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
                    zIndex: 20,
                  }}
                >
                  <button
                    type="button"
                    role="menuitem"
                    onClick={() => { setMenuOpen(false); setEditing(true); }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
                    style={{
                      width: '100%',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10,
                      padding: '9px 12px',
                      background: 'transparent',
                      border: 'none',
                      borderRadius: 8,
                      cursor: 'pointer',
                      color: 'var(--text)',
                      fontSize: 13,
                      fontFamily: 'inherit',
                      fontWeight: 500,
                      textAlign: 'left',
                    }}
                  >
                    <Pencil size={14} />
                    <span>Edit collection</span>
                  </button>
                  <button
                    type="button"
                    role="menuitem"
                    onClick={() => { setMenuOpen(false); setConfirmDelete(true); }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--coral-soft)'; }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
                    style={{
                      width: '100%',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10,
                      padding: '9px 12px',
                      background: 'transparent',
                      border: 'none',
                      borderRadius: 8,
                      cursor: 'pointer',
                      color: 'var(--coral)',
                      fontSize: 13,
                      fontFamily: 'inherit',
                      fontWeight: 500,
                      textAlign: 'left',
                    }}
                  >
                    <Trash2 size={14} />
                    <span>Delete collection</span>
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {collection?.description && (
          <div style={{ padding: '0 20px 14px', color: 'var(--text-secondary)', fontSize: 14, lineHeight: 1.6 }}>
            {collection.description}
          </div>
        )}
      </div>

      {collectionQuery.isError ? (
        <EmptyState title="Collection not found" subtitle="This collection may have been deleted." />
      ) : threadsQuery.isLoading ? (
        Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)
      ) : threadsQuery.isError ? (
        <EmptyState title="Couldn’t load posts" subtitle="Try again in a moment." />
      ) : threads.length === 0 ? (
        <div style={{ padding: '64px 24px', textAlign: 'center', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 10 }}>
          <Bookmark size={32} strokeWidth={1.6} color="var(--text-tertiary)" />
          <div style={{ fontSize: 16, fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.02em' }}>
            No posts in this collection yet
          </div>
          <div style={{ fontSize: 14, color: 'var(--text-tertiary)', maxWidth: 320, lineHeight: 1.6 }}>
            Save a bookmark to this collection from your bookmarks list.
          </div>
        </div>
      ) : (
        threads.map((thread: ThreadItem) => (
          <CollectionThreadRow
            key={thread.id}
            thread={thread}
            removing={removingThreadId === thread.id}
            onRemove={() => handleRemoveThread(thread.id)}
          />
        ))
      )}

      {editing && collection && (
        <CollectionFormModal
          mode="edit"
          initialName={collection.name}
          initialDescription={collection.description ?? ''}
          busy={updateMutation.isPending}
          errorMessage={formError}
          onSubmit={submitEdit}
          onClose={() => { if (!updateMutation.isPending) { setEditing(false); setFormError(null); } }}
        />
      )}

      <DeletePostModal
        open={confirmDelete}
        isPending={deleteMutation.isPending}
        onConfirm={handleDelete}
        onCancel={() => { if (!deleteMutation.isPending) setConfirmDelete(false); }}
        title="Delete collection?"
        description={
          collection
            ? `“${collection.name}” will be removed. The bookmarked posts will still be in your bookmarks.`
            : "This collection will be removed. The bookmarked posts will still be in your bookmarks."
        }
      />
    </div>
  );
}

interface CollectionThreadRowProps {
  thread: ThreadItem;
  removing: boolean;
  onRemove: () => void;
}

function CollectionThreadRow({ thread, removing, onRemove }: CollectionThreadRowProps) {
  return (
    <div style={{ position: 'relative' }}>
      <ThreadCard thread={thread} />
      <button
        type="button"
        onClick={onRemove}
        disabled={removing}
        aria-label="Remove from collection"
        title="Remove from collection"
        style={{
          position: 'absolute',
          top: 14,
          right: 14,
          height: 30,
          padding: '0 12px',
          borderRadius: 9999,
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          color: 'var(--coral)',
          cursor: removing ? 'not-allowed' : 'pointer',
          display: 'inline-flex',
          alignItems: 'center',
          gap: 6,
          fontSize: 12,
          fontWeight: 600,
          fontFamily: 'inherit',
          zIndex: 5,
          opacity: removing ? 0.6 : 1,
        }}
        onMouseEnter={(e) => { if (!removing) e.currentTarget.style.background = 'var(--coral-soft)'; }}
        onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--bg-card)'; }}
      >
        <Trash2 size={13} strokeWidth={2} />
        {removing ? 'Removing…' : 'Remove'}
      </button>
    </div>
  );
}
