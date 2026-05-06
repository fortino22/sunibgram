'use client';

import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { RightPanel } from './RightPanel';
import { AuthGuard } from '@/components/auth/AuthGuard';

const AUTH_PATHS = ['/login', '/signup'];

interface MainWrapperProps {
  children: ReactNode;
}

export function MainWrapper({ children }: MainWrapperProps) {
  const pathname = usePathname();
  const isAuthPath = AUTH_PATHS.some((p) => pathname.startsWith(p));

  if (isAuthPath) {
    return <>{children}</>;
  }

  return (
    <AuthGuard>
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          minHeight: '100vh',
          background: 'var(--bg)',
        }}
      >
        <div
          style={{
            display: 'flex',
            width: '100%',
            maxWidth: 1200,
            minHeight: '100vh',
          }}
        >
          <Sidebar />
          <main
            style={{
              flex: 1,
              minWidth: 0,
              minHeight: '100vh',
              background: 'var(--bg-card)',
              borderLeft: '1px solid var(--border)',
              borderRight: '1px solid var(--border)',
            }}
          >
            {children}
          </main>
          <RightPanel />
        </div>
      </div>
    </AuthGuard>
  );
}
