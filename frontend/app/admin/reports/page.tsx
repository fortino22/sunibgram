'use client';

import { useMemo, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { BarChart, Bar, CartesianGrid, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { Flag } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { StatusBadge } from '@/components/admin/StatusBadge';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { Pager } from '@/components/admin/Pager';
import { listReports, updateReportStatus, deleteReport } from '@/lib/services/admin/reports';
import type { Report } from '@/lib/services/admin/types';

const STATUS_OPTIONS = ['pending', 'resolved', 'dismissed'];
const PALETTE = ['#5272e8', '#e85772', '#72e857', '#e8a857', '#a857e8', '#57e8e8', '#e857a8'];

export default function AdminReportsPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const PAGE_SIZE = 20;

  const q = useQuery({
    queryKey: ['admin', 'reports'],
    queryFn: () => listReports(token!),
    enabled: !!token,
  });

  const statusMut = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => updateReportStatus(token!, id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin', 'reports'] }),
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteReport(token!, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'reports'] });
      setDeleteId(null);
    },
  });

  const reports = q.data ?? [];

  const chartData = useMemo(() => {
    const counts = new Map<string, number>();
    reports.forEach((r) => {
      const name = r.violation?.name ?? 'Unknown';
      counts.set(name, (counts.get(name) ?? 0) + 1);
    });
    return Array.from(counts.entries())
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count);
  }, [reports]);

  const filtered = statusFilter ? reports.filter((r) => r.status === statusFilter) : reports;
  const total = filtered.length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const pageRows = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const columns: Column<Report>[] = [
    {
      key: 'type',
      header: 'Type',
      render: (r) => <span style={{ fontWeight: 600 }}>{r.violation?.name ?? '—'}</span>,
    },
    {
      key: 'target',
      header: 'Target',
      render: (r) =>
        r.reported_user
          ? `@${r.reported_user.username}`
          : r.reported_post_id
          ? `post #${r.reported_post_id}`
          : '—',
    },
    {
      key: 'reporter',
      header: 'Reporter',
      render: (r) => (r.reporter_user ? `@${r.reporter_user.username}` : '—'),
    },
    {
      key: 'explanation',
      header: 'Reason',
      render: (r) => (
        <span style={{ color: 'var(--text-secondary)', fontSize: 13, display: 'block', maxWidth: 320 }}>
          {r.explanation}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (r) => (
        <select
          value={r.status}
          onChange={(e) => statusMut.mutate({ id: r.id, status: e.target.value })}
          className="admin-select admin-select-sm"
        >
          {STATUS_OPTIONS.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (r) => (
        <button
          type="button"
          onClick={() => setDeleteId(r.id)}
          className="admin-btn admin-btn-outline-danger admin-btn-sm"
        >
          Delete
        </button>
      ),
    },
  ];

  return (
    <AdminPage
      icon={Flag}
      title="Reports"
      description="Triage and resolve user reports. Chart breaks down by report type."
      actions={
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="admin-select admin-select-pill"
        >
          <option value="">All statuses</option>
          {STATUS_OPTIONS.map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
      }
    >
      <div style={{ display: 'flex', gap: 16, marginBottom: 20, flexWrap: 'wrap' }}>
        <div
          style={{
            flex: 2,
            minWidth: 360,
            padding: 18,
            borderRadius: 16,
            border: '1px solid var(--border)',
            background: 'var(--bg-card)',
          }}
        >
          <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 12, color: 'var(--text-secondary)' }}>
            Reports by Type
          </h3>
          {chartData.length === 0 ? (
            <div style={{ height: 220, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-tertiary)' }}>
              No data
            </div>
          ) : (
            <div style={{ width: '100%', height: 260 }}>
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData} margin={{ top: 10, right: 10, left: -10, bottom: 0 }}>
                  <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="name" stroke="var(--text-tertiary)" tick={{ fontSize: 11 }} />
                  <YAxis stroke="var(--text-tertiary)" tick={{ fontSize: 11 }} allowDecimals={false} />
                  <Tooltip contentStyle={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 8 }} />
                  <Bar dataKey="count" radius={[6, 6, 0, 0]}>
                    {chartData.map((_, i) => (
                      <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
        <div style={{ flex: 1, minWidth: 200, padding: 18, borderRadius: 16, border: '1px solid var(--border)', background: 'var(--bg-card)' }}>
          <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 12, color: 'var(--text-secondary)' }}>Summary</h3>
          <SummaryRow label="Total" value={reports.length} />
          {STATUS_OPTIONS.map((s) => (
            <SummaryRow
              key={s}
              label={s}
              value={reports.filter((r) => r.status === s).length}
              badge={<StatusBadge status={s} />}
            />
          ))}
        </div>
      </div>

      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Report> columns={columns} rows={pageRows} rowKey={(r) => r.id} />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}

      <ConfirmModal
        open={deleteId !== null}
        title="Delete this report?"
        description="This permanently removes the report record."
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)}
        onClose={() => setDeleteId(null)}
      />
    </AdminPage>
  );
}

function SummaryRow({ label, value, badge }: { label: string; value: number; badge?: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '6px 0' }}>
      <span style={{ fontSize: 13, color: 'var(--text-secondary)', textTransform: 'capitalize' }}>
        {badge ?? label}
      </span>
      <span style={{ fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>{value}</span>
    </div>
  );
}
