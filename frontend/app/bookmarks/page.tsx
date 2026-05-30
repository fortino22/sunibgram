'use client';

import { useState } from 'react';
import { FolderPlus, Plus } from 'lucide-react';
import { EmptyState } from '@/components/shared/EmptyState';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { AddToCollectionModal } from '@/components/bookmarks/AddToCollectionModal';
import { CollectionCard } from '@/components/bookmarks/CollectionCard';
import { CollectionCardSkeleton } from '@/components/bookmarks/CollectionCardSkeleton';
import { CollectionFormModal } from '@/components/bookmarks/CollectionFormModal';
import { DeletePostModal } from '@/components/shared/DeletePostModal';
import { useMyBookmarks } from '@/hooks/useMyBookmarks';
import {
  useBookmarkCollections,
  useCreateBookmarkCollection,
  useDeleteBookmarkCollection,
  useUpdateBookmarkCollection,
} from '@/hooks/useBookmarkCollections';
import {
  MAX_COLLECTIONS_PER_USER,
  type BookmarkCollection,
} from '@/lib/services/bookmarkCollections';
import type { ThreadItem } from '@/lib/services/threads';

type Tab = 'all' | 'collections';

export default function BookmarksPage() {
  const [activeTab, setActiveTab] = useState<Tab>('all');
  const [formMode, setFormMode] = useState<'create' | 'edit' | null>(null);
  const [editingCollection, setEditingCollection] = useState<BookmarkCollection | null>(null);
  const [deletingCollection, setDeletingCollection] = useState<BookmarkCollection | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [addToCollectionThreadId, setAddToCollectionThreadId] = useState<number | null>(null);

  const bookmarksQuery = useMyBookmarks();
  const collectionsQuery = useBookmarkCollections();

  const createMutation = useCreateBookmarkCollection();
  const updateMutation = useUpdateBookmarkCollection();
  const deleteMutation = useDeleteBookmarkCollection();

  const bookmarks = bookmarksQuery.data?.data ?? [];
  const collections = collectionsQuery.data?.data ?? [];

  const atCollectionLimit = collections.length >= MAX_COLLECTIONS_PER_USER;

  function openCreate() {
    if (atCollectionLimit) return;
    setEditingCollection(null);
    setFormError(null);
    setFormMode('create');
  }

  function openEdit(collection: BookmarkCollection) {
    setEditingCollection(collection);
    setFormError(null);
    setFormMode('edit');
  }

  function closeForm() {
    if (createMutation.isPending || updateMutation.isPending) return;
    setFormMode(null);
    setEditingCollection(null);
    setFormError(null);
  }

  async function submitForm({ name, description }: { name: string; description: string }) {
    setFormError(null);
    try {
      if (formMode === 'create') {
        await createMutation.mutateAsync({
          name,
          description: description ? description : undefined,
        });
      } else if (formMode === 'edit' && editingCollection) {
        await updateMutation.mutateAsync({
          id: editingCollection.id,
          input: { name, description: description ? description : undefined },
        });
      }
      setFormMode(null);
      setEditingCollection(null);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Something went wrong.');
    }
  }

  async function confirmDelete() {
    if (!deletingCollection) return;
    try {
      await deleteMutation.mutateAsync(deletingCollection.id);
      setDeletingCollection(null);
    } catch {
      // surface via UI later if needed
      setDeletingCollection(null);
    }
  }

  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'var(--bg-blur)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '20px 20px 16px',
          }}
        >
          <div style={{ fontWeight: 500, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em' }}>
            Bookmarks
          </div>
          {activeTab === 'collections' && (
            <button
              type="button"
              onClick={openCreate}
              disabled={atCollectionLimit}
              title={atCollectionLimit ? `You can have at most ${MAX_COLLECTIONS_PER_USER} collections.` : undefined}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6,
                padding: '8px 14px',
                borderRadius: 9999,
                border: 'none',
                background: atCollectionLimit ? 'var(--bg-hover)' : 'var(--accent)',
                color: '#fff',
                fontWeight: 600,
                fontSize: 13,
                fontFamily: 'inherit',
                cursor: atCollectionLimit ? 'not-allowed' : 'pointer',
                opacity: atCollectionLimit ? 0.6 : 1,
              }}
            >
              <Plus size={14} strokeWidth={2.5} />
              New collection
            </button>
          )}
        </div>

        <div role="tablist" style={{ display: 'flex', gap: 0 }}>
          <TabButton label="All" active={activeTab === 'all'} onClick={() => setActiveTab('all')} />
          <TabButton
            label="Collections"
            active={activeTab === 'collections'}
            onClick={() => setActiveTab('collections')}
          />
        </div>
      </div>

      {activeTab === 'all' ? (
        bookmarksQuery.isLoading ? (
          Array.from({ length: 3 }, (_, index) => <ThreadCardSkeleton key={index} />)
        ) : bookmarksQuery.isError ? (
          <EmptyState title="Bookmarks unavailable" subtitle="Your bookmarked posts could not be loaded right now." />
        ) : bookmarks.length === 0 ? (
          <EmptyState title="No bookmarks yet" subtitle="Posts you save will appear here." />
        ) : (
          bookmarks.map((thread: ThreadItem) => (
            <BookmarkThreadRow
              key={thread.id}
              thread={thread}
              onAddToCollection={() => setAddToCollectionThreadId(thread.id)}
            />
          ))
        )
      ) : (
        <CollectionsTab
          isLoading={collectionsQuery.isLoading}
          isError={collectionsQuery.isError}
          collections={collections}
          onCreate={openCreate}
          onEdit={openEdit}
          onDelete={(c) => setDeletingCollection(c)}
        />
      )}

      {formMode !== null && (
        <CollectionFormModal
          mode={formMode}
          initialName={editingCollection?.name ?? ''}
          initialDescription={editingCollection?.description ?? ''}
          busy={createMutation.isPending || updateMutation.isPending}
          errorMessage={formError}
          onSubmit={submitForm}
          onClose={closeForm}
        />
      )}

      <DeletePostModal
        open={deletingCollection !== null}
        isPending={deleteMutation.isPending}
        onConfirm={confirmDelete}
        onCancel={() => { if (!deleteMutation.isPending) setDeletingCollection(null); }}
        title="Delete collection?"
        description={
          deletingCollection
            ? `“${deletingCollection.name}” will be removed. The bookmarked posts will still be in your bookmarks.`
            : "This collection will be removed. The bookmarked posts will still be in your bookmarks."
        }
      />

      {addToCollectionThreadId !== null && (
        <AddToCollectionModal
          threadId={addToCollectionThreadId}
          onClose={() => setAddToCollectionThreadId(null)}
        />
      )}
    </div>
  );
}

interface TabButtonProps {
  label: string;
  active: boolean;
  onClick: () => void;
}

function TabButton({ label, active, onClick }: TabButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      style={{
        flex: 1,
        padding: '14px 0',
        background: 'transparent',
        border: 'none',
        cursor: 'pointer',
        color: active ? 'var(--text)' : 'var(--text-tertiary)',
        fontFamily: 'inherit',
        fontWeight: active ? 700 : 500,
        fontSize: 14,
        position: 'relative',
      }}
    >
      {label}
      {active && (
        <span
          aria-hidden
          style={{
            position: 'absolute',
            bottom: 0,
            left: '50%',
            transform: 'translateX(-50%)',
            width: 48,
            height: 3,
            borderRadius: 2,
            background: 'var(--accent)',
          }}
        />
      )}
    </button>
  );
}

interface BookmarkThreadRowProps {
  thread: ThreadItem;
  onAddToCollection: () => void;
}

function BookmarkThreadRow({ thread, onAddToCollection }: BookmarkThreadRowProps) {
  return (
    <div style={{ position: 'relative' }}>
      <ThreadCard thread={thread} />
      <button
        type="button"
        onClick={onAddToCollection}
        aria-label="Save to collection"
        title="Save to collection"
        style={{
          position: 'absolute',
          top: 14,
          right: 14,
          width: 32,
          height: 32,
          borderRadius: 9999,
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          color: 'var(--accent)',
          cursor: 'pointer',
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 5,
        }}
        onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--accent-soft)'; }}
        onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--bg-card)'; }}
      >
        <FolderPlus size={16} strokeWidth={2} />
      </button>
    </div>
  );
}

interface CollectionsTabProps {
  isLoading: boolean;
  isError: boolean;
  collections: BookmarkCollection[];
  onCreate: () => void;
  onEdit: (collection: BookmarkCollection) => void;
  onDelete: (collection: BookmarkCollection) => void;
}

function CollectionsTab({ isLoading, isError, collections, onCreate, onEdit, onDelete }: CollectionsTabProps) {
  if (isLoading) {
    return (
      <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 12, padding: 16 }}>
        {Array.from({ length: 3 }, (_, i) => (
          <CollectionCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (isError) {
    return <EmptyState title="Collections unavailable" subtitle="We couldn't load your collections right now." />;
  }

  if (collections.length === 0) {
    return (
      <div style={{ padding: '64px 24px', textAlign: 'center', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 14 }}>
        <div style={{ fontSize: 20, fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.02em' }}>
          No collections yet
        </div>
        <div style={{ fontSize: 14, color: 'var(--text-tertiary)', maxWidth: 320, lineHeight: 1.6 }}>
          Group related bookmarks together. Up to {MAX_COLLECTIONS_PER_USER} collections.
        </div>
        <button
          type="button"
          onClick={onCreate}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
            padding: '10px 18px',
            borderRadius: 9999,
            border: 'none',
            background: 'var(--accent)',
            color: '#fff',
            fontWeight: 600,
            fontSize: 14,
            fontFamily: 'inherit',
            cursor: 'pointer',
          }}
        >
          <Plus size={14} strokeWidth={2.5} />
          Create your first collection
        </button>
      </div>
    );
  }

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 12, padding: 16 }}>
      {collections.map((c) => (
        <CollectionCard key={c.id} collection={c} onEdit={onEdit} onDelete={onDelete} />
      ))}
    </div>
  );
}
