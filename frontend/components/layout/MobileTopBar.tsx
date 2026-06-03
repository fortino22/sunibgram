'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { Logo } from '@/components/ui/Logo';
import { Avatar } from '@/components/shared/Avatar';
import { useAuthStore } from '@/stores/authStore';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useMobileDrawerStore } from '@/stores/mobileDrawerStore';

const ROOT_PATHS = new Set<string>([
  '/',
  '/explore',
  '/notifications',
  '/messages',
  '/profile',
  '/communities',
  '/events',
  '/leaderboard',
  '/summaries',
  '/bookmarks',
  '/settings',
  '/feedback',
  '/admin',
]);

const TITLE_BY_PREFIX: Array<{ match: (p: string) => boolean; title: string }> = [
  { match: (p) => p.startsWith('/threads/'),       title: 'Post' },
  { match: (p) => p.startsWith('/profile/'),       title: 'Profile' },
  { match: (p) => p.startsWith('/about-profile/'), title: 'Profile' },
  { match: (p) => p.startsWith('/communities/'),   title: 'Community' },
  { match: (p) => p.startsWith('/events/'),        title: 'Event' },
  { match: (p) => p.startsWith('/messages/'),      title: 'Chat' },
  { match: (p) => p.startsWith('/settings/'),      title: 'Settings' },
  { match: (p) => p.startsWith('/admin/'),         title: 'Admin' },
  { match: (p) => p.startsWith('/bookmarks/'),     title: 'Bookmarks' },
  { match: (p) => p.startsWith('/about/'),         title: 'About' },
];

function getTitle(pathname: string): string {
  for (const t of TITLE_BY_PREFIX) if (t.match(pathname)) return t.title;
  return '';
}

export function MobileTopBar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useAuthStore();
  const { data: profile } = useMyProfile();
  const openDrawer = useMobileDrawerStore((s) => s.openDrawer);

  const isRoot = ROOT_PATHS.has(pathname);

  if (!isRoot) {
    return (
      <header className="mobile-top-bar mobile-top-bar-detail" role="banner">
        <button
          type="button"
          onClick={() => router.back()}
          aria-label="Back"
          className="mobile-top-bar-back"
        >
          <ArrowLeft size={22} strokeWidth={2} />
        </button>
        <span className="mobile-top-bar-title">{getTitle(pathname)}</span>
        <span className="mobile-top-bar-spacer" aria-hidden />
      </header>
    );
  }

  const avatarSrc = profile?.avatar ?? null;
  const username = user?.username ?? 'Account';
  const isPremium = profile?.is_premium ?? false;

  return (
    <header className="mobile-top-bar mobile-top-bar-root" role="banner">
      <button
        type="button"
        onClick={openDrawer}
        aria-label="Open menu"
        className="mobile-top-bar-avatar-btn"
      >
        <Avatar username={username} avatarUrl={avatarSrc} size={32} />
      </button>

      <Link href="/about" aria-label="About Sunibgram" className="mobile-top-bar-logo">
        <Logo size={28} style={{ color: 'var(--text)' }} />
      </Link>

      <Link
        href="/settings/subscription"
        className="mobile-top-bar-premium"
        aria-label={isPremium ? 'Premium' : 'Upgrade to Premium'}
      >
        {isPremium ? 'Premium' : 'Upgrade'}
      </Link>
    </header>
  );
}
