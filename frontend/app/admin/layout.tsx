'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ReactNode } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { AdminSubnav } from '@/components/admin/AdminSubnav';

export default function AdminLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const initialized = useAuthStore((s) => s.initialized);
  const isAdmin = user?.roleName === 'admin';

  useEffect(() => {
    if (initialized && !isAdmin) {
      router.replace('/');
    }
  }, [initialized, isAdmin, router]);

  if (!initialized) {
    return (
      <div style={{ padding: 40, color: 'var(--text-tertiary)' }}>Loading…</div>
    );
  }

  if (!isAdmin) {
    return (
      <div style={{ padding: 40, color: 'var(--coral)' }}>
        Forbidden. Admin access required.
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AdminSubnav />
      {children}
    </div>
  );
}
