'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { CalendarDays } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { StatusBadge } from '@/components/admin/StatusBadge';
import { Pager } from '@/components/admin/Pager';
import { listAdminEvents } from '@/lib/services/admin/events';
import type { AdminEvent } from '@/lib/services/admin/types';

const STATUS_OPTIONS = ['', 'pending', 'approved', 'declined', 'revision'];

export default function AdminEventsPage() {
  const token = useAuthStore((s) => s.token);
  const [status, setStatus] = useState<string>('pending');
  const [name, setName] = useState<string>('');
  const [page, setPage] = useState(1);

  const q = useQuery({
    queryKey: ['admin', 'events', status, name, page],
    queryFn: () => listAdminEvents(token!, { status: status || undefined, name: name || undefined, page, limit: 20 }),
    enabled: !!token,
  });

  const columns: Column<AdminEvent>[] = [
    {
      key: 'event',
      header: 'Event',
      render: (e) => (
        <div>
          <Link href={`/admin/events/${e.id}`} style={{ fontWeight: 600, color: 'var(--text)', textDecoration: 'none' }}>
            {e.title}
          </Link>
          <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{e.location || '—'}</div>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      render: (e) => (
        <span
          style={{
            display: 'inline-block',
            padding: '2px 10px',
            borderRadius: 9999,
            background: e.type === 'official' ? 'var(--accent-soft)' : 'var(--bg-subtle)',
            color: e.type === 'official' ? 'var(--accent)' : 'var(--text-secondary)',
            fontSize: 12,
            fontWeight: 600,
            textTransform: 'capitalize',
          }}
        >
          {e.type}
        </span>
      ),
    },
    { key: 'start',  header: 'Start',  render: (e) => new Date(e.start_time).toLocaleString() },
    { key: 'end',    header: 'End',    render: (e) => new Date(e.end_time).toLocaleString() },
    { key: 'status', header: 'Status', render: (e) => <StatusBadge status={e.status} /> },
    {
      key: 'actions',
      header: '',
      render: (e) => (
        <Link
          href={`/admin/events/${e.id}`}
          className="admin-btn admin-btn-outline admin-btn-sm"
          style={{ textDecoration: 'none' }}
        >
          Review
        </Link>
      ),
    },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AdminPage
      icon={CalendarDays}
      title="Events"
      description="Approve official events, request revisions, or decline submissions."
      actions={
        <div style={{ display: 'flex', gap: 8 }}>
          <input
            value={name}
            onChange={(e) => { setName(e.target.value); setPage(1); }}
            placeholder="Search by title"
            className="admin-input admin-input-pill"
            style={{ minWidth: 220 }}
          />
          <select
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1); }}
            className="admin-select admin-select-pill"
          >
            {STATUS_OPTIONS.map((s) => (
              <option key={s} value={s}>{s ? s : 'All'}</option>
            ))}
          </select>
        </div>
      }
    >
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<AdminEvent> columns={columns} rows={q.data?.data ?? []} rowKey={(e) => e.id} empty="No events match" />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}
    </AdminPage>
  );
}
