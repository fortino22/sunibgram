'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Receipt } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { StatusBadge } from '@/components/admin/StatusBadge';
import { Pager } from '@/components/admin/Pager';
import { fmtCurrency, fmtDateTime } from '@/components/admin/format';
import { listOrders } from '@/lib/services/admin/orders';
import type { Order } from '@/lib/services/admin/types';

export default function AdminOrdersPage() {
  const token = useAuthStore((s) => s.token);
  const [page, setPage] = useState(1);

  const q = useQuery({
    queryKey: ['admin', 'orders', page],
    queryFn: () => listOrders(token!, page, 20),
    enabled: !!token,
  });

  const columns: Column<Order>[] = [
    { key: 'order_id', header: 'Order',   sortValue: (o) => o.order_id, render: (o) => <code style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{o.order_id}</code> },
    { key: 'user_id',  header: 'User',    sortValue: (o) => o.user_id, render: (o) => <span style={{ color: 'var(--text-secondary)' }}>#{o.user_id}</span> },
    { key: 'purpose',  header: 'Purpose', sortValue: (o) => o.purpose, render: (o) => o.purpose },
    { key: 'amount',   header: 'Amount',  align: 'right', sortValue: (o) => o.amount, render: (o) => <span style={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>{fmtCurrency(o.amount)}</span> },
    { key: 'payment',  header: 'Method',  render: (o) => o.payment_type ? <span className="admin-chip">{o.payment_type}</span> : <span style={{ color: 'var(--text-tertiary)' }}>—</span> },
    { key: 'status',   header: 'Status',  sortValue: (o) => o.status, render: (o) => <StatusBadge status={o.status} /> },
    { key: 'created',  header: 'Created', sortValue: (o) => (o.created_at ? new Date(o.created_at).getTime() : 0), render: (o) => <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>{fmtDateTime(o.created_at)}</span> },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AdminPage icon={Receipt} title="Orders" description="All premium subscription payments.">
      {q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Order>
            columns={columns}
            rows={q.data?.data ?? []}
            rowKey={(o) => o.id}
            loading={q.isLoading}
            empty="No orders yet"
            emptyIcon={<Receipt size={28} strokeWidth={1.6} style={{ color: 'var(--text-tertiary)' }} />}
          />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}
    </AdminPage>
  );
}
