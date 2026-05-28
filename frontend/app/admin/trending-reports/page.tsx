'use client';

import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, CartesianGrid, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
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

  const statsColumns: Column<TrendReasonStats>[] = [
    { key: 'label', header: 'Reason', render: (r) => <span style={{ fontWeight: 600 }}>{r.label}</span> },
    { key: 'code', header: 'Code', render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{r.code}</code> },
    { key: 'count', header: 'Count', render: (r) => <span style={{ fontWeight: 700 }}>{r.count}</span> },
  ];

  const reasonColumns: Column<TrendReason>[] = [
    { key: 'label', header: 'Label', render: (r) => <span style={{ fontWeight: 600 }}>{r.label}</span> },
    { key: 'code',  header: 'Code',  render: (r) => <code style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>{r.code}</code> },
  ];

  return (
    <AdminPage
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
          <div style={{ height: 260, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
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
        <DataTable<TrendReasonStats> columns={statsColumns} rows={data} rowKey={(r) => r.code} empty="No trend report stats" />
      </div>

      <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 10, color: 'var(--text)' }}>Reason Registry (read-only)</h3>
      <p style={{ fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 10 }}>
        These reasons are seeded in the backend and not editable from the UI.
      </p>
      <DataTable<TrendReason> columns={reasonColumns} rows={reasons} rowKey={(r) => r.code} empty="No reasons registered" />
    </AdminPage>
  );
}
