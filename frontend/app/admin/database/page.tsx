'use client';

import { Database } from 'lucide-react';
import { AdminPage } from '@/components/admin/AdminPage';
import { ErdViewer } from '@/components/admin/ErdViewer';

export default function AdminDatabasePage() {
  return (
    <AdminPage
      icon={Database}
      title="Database"
      description="Interactive ERD of the current schema."
    >
      <ErdViewer />
    </AdminPage>
  );
}
