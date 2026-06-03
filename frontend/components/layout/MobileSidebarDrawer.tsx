'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { LogOut, X } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { Icon } from '@/components/ui/Icon';
import { useAuthStore } from '@/stores/authStore';
import { logout as logoutRequest } from '@/lib/services/auth';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useMobileDrawerStore } from '@/stores/mobileDrawerStore';

interface DrawerItem {
  id: string;
  href: string;
  icon: string;
  label: string;
}

const ITEMS: DrawerItem[] = [
  { id: 'profile',     href: '/profile',     icon: 'user',     label: 'Profile' },
  { id: 'communities', href: '/communities', icon: 'users',    label: 'Communities' },
  { id: 'events',      href: '/events',      icon: 'calendar', label: 'Events' },
  { id: 'leaderboard', href: '/leaderboard', icon: 'trophy',   label: 'Leaderboard' },
  { id: 'bookmarks',   href: '/bookmarks',   icon: 'bookmark', label: 'Bookmarks' },
  { id: 'feedback',    href: '/feedback',    icon: 'feedback', label: 'Feedback' },
  { id: 'settings',    href: '/settings',    icon: 'settings', label: 'Settings' },
];

const ADMIN_ITEM: DrawerItem = { id: 'admin', href: '/admin', icon: 'shield', label: 'Admin' };

export function MobileSidebarDrawer() {
  const router = useRouter();
  const pathname = usePathname();
  const open = useMobileDrawerStore((s) => s.open);
  const closeDrawer = useMobileDrawerStore((s) => s.closeDrawer);
  const { user, logout } = useAuthStore();
  const { data: profile } = useMyProfile();
  const queryClient = useQueryClient();

  // Close on route change.
  useEffect(() => {
    closeDrawer();
  }, [pathname, closeDrawer]);

  // Esc to close.
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeDrawer();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, closeDrawer]);

  // Lock body scroll while open.
  useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = prev;
    };
  }, [open]);

  const items = user?.roleName === 'admin' ? [...ITEMS, ADMIN_ITEM] : ITEMS;
  const avatarSrc = profile?.avatar ?? null;

  const handleLogout = async () => {
    const token = useAuthStore.getState().token;
    if (token) {
      try {
        await logoutRequest(token);
      } catch {
        // best-effort: tear down the client even if the backend call fails
      }
    }
    await fetch('/auth/clear-session', { method: 'POST' });
    logout();
    queryClient.clear();
    closeDrawer();
    router.replace('/login');
  };

  const isActive = (href: string) => pathname === href || pathname.startsWith(href + '/');

  return (
    <div
      className={`mobile-drawer-root${open ? ' is-open' : ''}`}
      aria-hidden={!open}
    >
      <div
        className="mobile-drawer-backdrop"
        onClick={closeDrawer}
        role="presentation"
      />

      <aside
        className="mobile-drawer-panel"
        role="dialog"
        aria-modal="true"
        aria-label="Account menu"
      >
        <div className="mobile-drawer-header">
          <button
            type="button"
            onClick={closeDrawer}
            aria-label="Close menu"
            className="mobile-drawer-close"
          >
            <X size={20} strokeWidth={2} />
          </button>

          <Link href="/profile" className="mobile-drawer-profile">
            <Avatar
              username={user?.username ?? 'Account'}
              avatarUrl={avatarSrc}
              size={56}
            />
            <div style={{ marginTop: 12 }}>
              <div className="mobile-drawer-name">
                <span>{profile?.name ?? user?.username ?? '—'}</span>
                {profile?.is_premium && <VerifiedBadge size={14} />}
              </div>
              <div className="mobile-drawer-handle">
                @{profile?.username ?? user?.username ?? ''}
              </div>
            </div>
          </Link>

          {profile && (
            <div className="mobile-drawer-counts">
              <Link href={`/profile/${profile.id}?tab=following`} className="mobile-drawer-count">
                <strong>{profile.following_count ?? 0}</strong>
                <span>Following</span>
              </Link>
              <Link href={`/profile/${profile.id}?tab=followers`} className="mobile-drawer-count">
                <strong>{profile.followers_count ?? 0}</strong>
                <span>Followers</span>
              </Link>
            </div>
          )}
        </div>

        <nav className="mobile-drawer-nav" aria-label="Account">
          {items.map((item) => {
            const active = isActive(item.href);
            return (
              <Link
                key={item.id}
                href={item.href}
                className={`mobile-drawer-item${active ? ' is-active' : ''}`}
                aria-current={active ? 'page' : undefined}
              >
                <Icon
                  name={item.icon}
                  size={24}
                  color="var(--text)"
                  strokeWidth={active ? 2.2 : 1.8}
                />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        <div className="mobile-drawer-footer">
          <button
            type="button"
            onClick={handleLogout}
            className="mobile-drawer-logout"
          >
            <LogOut size={18} />
            <span>Sign out</span>
          </button>
        </div>
      </aside>
    </div>
  );
}
