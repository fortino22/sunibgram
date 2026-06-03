'use client';

import { useMemo, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ShieldAlert } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
import { Pager } from '@/components/admin/Pager';
import {
  listViolations,
  createViolation,
  updateViolation,
  deleteViolation,
} from '@/lib/services/admin/violations';
import type { Violation } from '@/lib/services/admin/types';

export default function AdminViolationsPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [editing, setEditing] = useState<Violation | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [page, setPage] = useState(1);
  const PAGE_SIZE = 20;

  const q = useQuery({
    queryKey: ['admin', 'violations'],
    queryFn: () => listViolations(token!),
    enabled: !!token,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: ['admin', 'violations'] });

  const createMut = useMutation({
    mutationFn: () => createViolation(token!, { name, description }),
    onSuccess: () => { invalidate(); setCreating(false); setName(''); setDescription(''); },
  });
  const updateMut = useMutation({
    mutationFn: () => updateViolation(token!, editing!.id, { name, description }),
    onSuccess: () => { invalidate(); setEditing(null); },
  });
  const deleteMut = useMutation({
    mutationFn: () => deleteViolation(token!, deleteId!),
    onSuccess: () => { invalidate(); setDeleteId(null); },
  });

  const rows = useMemo(() => q.data ?? [], [q.data]);
  const total = rows.length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const pageRows = useMemo(
    () => rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE),
    [rows, page],
  );

  const columns: Column<Violation>[] = [
    { key: 'name', header: 'Name', sortValue: (r) => r.name?.toLowerCase(), render: (r) => <span style={{ fontWeight: 600 }}>{r.name}</span> },
    { key: 'desc', header: 'Description', render: (r) => <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>{r.description || '—'}</span> },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (r) => (
        <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={() => { setEditing(r); setName(r.name); setDescription(r.description); }}
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
      icon={ShieldAlert}
      title="Report Types"
      description="Manage the violation types users can choose when filing a report."
      actions={
        <button
          type="button"
          onClick={() => { setCreating(true); setName(''); setDescription(''); }}
          className="admin-btn admin-btn-primary"
        >
          + New Type
        </button>
      }
    >
      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Violation>
            columns={columns}
            rows={pageRows}
            rowKey={(r) => r.id}
            loading={q.isLoading}
            empty="No report types yet"
            emptyIcon={<ShieldAlert size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
          />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}

      <CrudFormModal
        open={creating}
        title="New Report Type"
        busy={createMut.isPending}
        onClose={() => setCreating(false)}
        onSubmit={() => createMut.mutate()}
      >
        <Field label="Name" value={name} onChange={setName} required />
        <Field label="Description" value={description} onChange={setDescription} textarea />
      </CrudFormModal>

      <CrudFormModal
        open={editing !== null}
        title="Edit Report Type"
        busy={updateMut.isPending}
        onClose={() => setEditing(null)}
        onSubmit={() => updateMut.mutate()}
      >
        <Field label="Name" value={name} onChange={setName} required />
        <Field label="Description" value={description} onChange={setDescription} textarea />
      </CrudFormModal>

      <ConfirmModal
        open={deleteId !== null}
        title="Delete this report type?"
        description="Existing reports using this type will be affected."
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setDeleteId(null)}
      />
    </AdminPage>
  );
}
