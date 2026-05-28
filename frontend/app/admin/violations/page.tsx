'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
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

  const columns: Column<Violation>[] = [
    { key: 'name', header: 'Name', render: (r) => <span style={{ fontWeight: 600 }}>{r.name}</span> },
    { key: 'desc', header: 'Description', render: (r) => <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>{r.description}</span> },
    {
      key: 'actions',
      header: '',
      render: (r) => (
        <div style={{ display: 'flex', gap: 6 }}>
          <button
            type="button"
            onClick={() => { setEditing(r); setName(r.name); setDescription(r.description); }}
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
      title="Report Types"
      description="Manage the violation types users can choose when filing a report."
      actions={
        <button
          type="button"
          onClick={() => { setCreating(true); setName(''); setDescription(''); }}
          className="btn-solid"
          style={{ padding: '8px 16px', borderRadius: 9999, background: 'var(--accent)', color: '#fff', border: 'none', fontWeight: 600, cursor: 'pointer' }}
        >
          + New Type
        </button>
      }
    >
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : (
        <DataTable<Violation> columns={columns} rows={q.data ?? []} rowKey={(r) => r.id} />
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
