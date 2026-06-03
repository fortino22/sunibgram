'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, CartesianGrid, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { TrendingUp } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { Pager } from '@/components/admin/Pager';
import { getTrendStats, listTrendReasons } from '@/lib/services/admin/trendReports';
import type { TrendReason, TrendReasonStats } from '@/lib/services/admin/types';

const PALETTE = ['#5272e8', '#e85772', '#72e857', '#e8a857', '#a857e8', '#57e8e8', '#e857a8'];

export default function AdminTrendingReportsPage() {
  const token = useAuthStore((s) => s.token);

  const statsQ = useQuery({
    queryKey: ['admin', 'trend-stats'],
    queryFn: () => getTrendStats(token!),
    enabled: !!token,
  });

  const reasonsQ = useQuery({
    queryKey: ['admin', 'trend-reasons'],
    queryFn: () => listTrendReasons(token!),
    enabled: !!token,
  });

  const data = statsQ.data ?? [];
  const reasons = reasonsQ.data ?? [];
  const PAGE_SIZE = 10;
  const [statsPage, setStatsPage] = useState(1);
  const [reasonsPage, setReasonsPage] = useState(1);
  const statsTotalPages = Math.max(1, Math.ceil(data.length / PAGE_SIZE));
  const reasonsTotalPages = Math.max(1, Math.ceil(reasons.length / PAGE_SIZE));
  const statsPageRows = data.slice((statsPage - 1) * PAGE_SIZE, statsPage * PAGE_SIZE);
  const reasonsPageRows = reasons.slice((reasonsPage - 1) * PAGE_SIZE, reasonsPage * PAGE_SIZE);

  const statsColumns: Column<TrendReasonStats>[] = [
    { key: 'label', header: 'Reason', sortValue: (r) => r.label?.toLowerCase(), render: (r) => <span style={{ fontWeight: 600 }}>{r.label}</span> },
    { key: 'code', header: 'Code', sortValue: (r) => r.code, render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{r.code}</code> },
    { key: 'count', header: 'Count', align: 'right', sortValue: (r) => r.count, render: (r) => <span style={{ fontWeight: 700 }}>{r.count}</span> },
  ];

  const reasonColumns: Column<TrendReason>[] = [
    { key: 'label', header: 'Label', sortValue: (r) => r.label?.toLowerCase(), render: (r) => <span style={{ fontWeight: 600 }}>{r.label}</span> },
    { key: 'code',  header: 'Code',  sortValue: (r) => r.code, render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{r.code}</code> },
  ];

  return (
    <AdminPage
      icon={TrendingUp}
      title="Trending Reports"
      description="Reports filed against trending hashtags, grouped by reason."
    >
      <div
        style={{
          padding: 18,
          borderRadius: 16,
          border: '1px solid var(--border)',
          background: 'var(--bg-card)',
          marginBottom: 20,
        }}
      >
        <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 12, color: 'var(--text-secondary)' }}>
          Reports by Reason
        </h3>
        {statsQ.isLoading ? (
          <div className="admin-skeleton" style={{ height: 320, borderRadius: 12 }} />
        ) : data.length === 0 ? (
          <div style={{ height: 260, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-tertiary)' }}>No data</div>
        ) : (
          <div style={{ width: '100%', height: 320 }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data} margin={{ top: 10, right: 10, left: -10, bottom: 0 }}>
                <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" vertical={false} />
                <XAxis dataKey="label" stroke="var(--text-tertiary)" tick={{ fontSize: 11 }} />
                <YAxis stroke="var(--text-tertiary)" tick={{ fontSize: 11 }} allowDecimals={false} />
                <Tooltip contentStyle={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 8 }} />
                <Bar dataKey="count" radius={[6, 6, 0, 0]}>
                  {data.map((_, i) => (
                    <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 10, color: 'var(--text)' }}>Stats per Reason</h3>
      <div style={{ marginBottom: 24 }}>
        <DataTable<TrendReasonStats> columns={statsColumns} rows={statsPageRows} rowKey={(r) => r.code} loading={statsQ.isLoading} empty="No trend report stats" />
        <Pager page={statsPage} totalPages={statsTotalPages} total={data.length} onChange={setStatsPage} />
      </div>

      <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 10, color: 'var(--text)' }}>Reason Registry (read-only)</h3>
      <p style={{ fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 10 }}>
        These reasons are seeded in the backend and not editable from the UI.
      </p>
      <DataTable<TrendReason> columns={reasonColumns} rows={reasonsPageRows} rowKey={(r) => r.code} loading={reasonsQ.isLoading} empty="No reasons registered" />
      <Pager page={reasonsPage} totalPages={reasonsTotalPages} total={reasons.length} onChange={setReasonsPage} />
    </AdminPage>
  );
}
