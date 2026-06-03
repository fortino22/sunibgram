'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { CalendarDays, MapPin, Search } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { Pager } from '@/components/admin/Pager';
import { StatusBadge } from '@/components/admin/StatusBadge';
import { fmtDateTime } from '@/components/admin/format';
import { AudienceBadges } from '@/components/events/AudienceBadges';
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
      sortValue: (e) => e.title?.toLowerCase(),
      render: (e) => (
        <div style={{ minWidth: 0 }}>
          <Link href={`/admin/events/${e.id}`} style={{ fontWeight: 600, color: 'var(--text)', textDecoration: 'none' }}>
            {e.title}
          </Link>
          <div style={{ fontSize: 12, color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center', gap: 4, marginTop: 2 }}>
            <MapPin size={11} strokeWidth={2} />
            {e.location || 'No venue'}
          </div>
        </div>
      ),
    },
    {
      key: 'audience',
      header: 'Region',
      render: (e) =>
        e.audience && e.audience.length > 0 ? (
          <AudienceBadges audience={e.audience} max={2} />
        ) : (
          <span style={{ color: 'var(--text-tertiary)' }}>—</span>
        ),
    },
    {
      key: 'type',
      header: 'Type',
      sortValue: (e) => e.type,
      render: (e) => <span className="admin-pill-outline">{e.type}</span>,
    },
    {
      key: 'when',
      header: 'When',
      sortValue: (e) => (e.start_time ? new Date(e.start_time).getTime() : 0),
      render: (e) => (
        <div style={{ fontSize: 13 }}>
          <div style={{ color: 'var(--text)' }}>{fmtDateTime(e.start_time)}</div>
          <div style={{ color: 'var(--text-tertiary)', fontSize: 12 }}>→ {fmtDateTime(e.end_time)}</div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortValue: (e) => e.status,
      render: (e) => <StatusBadge status={e.status} />,
    },
    {
      key: 'actions',
      header: '',
      render: (e) => (
        <Link href={`/admin/events/${e.id}`} className="admin-btn admin-btn-outline admin-btn-sm" style={{ textDecoration: 'none' }}>
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
      description="Review event submissions and manage approvals."
    >
      <div className="admin-toolbar">
        <div className="admin-search">
          <Search size={16} strokeWidth={2} />
          <input
            value={name}
            onChange={(e) => { setName(e.target.value); setPage(1); }}
            placeholder="Search by title"
            className="admin-input admin-input-pill"
            aria-label="Search events by title"
          />
        </div>
        <div className="admin-segmented" role="tablist" aria-label="Filter by status">
          {STATUS_OPTIONS.map((s) => (
            <button
              key={s}
              type="button"
              role="tab"
              aria-selected={status === s}
              data-active={status === s}
              onClick={() => { setStatus(s); setPage(1); }}
              className="admin-segmented-btn"
            >
              {s || 'All'}
            </button>
          ))}
        </div>
      </div>

      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<AdminEvent>
            columns={columns}
            rows={q.data?.data ?? []}
            rowKey={(e) => e.id}
            loading={q.isLoading}
            empty="No events match"
            emptyIcon={<CalendarDays size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
          />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}
    </AdminPage>
  );
}
