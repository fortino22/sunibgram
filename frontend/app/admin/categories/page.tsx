'use client';

import { useMemo, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Tags } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
import { Pager } from '@/components/admin/Pager';
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
  const [page, setPage] = useState(1);
  const PAGE_SIZE = 20;

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

  const rows = useMemo(() => q.data ?? [], [q.data]);
  const total = rows.length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const pageRows = useMemo(
    () => rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE),
    [rows, page],
  );

  const columns: Column<Category>[] = [
    { key: 'name', header: 'Name', sortValue: (r) => r.name?.toLowerCase(), render: (r) => <span style={{ fontWeight: 600 }}>{r.name}</span> },
    { key: 'id',   header: 'ID',   render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{r.id}</code> },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (r) => (
        <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={() => { setEditing(r); setName(r.name); }}
            className="admin-btn admin-btn-outline admin-btn-sm"
          >Edit</button>
          <button
            type="button"
            onClick={() => setDeleteId(r.id)}
            className="admin-btn admin-btn-outline-danger admin-btn-sm"
          >Delete</button>
        </div>
      ),
    },
  ];

  return (
    <AdminPage
      icon={Tags}
      title="Categories"
      description="Manage community categories."
      actions={
        <button
          type="button"
          onClick={() => { setCreating(true); setName(''); }}
          className="admin-btn admin-btn-primary"
        >
          + New Category
        </button>
      }
    >
      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Category>
            columns={columns}
            rows={pageRows}
            rowKey={(r) => r.id}
            loading={q.isLoading}
            empty="No categories yet"
            emptyIcon={<Tags size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
          />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
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
        description="Communities currently using this category will no longer be filterable by it."
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setDeleteId(null)}
      />
    </AdminPage>
  );
}
