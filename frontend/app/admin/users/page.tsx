'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { Avatar } from '@/components/shared/Avatar';
import { listUsers, getUser } from '@/lib/services/admin/users';
import type { AdminUser } from '@/lib/services/admin/types';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';

export default function AdminUsersPage() {
  const token = useAuthStore((s) => s.token);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [selectedId, setSelectedId] = useState<number | null>(null);

  const q = useQuery({
    queryKey: ['admin', 'users', page],
    queryFn: () => listUsers(token!, page, 20),
    enabled: !!token,
  });

  const detailQ = useQuery({
    queryKey: ['admin', 'user', selectedId],
    queryFn: () => getUser(token!, selectedId!),
    enabled: !!token && selectedId !== null,
  });

  const filtered = (q.data?.data ?? []).filter((u) =>
    search.trim() === ''
      ? true
      : `${u.name} ${u.username} ${u.email}`.toLowerCase().includes(search.toLowerCase()),
  );

  const columns: Column<AdminUser>[] = [
    {
      key: 'user',
      header: 'User',
      render: (u) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Avatar username={u.username} avatarUrl={u.avatar_url ?? null} size={36} />
          <div>
            <div style={{ fontWeight: 600 }}>{u.name}</div>
            <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{u.username}</div>
          </div>
        </div>
      ),
    },
    { key: 'email',   header: 'Email',   render: (u) => u.email },
    { key: 'role',    header: 'Role',    render: (u) => u.role_name },
    { key: 'faculty', header: 'Faculty', render: (u) => u.faculty?.faculty_name ?? '—' },
    {
      key: 'actions',
      header: '',
      render: (u) => (
        <button
          type="button"
          onClick={() => setSelectedId(u.user_id)}
          className="btn-ghost"
          style={{
            padding: '6px 12px',
            border: '1px solid var(--border)',
            borderRadius: 9999,
            background: 'transparent',
            color: 'var(--text)',
            cursor: 'pointer',
            fontSize: 13,
          }}
        >
          View
        </button>
      ),
    },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AdminPage
      title="Users"
      description="All registered users in Sunibgram."
      actions={
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search name / username / email"
          style={{
            padding: '8px 14px',
            borderRadius: 9999,
            border: '1px solid var(--border)',
            background: 'var(--bg-card)',
            color: 'var(--text)',
            fontSize: 13,
            minWidth: 280,
            outline: 'none',
          }}
        />
      }
    >
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<AdminUser> columns={columns} rows={filtered} rowKey={(u) => u.user_id} />
          <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              {total > 0 ? `Page ${page} of ${totalPages} (${total} total)` : ''}
            </span>
            <div style={{ display: 'flex', gap: 8 }}>
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="btn-ghost"
                style={{
                  padding: '6px 14px',
                  borderRadius: 9999,
                  border: '1px solid var(--border)',
                  background: 'transparent',
                  color: 'var(--text)',
                  cursor: page <= 1 ? 'not-allowed' : 'pointer',
                  opacity: page <= 1 ? 0.4 : 1,
                }}
              >
                Prev
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="btn-ghost"
                style={{
                  padding: '6px 14px',
                  borderRadius: 9999,
                  border: '1px solid var(--border)',
                  background: 'transparent',
                  color: 'var(--text)',
                  cursor: page >= totalPages ? 'not-allowed' : 'pointer',
                  opacity: page >= totalPages ? 0.4 : 1,
                }}
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}

      <Dialog open={selectedId !== null} onOpenChange={(v) => { if (!v) setSelectedId(null); }}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>User Detail</DialogTitle>
          </DialogHeader>
          {detailQ.isLoading && <div>Loading…</div>}
          {detailQ.data && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div style={{ display: 'flex', gap: 14, alignItems: 'center' }}>
                <Avatar username={detailQ.data.username} avatarUrl={detailQ.data.avatar_url ?? null} size={64} />
                <div>
                  <div style={{ fontWeight: 700, fontSize: 18 }}>{detailQ.data.name}</div>
                  <div style={{ color: 'var(--text-tertiary)' }}>@{detailQ.data.username}</div>
                </div>
              </div>
              <Detail label="Email"     value={detailQ.data.email} />
              <Detail label="Role"      value={detailQ.data.role_name} />
              <Detail label="Faculty"   value={detailQ.data.faculty?.faculty_name ?? '—'} />
              <Detail label="Premium"   value={detailQ.data.is_premium ? 'Yes' : 'No'} />
              <Detail label="Followers" value={String(detailQ.data.follower_count ?? 0)} />
              <Detail label="Following" value={String(detailQ.data.following_count ?? 0)} />
              <Detail label="Joined"    value={detailQ.data.created_at ?? '—'} />
              {detailQ.data.bio && <Detail label="Bio" value={detailQ.data.bio} />}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </AdminPage>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border)', padding: '8px 0' }}>
      <span style={{ color: 'var(--text-tertiary)', fontSize: 13 }}>{label}</span>
      <span style={{ color: 'var(--text)', fontSize: 14, fontWeight: 500, maxWidth: '60%', textAlign: 'right' }}>{value}</span>
    </div>
  );
}
