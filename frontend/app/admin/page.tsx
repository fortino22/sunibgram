'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ArrowUpRight,
  CalendarDays,
  Flag,
  GraduationCap,
  LayoutDashboard,
  Receipt,
  ShieldAlert,
  ShieldCheck,
  Tags,
  Users as UsersIcon,
  type LucideIcon,
} from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import { listUsers } from '@/lib/services/admin/users';
import { listAdminEvents } from '@/lib/services/admin/events';
import { listReports } from '@/lib/services/admin/reports';
import { listOrders } from '@/lib/services/admin/orders';
import { listFacultiesPaginated } from '@/lib/services/admin/faculties';
import { listCategories } from '@/lib/services/admin/categories';
import { listViolations } from '@/lib/services/admin/violations';
import { listRoles } from '@/lib/services/admin/roles';
import { clearCache } from '@/lib/services/admin/cache';

interface StatCardProps {
  label: string;
  value: string | number;
  hint?: string;
  icon: LucideIcon;
  href: string;
  loading?: boolean;
}

function StatCard({ label, value, hint, icon: Icon, href, loading }: StatCardProps) {
  return (
    <Link
      href={href}
      className="admin-stat-card admin-stat-card-link"
      style={{
        display: 'block',
        padding: 18,
        borderRadius: 16,
        border: '1px solid var(--border)',
        background: 'var(--bg-card)',
        color: 'var(--text)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          {label}
        </div>
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
      </div>
      <div style={{ marginTop: 8, minHeight: 34, display: 'flex', alignItems: 'center' }}>
        {loading ? (
          <div className="admin-skeleton" style={{ height: 26, width: 56, borderRadius: 8 }} />
        ) : (
          <span style={{ fontSize: 28, fontWeight: 700, letterSpacing: '-0.02em' }}>{value}</span>
        )}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, marginTop: 4 }}>
        <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{hint ?? ' '}</span>
        <ArrowUpRight className="admin-stat-arrow" size={15} strokeWidth={2.2} style={{ color: 'var(--text-tertiary)' }} />
      </div>
    </Link>
  );
}

export default function AdminOverviewPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [confirmOpen, setConfirmOpen] = useState(false);

  const usersQ = useQuery({ queryKey: ['admin', 'users', 'overview'], queryFn: () => listUsers(token!, 1, 1), enabled: !!token });
  const pendingEventsQ = useQuery({ queryKey: ['admin', 'events', 'pending', 'overview'], queryFn: () => listAdminEvents(token!, { status: 'pending', page: 1, limit: 1 }), enabled: !!token });
  const reportsQ = useQuery({ queryKey: ['admin', 'reports', 'overview'], queryFn: () => listReports(token!), enabled: !!token });
  const ordersQ = useQuery({ queryKey: ['admin', 'orders', 'overview'], queryFn: () => listOrders(token!, 1, 1), enabled: !!token });
  const facultiesQ = useQuery({ queryKey: ['admin', 'faculties', 'overview'], queryFn: () => listFacultiesPaginated(token!, 1, 1), enabled: !!token });
  const categoriesQ = useQuery({ queryKey: ['admin', 'categories', 'overview'], queryFn: () => listCategories(token!), enabled: !!token });
  const violationsQ = useQuery({ queryKey: ['admin', 'violations', 'overview'], queryFn: () => listViolations(token!), enabled: !!token });
  const rolesQ = useQuery({ queryKey: ['admin', 'roles', 'overview'], queryFn: () => listRoles(token!), enabled: !!token });

  const openReports = (reportsQ.data ?? []).filter((r) => ['open', 'in_progress'].includes((r.status ?? '').toLowerCase())).length;

  const clearMut = useMutation({
    mutationFn: () => clearCache(token!),
    onSuccess: () => {
      qc.clear();
      setConfirmOpen(false);
    },
  });

  const cards: StatCardProps[] = [
    { label: 'Users', value: usersQ.data?.meta?.total ?? '—', icon: UsersIcon, href: '/admin/users', loading: usersQ.isLoading, hint: 'Registered accounts' },
    { label: 'Pending Events', value: pendingEventsQ.data?.meta?.total ?? '—', icon: CalendarDays, href: '/admin/events', loading: pendingEventsQ.isLoading, hint: 'Awaiting review' },
    { label: 'Reports', value: reportsQ.data?.length ?? '—', icon: Flag, href: '/admin/reports', loading: reportsQ.isLoading, hint: reportsQ.data ? `${openReports} open` : undefined },
    { label: 'Orders', value: ordersQ.data?.meta?.total ?? '—', icon: Receipt, href: '/admin/orders', loading: ordersQ.isLoading, hint: 'Premium transactions' },
    { label: 'Faculties', value: facultiesQ.data?.meta?.total ?? '—', icon: GraduationCap, href: '/admin/faculties', loading: facultiesQ.isLoading, hint: 'Configured' },
    { label: 'Categories', value: categoriesQ.data?.length ?? '—', icon: Tags, href: '/admin/categories', loading: categoriesQ.isLoading, hint: 'Community categories' },
    { label: 'Report Types', value: violationsQ.data?.length ?? '—', icon: ShieldAlert, href: '/admin/violations', loading: violationsQ.isLoading, hint: 'Violation rules' },
    { label: 'Roles', value: rolesQ.data?.length ?? '—', icon: ShieldCheck, href: '/admin/roles', loading: rolesQ.isLoading, hint: 'Access levels' },
  ];

  return (
    <AdminPage
      icon={LayoutDashboard}
      title="Admin Dashboard"
      description="Operational overview and platform controls."
      actions={
        <button type="button" onClick={() => setConfirmOpen(true)} className="admin-btn admin-btn-danger">
          Clear App Cache
        </button>
      }
    >
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(210px, 1fr))', gap: 12 }}>
        {cards.map((c) => (
          <StatCard key={c.label} {...c} />
        ))}
      </div>

      <div style={{ marginTop: 24, padding: 20, borderRadius: 16, border: '1px solid var(--border)', background: 'var(--bg-card)' }}>
        <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 8 }}>Welcome back, admin</h2>
        <p style={{ fontSize: 14, color: 'var(--text-secondary)', lineHeight: 1.6 }}>
          Every card above links straight to its tab. Use the navigation bar to manage users, roles, reports, events,
          faculties, report types, categories, orders and feedback. The Reports and Trending tabs include chart
          breakdowns.
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
