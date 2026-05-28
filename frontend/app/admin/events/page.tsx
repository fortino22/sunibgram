'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { StatusBadge } from '@/components/admin/StatusBadge';
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
          className="btn-ghost"
          style={{ padding: '6px 12px', border: '1px solid var(--border)', borderRadius: 9999, background: 'transparent', color: 'var(--text)', fontSize: 13, textDecoration: 'none' }}
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
      title="Events"
      description="Approve official events, request revisions, or decline submissions."
      actions={
        <div style={{ display: 'flex', gap: 8 }}>
          <input
            value={name}
            onChange={(e) => { setName(e.target.value); setPage(1); }}
            placeholder="Search by title"
            style={{ padding: '8px 14px', borderRadius: 9999, border: '1px solid var(--border)', background: 'var(--bg-card)', color: 'var(--text)', fontSize: 13, minWidth: 220 }}
          />
          <select
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1); }}
            style={{ padding: '8px 14px', borderRadius: 9999, background: 'var(--bg-card)', border: '1px solid var(--border)', color: 'var(--text)', fontSize: 13 }}
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
          <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              {total > 0 ? `Page ${page} of ${totalPages} (${total} total)` : ''}
            </span>
            <div style={{ display: 'flex', gap: 8 }}>
              <button type="button" disabled={page <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))} className="btn-ghost" style={{ padding: '6px 14px', borderRadius: 9999, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text)', cursor: page <= 1 ? 'not-allowed' : 'pointer', opacity: page <= 1 ? 0.4 : 1 }}>Prev</button>
              <button type="button" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} className="btn-ghost" style={{ padding: '6px 14px', borderRadius: 9999, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text)', cursor: page >= totalPages ? 'not-allowed' : 'pointer', opacity: page >= totalPages ? 0.4 : 1 }}>Next</button>
            </div>
          </div>
        </>
      )}
    </AdminPage>
  );
}
