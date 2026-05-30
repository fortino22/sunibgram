'use client';

import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { RightPanel } from './RightPanel';
import { AuthGuard } from '@/components/auth/AuthGuard';
import { NewPostModal } from '@/components/shared/NewPostModal';
import { ReplyModal } from '@/components/shared/ReplyModal';
import { useNotificationSocket } from '@/hooks/useNotificationSocket';
import { useAuthStore } from '@/stores/authStore';

const AUTH_PATHS = ['/login', '/signup', '/forgot-password', '/about'];
const NO_RIGHT_PANEL_PATHS = ['/settings', '/messages'];

function NotificationSocketMount() {
  const token = useAuthStore((s) => s.token);
  useNotificationSocket({ accessToken: token });
  return null;
}

interface MainWrapperProps {
  children: ReactNode;
}

export function MainWrapper({ children }: MainWrapperProps) {
  const pathname = usePathname();
  const isAuthPath = AUTH_PATHS.some((p) => pathname.startsWith(p));
  const hideRightPanel = NO_RIGHT_PANEL_PATHS.some((p) => pathname.startsWith(p));

  if (isAuthPath) {
    return <>{children}</>;
  }

  return (
    <AuthGuard>
      <NotificationSocketMount />
      <NewPostModal />
      <ReplyModal />
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
            maxWidth: 1280,
            minHeight: '100vh',
          }}
        >
          <Sidebar />
          <main
            style={{
              flex: 1,
              minWidth: 0,
              minHeight: '100vh',
              background: 'var(--bg)',
              borderLeft: '1px solid var(--border)',
              borderRight: '1px solid var(--border)',
            }}
          >
            {children}
          </main>
          {!hideRightPanel && <RightPanel />}
        </div>
      </div>
    </AuthGuard>
  );
}
