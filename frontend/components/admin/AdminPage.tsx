'use client';

import { ReactNode } from 'react';

interface AdminPageProps {
  title: string;
  description?: string;
  actions?: ReactNode;
  children: ReactNode;
}

export function AdminPage({ title, description, actions, children }: AdminPageProps) {
  return (
    <div style={{ padding: '24px 32px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <header style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <h1 style={{ fontSize: 26, fontWeight: 700, letterSpacing: '-0.02em', color: 'var(--text)' }}>
            {title}
          </h1>
          {description && (
            <p style={{ marginTop: 6, color: 'var(--text-secondary)', fontSize: 14 }}>{description}</p>
          )}
        </div>
        {actions && <div style={{ display: 'flex', gap: 8 }}>{actions}</div>}
      </header>
      <div>{children}</div>
    </div>
  );
}
