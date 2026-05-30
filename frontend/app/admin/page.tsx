'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { CalendarDays, Flag, LayoutDashboard, Receipt, Users as UsersIcon, type LucideIcon } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { listUsers } from '@/lib/services/admin/users';
import { listAdminEvents } from '@/lib/services/admin/events';
import { listReports } from '@/lib/services/admin/reports';
import { listOrders } from '@/lib/services/admin/orders';
import { clearCache } from '@/lib/services/admin/cache';

interface StatCardProps {
  label: string;
  value: string | number;
  hint?: string;
  icon?: LucideIcon;
}

function StatCard({ label, value, hint, icon: Icon }: StatCardProps) {
  return (
    <div
      className="admin-stat-card"
      style={{
        flex: 1,
        minWidth: 180,
        padding: 18,
        borderRadius: 16,
        border: '1px solid var(--border)',
        background: 'var(--bg-card)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          {label}
        </div>
        {Icon && (
          <span
            aria-hidden
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 30,
              height: 30,
              borderRadius: 9,
              background: 'var(--bg-subtle)',
              color: 'var(--text-secondary)',
              border: '1px solid var(--border)',
            }}
          >
            <Icon size={16} strokeWidth={2.2} />
          </span>
        )}
      </div>
      <div
        style={{
          marginTop: 8,
          fontSize: 28,
          fontWeight: 700,
          color: 'var(--text)',
          letterSpacing: '-0.02em',
          transition: 'color 200ms ease',
        }}
      >
        {value}
      </div>
      {hint && <div style={{ marginTop: 4, fontSize: 12, color: 'var(--text-secondary)' }}>{hint}</div>}
    </div>
  );
}

export default function AdminOverviewPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [confirmOpen, setConfirmOpen] = useState(false);

  const usersQ = useQuery({
    queryKey: ['admin', 'users', 'overview'],
    queryFn: () => listUsers(token!, 1, 1),
    enabled: !!token,
  });
  const pendingEventsQ = useQuery({
    queryKey: ['admin', 'events', 'pending', 'overview'],
    queryFn: () => listAdminEvents(token!, { status: 'pending', page: 1, limit: 1 }),
    enabled: !!token,
  });
  const reportsQ = useQuery({
    queryKey: ['admin', 'reports', 'overview'],
    queryFn: () => listReports(token!),
    enabled: !!token,
  });
  const ordersQ = useQuery({
    queryKey: ['admin', 'orders', 'overview'],
    queryFn: () => listOrders(token!, 1, 1),
    enabled: !!token,
  });

  const clearMut = useMutation({
    mutationFn: () => clearCache(token!),
    onSuccess: () => {
      qc.clear();
      setConfirmOpen(false);
    },
  });

  return (
    <AdminPage
      icon={LayoutDashboard}
      title="Admin Dashboard"
      description="Operational overview and platform controls."
      actions={
        <button
          type="button"
          onClick={() => setConfirmOpen(true)}
          className="admin-btn admin-btn-danger"
        >
          Clear App Cache
        </button>
      }
    >
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
        <StatCard label="Users" value={usersQ.data?.meta?.total ?? '—'} icon={UsersIcon} />
        <StatCard label="Pending Events" value={pendingEventsQ.data?.meta?.total ?? '—'} icon={CalendarDays} />
        <StatCard label="Total Reports" value={reportsQ.data?.length ?? '—'} icon={Flag} />
        <StatCard label="Orders" value={ordersQ.data?.meta?.total ?? '—'} icon={Receipt} />
      </div>

      <div style={{ marginTop: 24, padding: 18, borderRadius: 16, border: '1px solid var(--border)', background: 'var(--bg-card)' }}>
        <h2 style={{ fontSize: 16, fontWeight: 600, marginBottom: 8 }}>Welcome</h2>
        <p style={{ fontSize: 14, color: 'var(--text-secondary)', lineHeight: 1.6 }}>
          Use the tabs above to manage users, reports, events, faculties, report types, categories, orders and feedback.
          For chart breakdowns of reports and trending reports, open the respective tabs.
        </p>
      </div>

      <ConfirmModal
        open={confirmOpen}
        title="Clear all caches?"
        description="This will flush every Redis key used by the app. Reads will rebuild from the database on the next request. Use this if data looks stale."
        destructive
        confirmLabel="Clear Cache"
        busy={clearMut.isPending}
        onConfirm={() => clearMut.mutate()}
        onClose={() => setConfirmOpen(false)}
      />
    </AdminPage>
  );
}
