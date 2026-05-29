'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { MessageSquareText } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { Pager } from '@/components/admin/Pager';
import { listFeedback, updateFeedbackStatus } from '@/lib/services/admin/feedback';
import type { Feedback } from '@/lib/services/admin/types';

const STATUS_OPTIONS = ['open', 'in_progress', 'resolved', 'dismissed'];

export default function AdminFeedbackPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);

  const q = useQuery({
    queryKey: ['admin', 'feedback', statusFilter, page],
    queryFn: () => listFeedback(token!, { status: statusFilter || undefined, page, limit: 20 }),
    enabled: !!token,
  });

  const statusMut = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => updateFeedbackStatus(token!, id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin', 'feedback'] }),
  });

  const columns: Column<Feedback>[] = [
    { key: 'type', header: 'Type', render: (r) => r.type },
    { key: 'title', header: 'Title', render: (r) => <span style={{ fontWeight: 600 }}>{r.title}</span> },
    { key: 'author', header: 'Author', render: (r) => r.author ? `@${r.author.username}` : '—' },
    { key: 'body', header: 'Body', render: (r) => <span style={{ color: 'var(--text-secondary)', fontSize: 13, display: 'block', maxWidth: 360 }}>{r.body}</span> },
    {
      key: 'status',
      header: 'Status',
      render: (r) => (
        <select
          value={r.status}
          onChange={(e) => statusMut.mutate({ id: r.id, status: e.target.value })}
          className="admin-select admin-select-sm"
        >
          {STATUS_OPTIONS.map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
      ),
    },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AdminPage
      icon={MessageSquareText}
      title="Feedback"
      description="User-submitted feedback and bug reports."
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
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Feedback> columns={columns} rows={q.data?.data ?? []} rowKey={(r) => r.id} empty="No feedback" />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}
    </AdminPage>
  );
}
