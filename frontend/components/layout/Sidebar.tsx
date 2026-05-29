'use client';

import Link from 'next/link';
import { Logo } from '@/components/ui/Logo';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
import { LogOut } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { Icon } from '@/components/ui/Icon';
import { useAuthStore } from '@/stores/authStore';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useComposeStore } from '@/stores/composeStore';
import { useNotificationStore } from '@/stores/notificationStore';

interface NavItem {
  id: string;
  href: string;
  icon: string;
  label: string;
  badge?: number;
}

const NAV_ITEMS: NavItem[] = [
  { id: 'home',          href: '/',              icon: 'home',     label: 'Home' },
  { id: 'explore',       href: '/explore',       icon: 'explore',  label: 'Explore' },
  { id: 'notifications', href: '/notifications', icon: 'bell',     label: 'Notifications'},
  { id: 'messages',      href: '/messages',      icon: 'messages', label: 'Messages'     },
  { id: 'communities',   href: '/communities',   icon: 'users',    label: 'Communities'},
  { id: 'events',        href: '/events',        icon: 'calendar', label: 'Events' },
  { id: 'leaderboard',   href: '/leaderboard',   icon: 'trophy',   label: 'Leaderboard' },
  { id: 'profile',       href: '/profile',       icon: 'user',     label: 'Profile' },
];

const MORE_NAV_ITEMS: NavItem[] = [
  { id: 'bookmarks', href: '/bookmarks', icon: 'bookmark', label: 'Bookmarks' },
  { id: 'feedback',  href: '/feedback',  icon: 'feedback', label: 'Feedback' },
  { id: 'settings',  href: '/settings',  icon: 'settings', label: 'Settings' },
];

const ADMIN_NAV_ITEM: NavItem = { id: 'admin', href: '/admin', icon: 'shield', label: 'Admin' };

interface SidebarItemProps {
  item: NavItem;
  active: boolean;
  collapsed: boolean;
}

function SidebarItem({ item, active, collapsed }: SidebarItemProps) {
  const [hovered, setHovered] = useState(false);

  const iconColor =
    active || hovered
      ? 'var(--text-on-dark)'
      : 'var(--text-on-dark-muted)';

  return (
    <Link
      href={item.href}
      title={collapsed ? item.label : undefined}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: collapsed ? 'center' : undefined,
        gap: 14,
        padding: collapsed ? '12px 0' : '12px 14px',
        borderRadius: 10,
        background: active
          ? 'var(--sidebar-active)'
          : hovered
          ? 'var(--sidebar-hover)'
          : 'transparent',
        textDecoration: 'none',
        transition: 'background 0.12s ease',
        marginBottom: 2,
        position: 'relative',
        width: '100%',
      }}
    >
      {active && (
        <div
          style={{
            position: 'absolute',
            left: 0,
            top: '50%',
            transform: 'translateY(-50%)',
            width: 3,
            height: '55%',
            background: 'var(--accent)',
            borderRadius: '0 2px 2px 0',
          }}
        />
      )}
      <Icon name={item.icon} size={24} color={iconColor} strokeWidth={active ? 2 : 1.7} />
      <span
        className="sidebar-label"
        style={{
          fontWeight: active ? 700 : 500,
          fontSize: 18,
          color: iconColor,
          letterSpacing: '-0.01em',
          lineHeight: 1,
        }}
      >
        {item.label}
      </span>
      {item.badge !== undefined && (
        <span
          style={{
            marginLeft: 'auto',
            minWidth: 18,
            height: 18,
            padding: '0 5px',
            background: 'var(--accent)',
            borderRadius: 9999,
            fontSize: 11,
            fontWeight: 500,
            color: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {item.badge}
        </span>
      )}
    </Link>
  );
}

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const { data: profile } = useMyProfile();
  const queryClient = useQueryClient();
  const [postHovered, setPostHovered] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [moreOpen, setMoreOpen] = useState(false);
  const [moreHovered, setMoreHovered] = useState(false);
  const moreRef = useRef<HTMLDivElement>(null);
  const avatarSrc = profile?.avatar ?? null;
  const isCollapsed = pathname.startsWith('/messages') || pathname.startsWith('/admin') || pathname === '/events';
  const moreItems = user?.roleName === 'admin' ? [...MORE_NAV_ITEMS, ADMIN_NAV_ITEM] : MORE_NAV_ITEMS;
  const moreActive = moreItems.some((item) => pathname.startsWith(item.href));

  useEffect(() => {
    if (!moreOpen) return;
    const handleClick = (e: MouseEvent) => {
      if (moreRef.current && !moreRef.current.contains(e.target as Node)) {
        setMoreOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [moreOpen]);
  const openCompose = useComposeStore((s) => s.openCompose);
  const unreadNotifications = useNotificationStore((s) => s.unreadCount);

  const handleLogout = async () => {
    await fetch('/auth/clear-session', { method: 'POST' });
    logout();
    queryClient.clear();
    router.replace('/login');
  };

  const handleNewPostClick = () => openCompose();

  const isActive = (item: NavItem): boolean => {
    if (item.href === '/') return pathname === '/';
    return pathname.startsWith(item.href);
  };

  return (
    <nav className={`sidebar-nav${isCollapsed ? ' sidebar-collapsed' : ''}`}>
      {/* Logo */}
      <div style={{ padding: isCollapsed ? '20px 0' : '24px 20px 20px', display: 'flex', justifyContent: isCollapsed ? 'center' : undefined }}>
        <button
          type="button"
          onClick={() => window.location.reload()}
          aria-label="Refresh Sunibgram"
          className="sidebar-logo-btn"
          style={{ display: 'flex', alignItems: 'center', gap: 6, background: 'none', border: 'none', padding: 0, cursor: 'pointer' }}
        >
          <span className="sidebar-logo-hit">
            <Logo size={32} style={{ flexShrink: 0, color: 'var(--text-on-dark)' }} />
          </span>
          <span
            className="sidebar-label"
            style={{ fontWeight: 500, fontSize: 20, color: 'var(--text-on-dark)', letterSpacing: '-0.02em' }}
          >
            Sunibgram
          </span>
        </button>
      </div>

      {/* Nav items */}
      <div style={{ padding: '0 10px 0 10px', paddingRight: isCollapsed ? 10 : 20, flex: 1 }}>
        {NAV_ITEMS.map((item) => {
          const badge = item.id === 'notifications' && unreadNotifications > 0
            ? (unreadNotifications > 99 ? 99 : unreadNotifications)
            : item.badge;
          return (
            <SidebarItem key={item.id} item={{ ...item, badge }} active={isActive(item)} collapsed={isCollapsed} />
          );
        })}

        {/* More button + popup */}
        <div ref={moreRef} style={{ position: 'relative' }}>
          <button
            type="button"
            title={isCollapsed ? 'More' : undefined}
            onClick={() => setMoreOpen((v) => !v)}
            onMouseEnter={() => setMoreHovered(true)}
            onMouseLeave={() => setMoreHovered(false)}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: isCollapsed ? 'center' : undefined,
              gap: 14,
              padding: isCollapsed ? '12px 0' : '12px 14px',
              borderRadius: 10,
              background: moreActive
                ? 'var(--sidebar-active)'
                : moreHovered || moreOpen
                ? 'var(--sidebar-hover)'
                : 'transparent',
              border: 'none',
              cursor: 'pointer',
              transition: 'background 0.12s ease',
              marginBottom: 2,
              position: 'relative',
              width: '100%',
              fontFamily: 'inherit',
              textAlign: 'left',
            }}
          >
            {moreActive && (
              <div
                style={{
                  position: 'absolute',
                  left: 0,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  width: 3,
                  height: '55%',
                  background: 'var(--accent)',
                  borderRadius: '0 2px 2px 0',
                }}
              />
            )}
            <Icon
              name="more"
              size={24}
              color={moreActive || moreHovered || moreOpen ? 'var(--text-on-dark)' : 'var(--text-on-dark-muted)'}
              strokeWidth={moreActive ? 2 : 1.7}
            />
            <span
              className="sidebar-label"
              style={{
                fontWeight: moreActive ? 700 : 500,
                fontSize: 18,
                color: moreActive || moreHovered || moreOpen ? 'var(--text-on-dark)' : 'var(--text-on-dark-muted)',
                letterSpacing: '-0.01em',
                lineHeight: 1,
              }}
            >
              More
            </span>
          </button>

          {moreOpen && (
            <div
              style={{
                position: 'absolute',
                bottom: '100%',
                left: isCollapsed ? '100%' : 0,
                marginBottom: isCollapsed ? 0 : 8,
                marginLeft: isCollapsed ? 8 : 0,
                minWidth: 260,
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 16,
                padding: 6,
                zIndex: 50,
                boxShadow: '0 8px 28px rgba(0,0,0,0.5)',
              }}
            >
              {moreItems.map((item) => {
                const active = pathname.startsWith(item.href);
                return (
                  <Link
                    key={item.id}
                    href={item.href}
                    onClick={() => setMoreOpen(false)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 14,
                      padding: '12px 14px',
                      borderRadius: 10,
                      textDecoration: 'none',
                      color: 'var(--text-on-dark)',
                      background: active ? 'var(--sidebar-active)' : 'transparent',
                      transition: 'background 0.12s ease',
                    }}
                    onMouseEnter={(e) => {
                      if (!active) e.currentTarget.style.background = 'var(--sidebar-hover)';
                    }}
                    onMouseLeave={(e) => {
                      if (!active) e.currentTarget.style.background = 'transparent';
                    }}
                  >
                    <Icon
                      name={item.icon}
                      size={22}
                      color="var(--text-on-dark)"
                      strokeWidth={active ? 2 : 1.7}
                    />
                    <span
                      style={{
                        fontWeight: active ? 700 : 500,
                        fontSize: 16,
                        color: 'var(--text-on-dark)',
                        letterSpacing: '-0.01em',
                      }}
                    >
                      {item.label}
                    </span>
                  </Link>
                );
              })}
            </div>
          )}
        </div>

        {/* New Post button */}
        <div style={{ marginTop: 16, padding: isCollapsed ? '0' : '0 2px 0 0', display: 'flex', justifyContent: isCollapsed ? 'center' : undefined }}>
          <button
            type="button"
            title={isCollapsed ? 'New Post' : undefined}
            onClick={handleNewPostClick}
            onMouseEnter={() => setPostHovered(true)}
            onMouseLeave={() => setPostHovered(false)}
            style={{
              width: isCollapsed ? 44 : '100%',
              height: isCollapsed ? 44 : undefined,
              padding: isCollapsed ? undefined : '13px 0',
              background: postHovered ? '#e6e6e6' : '#fff',
              border: 'none',
              cursor: 'pointer',
              borderRadius: 9999,
              color: '#000',
              fontWeight: 600,
              fontSize: 16,
              letterSpacing: '-0.02em',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 8,
              transition: 'background 0.15s ease',
            }}
          >
            <span className="sidebar-label">Post</span>
          </button>
        </div>
      </div>

      {/* Profile footer */}
      <div style={{ borderTop: '1px solid var(--border)', position: 'relative' }}>
        {menuOpen && (
          <div
            style={{
              position: 'absolute',
              bottom: '100%',
              left: 8,
              right: 8,
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 12,
              padding: 6,
              marginBottom: 4,
              zIndex: 50,
              boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
            }}
          >
            <button
              onClick={handleLogout}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: '9px 12px',
                background: 'none',
                border: 'none',
                borderRadius: 8,
                cursor: 'pointer',
                color: 'var(--coral)',
                fontSize: 14,
                fontFamily: 'inherit',
                fontWeight: 500,
                textAlign: 'left',
              }}
              onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--coral-soft)'; }}
              onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; }}
            >
              <LogOut size={16} />
              <span className="sidebar-label">Sign out</span>
            </button>
          </div>
        )}

        <button
          title={isCollapsed ? (profile?.name ?? user?.username ?? 'Account') : undefined}
          onClick={() => setMenuOpen((v) => !v)}
          className={`sidebar-account-btn${menuOpen ? ' is-open' : ''}`}
          style={{
            width: '100%',
            padding: '14px 12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: isCollapsed ? 'center' : undefined,
            gap: 10,
            background: menuOpen ? 'var(--sidebar-hover)' : 'none',
            border: 'none',
            cursor: 'pointer',
            fontFamily: 'inherit',
            transition: 'background 0.12s ease',
            textAlign: 'left',
          }}
          onMouseEnter={(e) => { if (!menuOpen) e.currentTarget.style.background = 'var(--sidebar-hover)'; }}
          onMouseLeave={(e) => { if (!menuOpen) e.currentTarget.style.background = 'none'; }}
        >
          <Avatar username={user?.username ?? 'Current user'} avatarUrl={avatarSrc} size={36} />
          <div className="sidebar-label" style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text-on-dark)', letterSpacing: '-0.02em', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'flex', alignItems: 'center', gap: 4 }}>
              <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {profile?.name ?? user?.username ?? '—'}
              </span>
              {profile?.is_premium && <VerifiedBadge size={14} />}
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-on-dark-muted)' }}>
              {profile?.username ?? user?.username ?? ''}
            </div>
          </div>
          <span className="sidebar-label sidebar-more-hit">
            <Icon name="more" size={16} />
          </span>
        </button>
      </div>
    </nav>
  );
}
