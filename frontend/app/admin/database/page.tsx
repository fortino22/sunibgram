'use client';

import { Suspense } from 'react';
import { Database } from 'lucide-react';
import { AdminPage } from '@/components/admin/AdminPage';
import { SchemaBrowser } from '@/components/admin/SchemaBrowser';

export default function AdminDatabasePage() {
  return (
    <AdminPage
      icon={Database}
      title="Database"
      description="Live database schema — tables, columns, and relationships."
    >
      <Suspense fallback={null}>
        <SchemaBrowser />
      </Suspense>
    </AdminPage>
  );
}
