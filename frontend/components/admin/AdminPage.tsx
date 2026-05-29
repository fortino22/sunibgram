'use client';

import { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';

interface AdminPageProps {
  title: string;
  description?: string;
  icon?: LucideIcon;
  actions?: ReactNode;
  children: ReactNode;
}

export function AdminPage({ title, description, icon: Icon, actions, children }: AdminPageProps) {
  return (
    <div style={{ padding: '24px 32px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <header
        className="admin-page-header"
        style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          {Icon && (
            <span
              aria-hidden
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: 40,
                height: 40,
                borderRadius: 12,
                background: 'var(--bg-subtle)',
                color: 'var(--text-secondary)',
                border: '1px solid var(--border)',
                flexShrink: 0,
              }}
            >
              <Icon size={20} strokeWidth={2} />
            </span>
          )}
          <div>
            <h1 style={{ fontSize: 26, fontWeight: 700, letterSpacing: '-0.02em', color: 'var(--text)' }}>
              {title}
            </h1>
            {description && (
              <p style={{ marginTop: 6, color: 'var(--text-secondary)', fontSize: 14 }}>{description}</p>
            )}
          </div>
        </div>
        {actions && <div style={{ display: 'flex', gap: 8 }}>{actions}</div>}
      </header>
      <div>{children}</div>
    </div>
  );
}
