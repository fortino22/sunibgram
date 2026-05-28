'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
import {
  listFacultiesPaginated,
  createFaculty,
  updateFaculty,
  deleteFaculty,
} from '@/lib/services/admin/faculties';
import type { Faculty } from '@/lib/services/admin/types';

export default function AdminFacultiesPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [editing, setEditing] = useState<Faculty | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [name, setName] = useState('');

  const q = useQuery({
    queryKey: ['admin', 'faculties', page],
    queryFn: () => listFacultiesPaginated(token!, page, 20),
    enabled: !!token,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: ['admin', 'faculties'] });

  const createMut = useMutation({
    mutationFn: () => createFaculty(token!, name),
    onSuccess: () => { invalidate(); setCreating(false); setName(''); },
  });
  const updateMut = useMutation({
    mutationFn: () => updateFaculty(token!, editing!.faculty_id, name),
    onSuccess: () => { invalidate(); setEditing(null); setName(''); },
  });
  const deleteMut = useMutation({
    mutationFn: () => deleteFaculty(token!, deleteId!),
    onSuccess: () => { invalidate(); setDeleteId(null); },
  });

  const columns: Column<Faculty>[] = [
    { key: 'id',    header: 'ID',    render: (r) => r.faculty_id, width: 80 },
    { key: 'name',  header: 'Name',  render: (r) => <span style={{ fontWeight: 600 }}>{r.faculty_name}</span> },
    { key: 'users', header: 'Users', render: (r) => r.total_users ?? 0 },
    {
      key: 'actions',
      header: '',
      render: (r) => (
        <div style={{ display: 'flex', gap: 6 }}>
          <button
            type="button"
            onClick={() => { setEditing(r); setName(r.faculty_name); }}
            className="btn-ghost"
            style={{ padding: '6px 12px', border: '1px solid var(--border)', borderRadius: 9999, background: 'transparent', color: 'var(--text)', cursor: 'pointer', fontSize: 13 }}
          >Edit</button>
          <button
            type="button"
            onClick={() => setDeleteId(r.faculty_id)}
            className="btn-ghost"
            style={{ padding: '6px 12px', border: '1px solid var(--border)', borderRadius: 9999, background: 'transparent', color: 'var(--coral)', cursor: 'pointer', fontSize: 13 }}
          >Delete</button>
        </div>
      ),
    },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AdminPage
      title="Faculties"
      description="Manage the list of faculties students can belong to."
      actions={
        <button
          type="button"
          onClick={() => { setCreating(true); setName(''); }}
          className="btn-solid"
          style={{
            padding: '8px 16px',
            borderRadius: 9999,
            background: 'var(--accent)',
            color: '#fff',
            border: 'none',
            fontWeight: 600,
            cursor: 'pointer',
          }}
        >
          + New Faculty
        </button>
      }
    >
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : (
        <>
          <DataTable<Faculty> columns={columns} rows={q.data?.data ?? []} rowKey={(r) => r.faculty_id} />
          <Pager page={page} totalPages={totalPages} onChange={setPage} />
        </>
      )}

      <CrudFormModal
        open={creating}
        title="New Faculty"
        busy={createMut.isPending}
        onClose={() => setCreating(false)}
        onSubmit={() => createMut.mutate()}
      >
        <Field label="Faculty name" value={name} onChange={setName} required />
      </CrudFormModal>

      <CrudFormModal
        open={editing !== null}
        title="Edit Faculty"
        busy={updateMut.isPending}
        onClose={() => setEditing(null)}
        onSubmit={() => updateMut.mutate()}
      >
        <Field label="Faculty name" value={name} onChange={setName} required />
      </CrudFormModal>

      <ConfirmModal
        open={deleteId !== null}
        title="Delete this faculty?"
        description="Users assigned to this faculty will lose their assignment."
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setDeleteId(null)}
      />
    </AdminPage>
  );
}

function Pager({ page, totalPages, onChange }: { page: number; totalPages: number; onChange: (p: number) => void }) {
  return (
    <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Page {page} of {totalPages}</span>
      <div style={{ display: 'flex', gap: 8 }}>
        <button
          type="button"
          disabled={page <= 1}
          onClick={() => onChange(Math.max(1, page - 1))}
          className="btn-ghost"
          style={{ padding: '6px 14px', borderRadius: 9999, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text)', cursor: page <= 1 ? 'not-allowed' : 'pointer', opacity: page <= 1 ? 0.4 : 1 }}
        >Prev</button>
        <button
          type="button"
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
          className="btn-ghost"
          style={{ padding: '6px 14px', borderRadius: 9999, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text)', cursor: page >= totalPages ? 'not-allowed' : 'pointer', opacity: page >= totalPages ? 0.4 : 1 }}
        >Next</button>
      </div>
    </div>
  );
}
