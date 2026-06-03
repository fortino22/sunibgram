'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { GraduationCap } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
import { Pager } from '@/components/admin/Pager';
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
    { key: 'id',    header: 'ID',    width: 80, sortValue: (r) => r.faculty_id, render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>#{r.faculty_id}</code> },
    { key: 'name',  header: 'Name',  sortValue: (r) => r.faculty_name?.toLowerCase(), render: (r) => <span style={{ fontWeight: 600 }}>{r.faculty_name}</span> },
    { key: 'users', header: 'Users', align: 'right', sortValue: (r) => r.total_users ?? 0, render: (r) => (r.total_users ?? 0).toLocaleString() },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (r) => (
        <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={() => { setEditing(r); setName(r.faculty_name); }}
            className="admin-btn admin-btn-outline admin-btn-sm"
          >Edit</button>
          <button
            type="button"
            onClick={() => setDeleteId(r.faculty_id)}
            className="admin-btn admin-btn-outline-danger admin-btn-sm"
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
      icon={GraduationCap}
      title="Faculties"
      description="Manage the list of faculties students can belong to."
      actions={
        <button
          type="button"
          onClick={() => { setCreating(true); setName(''); }}
          className="admin-btn admin-btn-primary"
        >
          + New Faculty
        </button>
      }
    >
      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Faculty>
            columns={columns}
            rows={q.data?.data ?? []}
            rowKey={(r) => r.faculty_id}
            loading={q.isLoading}
            empty="No faculties yet"
            emptyIcon={<GraduationCap size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
          />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
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

