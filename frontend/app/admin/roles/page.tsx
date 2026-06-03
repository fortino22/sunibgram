'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ShieldCheck } from 'lucide-react';
import { toast } from 'sonner';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { CrudFormModal, Field } from '@/components/admin/CrudFormModal';
import { listRoles, createRole, updateRole, deleteRole } from '@/lib/services/admin/roles';
import type { Role } from '@/lib/services/admin/types';

// Core roles the platform relies on — deletion is blocked to avoid orphaning
// every account assigned to them.
const PROTECTED_ROLES = new Set(['admin', 'user', 'lecturer']);

export default function AdminRolesPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [editing, setEditing] = useState<Role | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Role | null>(null);
  const [name, setName] = useState('');

  const q = useQuery({
    queryKey: ['admin', 'roles'],
    queryFn: () => listRoles(token!),
    enabled: !!token,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: ['admin', 'roles'] });

  const createMut = useMutation({
    mutationFn: () => createRole(token!, name.trim()),
    onSuccess: () => { invalidate(); setCreating(false); setName(''); toast.success('Role created'); },
    onError: (err: Error) => toast.error(err.message),
  });
  const updateMut = useMutation({
    mutationFn: () => updateRole(token!, editing!.role_id, name.trim()),
    onSuccess: () => { invalidate(); setEditing(null); toast.success('Role updated'); },
    onError: (err: Error) => toast.error(err.message),
  });
  const deleteMut = useMutation({
    mutationFn: () => deleteRole(token!, deleteTarget!.role_id),
    onSuccess: () => { invalidate(); setDeleteTarget(null); toast.success('Role deleted'); },
    onError: (err: Error) => toast.error(err.message),
  });

  const columns: Column<Role>[] = [
    {
      key: 'name',
      header: 'Role',
      sortValue: (r) => r.name,
      render: (r) => {
        const cls = r.name === 'admin' ? 'admin-chip admin-chip-accent' : r.name === 'lecturer' ? 'admin-chip admin-chip-green' : 'admin-chip';
        return (
          <span className={cls}>
            <ShieldCheck size={12} strokeWidth={2.2} />
            {r.name}
          </span>
        );
      },
    },
    { key: 'id', header: 'ID', sortValue: (r) => r.role_id, render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>#{r.role_id}</code> },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (r) => {
        const locked = PROTECTED_ROLES.has(r.name);
        return (
          <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
            <button
              type="button"
              onClick={() => { setEditing(r); setName(r.name); }}
              className="admin-btn admin-btn-outline admin-btn-sm"
            >
              Edit
            </button>
            <button
              type="button"
              onClick={() => setDeleteTarget(r)}
              disabled={locked}
              title={locked ? 'Core roles cannot be deleted' : undefined}
              className="admin-btn admin-btn-outline-danger admin-btn-sm"
            >
              Delete
            </button>
          </div>
        );
      },
    },
  ];

  return (
    <AdminPage
      icon={ShieldCheck}
      title="Roles"
      description="Access levels assigned to user accounts."
      actions={
        <button type="button" onClick={() => { setCreating(true); setName(''); }} className="admin-btn admin-btn-primary">
          + New Role
        </button>
      }
    >
      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <DataTable<Role>
          columns={columns}
          rows={q.data ?? []}
          rowKey={(r) => r.role_id}
          loading={q.isLoading}
          empty="No roles defined"
          emptyIcon={<ShieldCheck size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
        />
      )}

      <CrudFormModal
        open={creating}
        title="New Role"
        description="Roles control what a user can access. Use a short lowercase identifier."
        busy={createMut.isPending}
        onClose={() => setCreating(false)}
        onSubmit={() => createMut.mutate()}
      >
        <Field label="Name" value={name} onChange={setName} placeholder="e.g. moderator" required />
      </CrudFormModal>

      <CrudFormModal
        open={editing !== null}
        title="Edit Role"
        busy={updateMut.isPending}
        onClose={() => setEditing(null)}
        onSubmit={() => updateMut.mutate()}
      >
        <Field label="Name" value={name} onChange={setName} required />
      </CrudFormModal>

      <ConfirmModal
        open={deleteTarget !== null}
        title={`Delete the "${deleteTarget?.name}" role?`}
        description="Accounts assigned this role will need to be reassigned. This cannot be undone."
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setDeleteTarget(null)}
      />
    </AdminPage>
  );
}
