'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
import {
  listCategories,
  createCategory,
  updateCategory,
  deleteCategory,
} from '@/lib/services/admin/categories';
import type { Category } from '@/lib/services/admin/types';

export default function AdminCategoriesPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [editing, setEditing] = useState<Category | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [name, setName] = useState('');

  const q = useQuery({
    queryKey: ['admin', 'categories'],
    queryFn: () => listCategories(token!),
    enabled: !!token,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: ['admin', 'categories'] });

  const createMut = useMutation({
    mutationFn: () => createCategory(token!, name),
    onSuccess: () => { invalidate(); setCreating(false); setName(''); },
  });
  const updateMut = useMutation({
    mutationFn: () => updateCategory(token!, editing!.id, name),
    onSuccess: () => { invalidate(); setEditing(null); },
  });
  const deleteMut = useMutation({
    mutationFn: () => deleteCategory(token!, deleteId!),
    onSuccess: () => { invalidate(); setDeleteId(null); },
  });

  const columns: Column<Category>[] = [
    { key: 'name', header: 'Name', render: (r) => <span style={{ fontWeight: 600 }}>{r.name}</span> },
    { key: 'id',   header: 'ID',   render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{r.id}</code> },
    {
      key: 'actions',
      header: '',
      render: (r) => (
        <div style={{ display: 'flex', gap: 6 }}>
          <button
            type="button"
            onClick={() => { setEditing(r); setName(r.name); }}
            className="btn-ghost"
            style={{ padding: '6px 12px', border: '1px solid var(--border)', borderRadius: 9999, background: 'transparent', color: 'var(--text)', cursor: 'pointer', fontSize: 13 }}
          >Edit</button>
          <button
            type="button"
            onClick={() => setDeleteId(r.id)}
            className="btn-ghost"
            style={{ padding: '6px 12px', border: '1px solid var(--border)', borderRadius: 9999, background: 'transparent', color: 'var(--coral)', cursor: 'pointer', fontSize: 13 }}
          >Delete</button>
        </div>
      ),
    },
  ];

  return (
    <AdminPage
      title="Categories"
      description="Manage community categories."
      actions={
        <button
          type="button"
          onClick={() => { setCreating(true); setName(''); }}
          className="btn-solid"
          style={{ padding: '8px 16px', borderRadius: 9999, background: 'var(--accent)', color: '#fff', border: 'none', fontWeight: 600, cursor: 'pointer' }}
        >
          + New Category
        </button>
      }
    >
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : (
        <DataTable<Category> columns={columns} rows={q.data ?? []} rowKey={(r) => r.id} />
      )}

      <CrudFormModal
        open={creating}
        title="New Category"
        busy={createMut.isPending}
        onClose={() => setCreating(false)}
        onSubmit={() => createMut.mutate()}
      >
        <Field label="Name" value={name} onChange={setName} required />
      </CrudFormModal>

      <CrudFormModal
        open={editing !== null}
        title="Edit Category"
        busy={updateMut.isPending}
        onClose={() => setEditing(null)}
        onSubmit={() => updateMut.mutate()}
      >
        <Field label="Name" value={name} onChange={setName} required />
      </CrudFormModal>

      <ConfirmModal
        open={deleteId !== null}
        title="Delete this category?"
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setDeleteId(null)}
      />
    </AdminPage>
  );
}
