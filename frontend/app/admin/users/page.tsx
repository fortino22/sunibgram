'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Users } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { DetailModal, DetailRow } from '@/components/admin/DetailModal';
import { Pager } from '@/components/admin/Pager';
import { Avatar } from '@/components/shared/Avatar';
import { listUsers, getUser } from '@/lib/services/admin/users';
import type { AdminUser } from '@/lib/services/admin/types';

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
          className="admin-btn admin-btn-outline admin-btn-sm"
        >
          View
        </button>
      ),
    },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const detail = detailQ.data;

  return (
    <AdminPage
      icon={Users}
      title="Users"
      description="All registered users in Sunibgram."
      actions={
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search name / username / email"
          className="admin-input admin-input-pill"
          style={{ minWidth: 280 }}
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
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}

      <DetailModal
        open={selectedId !== null}
        onClose={() => setSelectedId(null)}
        hero={
          detail && (
            <div className="admin-detail-hero">
              <Avatar username={detail.username} avatarUrl={detail.avatar_url ?? null} size={72} />
              <div style={{ minWidth: 0, flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                  <span style={{ fontWeight: 700, fontSize: 20, color: 'var(--text)' }}>{detail.name}</span>
                  {detail.is_premium && <span className="admin-pill-outline">Premium</span>}
                </div>
                <div style={{ fontSize: 14, color: 'var(--text-tertiary)', marginTop: 2 }}>@{detail.username}</div>
                {detail.bio && (
                  <p style={{ marginTop: 8, fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5 }}>{detail.bio}</p>
                )}
              </div>
            </div>
          )
        }
      >
        {detailQ.isLoading && !detail && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
        )}
        {detail && (
          <>
            <DetailRow label="Email">{detail.email}</DetailRow>
            <DetailRow label="Role">{detail.role_name}</DetailRow>
            <DetailRow label="Faculty">{detail.faculty?.faculty_name ?? '—'}</DetailRow>
            <DetailRow label="Premium">{detail.is_premium ? 'Yes' : 'No'}</DetailRow>
            <DetailRow label="Followers">{(detail.follower_count ?? 0).toLocaleString()}</DetailRow>
            <DetailRow label="Following">{(detail.following_count ?? 0).toLocaleString()}</DetailRow>
            <DetailRow label="Joined">
              {detail.created_at ? new Date(detail.created_at).toLocaleString() : '—'}
            </DetailRow>
          </>
        )}
      </DetailModal>
    </AdminPage>
  );
}
