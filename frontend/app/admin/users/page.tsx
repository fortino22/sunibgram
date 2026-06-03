'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { CalendarClock, GraduationCap, MapPin, ShieldCheck, Users as UsersIcon } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { DetailModal, DetailRow } from '@/components/admin/DetailModal';
import { Pager } from '@/components/admin/Pager';
import { fmtDate, fmtDateTime } from '@/components/admin/format';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { listUsers, getUser } from '@/lib/services/admin/users';
import { listRoles } from '@/lib/services/admin/roles';
import type { AdminUser } from '@/lib/services/admin/types';

function RoleChip({ role }: { role: string }) {
  const r = (role ?? '').toLowerCase();
  const cls = r === 'admin' ? 'admin-chip admin-chip-accent' : r === 'lecturer' ? 'admin-chip admin-chip-green' : 'admin-chip';
  return (
    <span className={cls}>
      <ShieldCheck size={12} strokeWidth={2.2} />
      {role || 'user'}
    </span>
  );
}

export default function AdminUsersPage() {
  const token = useAuthStore((s) => s.token);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [detailTab, setDetailTab] = useState<'account' | 'profile'>('account');

  const q = useQuery({
    queryKey: ['admin', 'users', page],
    queryFn: () => listUsers(token!, page, 20),
    enabled: !!token,
  });

  const rolesQ = useQuery({
    queryKey: ['admin', 'roles'],
    queryFn: () => listRoles(token!),
    enabled: !!token,
  });

  const detailQ = useQuery({
    queryKey: ['admin', 'user', selectedId],
    queryFn: () => getUser(token!, selectedId!),
    enabled: !!token && selectedId !== null,
  });

  const filtered = (q.data?.data ?? []).filter((u) => {
    const matchesSearch =
      search.trim() === '' ||
      `${u.name} ${u.username} ${u.email}`.toLowerCase().includes(search.toLowerCase());
    const matchesRole = roleFilter === '' || u.role_name === roleFilter;
    return matchesSearch && matchesRole;
  });

  const columns: Column<AdminUser>[] = [
    {
      key: 'user',
      header: 'User',
      sortValue: (u) => u.name?.toLowerCase(),
      render: (u) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Avatar username={u.username} avatarUrl={u.avatar_url ?? null} size={36} />
          <div style={{ minWidth: 0 }}>
            <div style={{ fontWeight: 600, display: 'flex', alignItems: 'center', gap: 5 }}>
              <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{u.name}</span>
              {u.is_premium && <VerifiedBadge size={15} />}
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{u.username}</div>
          </div>
        </div>
      ),
    },
    { key: 'email', header: 'Email', sortValue: (u) => u.email?.toLowerCase(), render: (u) => u.email },
    { key: 'role', header: 'Role', sortValue: (u) => u.role_name, render: (u) => <RoleChip role={u.role_name} /> },
    {
      key: 'faculty',
      header: 'Faculty',
      sortValue: (u) => u.faculty?.faculty_name ?? '',
      render: (u) =>
        u.faculty?.faculty_name ? (
          <span style={{ color: 'var(--text-secondary)' }}>{u.faculty.faculty_name}</span>
        ) : (
          <span style={{ color: 'var(--text-tertiary)' }}>—</span>
        ),
    },
    {
      key: 'joined',
      header: 'Joined',
      sortValue: (u) => (u.created_at ? new Date(u.created_at).getTime() : 0),
      render: (u) => <span style={{ color: 'var(--text-secondary)' }}>{fmtDate(u.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      render: (u) => (
        <button type="button" onClick={() => setSelectedId(u.user_id)} className="admin-btn admin-btn-outline admin-btn-sm">
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
      icon={UsersIcon}
      title="Users"
      description="All registered users in Sunibgram."
      actions={
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search name / username / email"
            className="admin-input admin-input-pill"
            style={{ minWidth: 240 }}
          />
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="admin-select admin-select-pill"
            aria-label="Filter by role"
          >
            <option value="">All roles</option>
            {(rolesQ.data ?? []).map((r) => (
              <option key={r.role_id} value={r.name}>
                {r.name}
              </option>
            ))}
          </select>
        </div>
      }
    >
      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<AdminUser>
            columns={columns}
            rows={filtered}
            rowKey={(u) => u.user_id}
            loading={q.isLoading}
            onRowClick={(u) => setSelectedId(u.user_id)}
            empty="No users match"
            emptyIcon={<UsersIcon size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
          />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}

      <DetailModal
        open={selectedId !== null}
        onClose={() => { setSelectedId(null); setDetailTab('account'); }}
        title="User detail"
        hero={
          detail && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
              {detail.banner_url && (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={detail.banner_url}
                  alt=""
                  style={{ width: '100%', height: 96, objectFit: 'cover', borderRadius: '14px 14px 0 0', border: '1px solid var(--border)', borderBottom: 'none' }}
                />
              )}
              <div className="admin-detail-hero" style={detail.banner_url ? { borderRadius: '0 0 16px 16px' } : undefined}>
                <Avatar username={detail.username} avatarUrl={detail.avatar_url ?? null} size={72} />
                <div style={{ minWidth: 0, flex: 1 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, fontWeight: 700, fontSize: 20, color: 'var(--text)' }}>
                      {detail.name}
                      {detail.is_premium && <VerifiedBadge size={18} />}
                    </span>
                    <RoleChip role={detail.role_name} />
                  </div>
                  <div style={{ fontSize: 14, color: 'var(--text-tertiary)', marginTop: 2 }}>@{detail.username}</div>
                  {detail.bio && (
                    <p style={{ marginTop: 8, fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5 }}>{detail.bio}</p>
                  )}
                </div>
              </div>
              {detail && (
                <div className="admin-segmented" role="tablist" aria-label="Detail sections" style={{ marginTop: 14, alignSelf: 'flex-start' }}>
                  {(['account', 'profile'] as const).map((t) => (
                    <button
                      key={t}
                      type="button"
                      role="tab"
                      aria-selected={detailTab === t}
                      data-active={detailTab === t}
                      onClick={() => setDetailTab(t)}
                      className="admin-segmented-btn"
                    >
                      {t}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )
        }
      >
        {detailQ.isLoading && !detail && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
        )}
        {detail && detailTab === 'account' && (
          <>
            <DetailRow label="User ID">#{detail.user_id}</DetailRow>
            <DetailRow label="Email">{detail.email}</DetailRow>
            <DetailRow label="Role">
              <RoleChip role={detail.role_name} />
            </DetailRow>
            <DetailRow label="Premium">
              {detail.is_premium ? (
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                  <VerifiedBadge size={14} /> Yes
                </span>
              ) : (
                'No'
              )}
            </DetailRow>
            <DetailRow label="Onboarding">{detail.onboarding_completed ? 'Completed' : 'Incomplete'}</DetailRow>
            <DetailRow label="Joined">
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <CalendarClock size={14} strokeWidth={2} style={{ color: 'var(--text-tertiary)' }} />
                {fmtDateTime(detail.created_at)}
              </span>
            </DetailRow>
          </>
        )}
        {detail && detailTab === 'profile' && (
          <>
            <DetailRow label="Faculty">
              {detail.faculty?.faculty_name ? (
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                  <GraduationCap size={14} strokeWidth={2} style={{ color: 'var(--text-tertiary)' }} />
                  {detail.faculty.faculty_name}
                </span>
              ) : (
                '—'
              )}
            </DetailRow>
            <DetailRow label="Campus">
              {detail.faculty_location ? (
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                  <MapPin size={14} strokeWidth={2} style={{ color: 'var(--text-tertiary)' }} />
                  {detail.faculty_location.name}
                </span>
              ) : (
                '—'
              )}
            </DetailRow>
            <DetailRow label="Gender">{detail.gender ? detail.gender.replace(/_/g, ' ') : '—'}</DetailRow>
            <DetailRow label="Birthday">{detail.birthday_date ? fmtDate(detail.birthday_date) : '—'}</DetailRow>
            <DetailRow label="Followers">{(detail.follower_count ?? 0).toLocaleString()}</DetailRow>
            <DetailRow label="Following">{(detail.following_count ?? 0).toLocaleString()}</DetailRow>
          </>
        )}
      </DetailModal>
    </AdminPage>
  );
}
