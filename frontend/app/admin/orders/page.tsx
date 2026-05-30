'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Receipt } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { DataTable, Column } from '@/components/admin/DataTable';
import { StatusBadge } from '@/components/admin/StatusBadge';
import { Pager } from '@/components/admin/Pager';
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
    { key: 'order_id', header: 'Order',  render: (o) => <code style={{ fontSize: 12 }}>{o.order_id}</code> },
    { key: 'user_id',  header: 'User',   render: (o) => `#${o.user_id}` },
    { key: 'purpose',  header: 'Purpose', render: (o) => o.purpose },
    { key: 'amount',   header: 'Amount',  render: (o) => `Rp ${o.amount.toLocaleString()}` },
    { key: 'payment',  header: 'Method',  render: (o) => o.payment_type ?? '—' },
    { key: 'status',   header: 'Status',  render: (o) => <StatusBadge status={o.status} /> },
    { key: 'created',  header: 'Created', render: (o) => o.created_at ? new Date(o.created_at).toLocaleString() : '—' },
  ];

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AdminPage icon={Receipt} title="Orders" description="All premium subscription payments.">
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : (
        <>
          <DataTable<Order> columns={columns} rows={q.data?.data ?? []} rowKey={(o) => o.id} empty="No orders" />
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}
    </AdminPage>
  );
}
