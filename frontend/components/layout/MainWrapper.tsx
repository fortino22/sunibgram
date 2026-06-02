'use client';

import { usePathname } from 'next/navigation';
import { useState, type ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { RightPanel } from './RightPanel';
import { FloatingFoxButton } from './FloatingFoxButton';
import { ThreadSummaryPanel } from '@/components/thread/ThreadSummaryPanel';
import { AuthGuard } from '@/components/auth/AuthGuard';
import { NewPostModal } from '@/components/shared/NewPostModal';
import { ReplyModal } from '@/components/shared/ReplyModal';
import { OnboardingModal } from '@/components/onboarding/OnboardingModal';
import { AboutIntroModal } from '@/components/onboarding/AboutIntroModal';
import { useNotificationSocket } from '@/hooks/useNotificationSocket';
import { useCompleteOnboarding } from '@/hooks/useCompleteOnboarding';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useAuthStore } from '@/stores/authStore';

const AUTH_PATHS = ['/login', '/signup', '/forgot-password', '/about'];
const NO_RIGHT_PANEL_PATHS = ['/settings', '/messages'];

function NotificationSocketMount() {
  const token = useAuthStore((s) => s.token);
  useNotificationSocket({ accessToken: token });
  return null;
}

// First-time gate, server-driven by `onboarding_completed`. Two steps while the
// flag is false: (1) the hard follow gate, then (2) a closable About intro. The
// flag flips true only after About is dismissed — so the intro is the final
// onboarding step and never reshows once complete (no localStorage). Mounted
// only inside AuthGuard, so the profile query is always authenticated here.
function OnboardingGate() {
  const { data: profile } = useMyProfile();
  const complete = useCompleteOnboarding();
  // In-memory only: the follow step is "done" for this session once the user
  // clears the gate. On reload the follow gate reappears but is instantly
  // satisfied (follow relationships persist), so no progress is lost.
  const [followStepDone, setFollowStepDone] = useState(false);

  if (!profile || profile.onboarding_completed !== false) return null;

  if (!followStepDone) {
    return <OnboardingModal open onFollowStepDone={() => setFollowStepDone(true)} />;
  }

  // About is the final step — dismissing it completes onboarding server-side.
  return <AboutIntroModal open onClose={() => complete.mutate()} />;
}

interface MainWrapperProps {
  children: ReactNode;
}

export function MainWrapper({ children }: MainWrapperProps) {
  const pathname = usePathname();
  // Match exact path or a sub-path boundary so e.g. "/about" does not swallow
  // "/about-profile" (an in-app page that needs the full shell).
  const isAuthPath = AUTH_PATHS.some((p) => pathname === p || pathname.startsWith(p + '/'));
  const hideRightPanel = NO_RIGHT_PANEL_PATHS.some((p) => pathname === p || pathname.startsWith(p + '/'));

  if (isAuthPath) {
    return <>{children}</>;
  }

  return (
    <AuthGuard>
      <NotificationSocketMount />
      <OnboardingGate />
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
      <FloatingFoxButton />
      <ThreadSummaryPanel />
    </AuthGuard>
  );
}
