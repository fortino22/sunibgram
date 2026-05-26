'use client';

import Link from 'next/link';
import Image from 'next/image';
import { usePathname, useRouter } from 'next/navigation';
import { useState } from 'react';
import { LogOut } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { Avatar } from '@/components/shared/Avatar';
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
  { id: 'communities',   href: '/communities',   icon: 'users',    label: 'Communities'},
  { id: 'events',        href: '/events',        icon: 'calendar', label: 'Events'       },
  { id: 'messages',      href: '/messages',      icon: 'messages', label: 'Messages'     },
  { id: 'notifications', href: '/notifications', icon: 'bell',     label: 'Notifications'},
  { id: 'bookmarks',     href: '/bookmarks',     icon: 'bookmark', label: 'Bookmarks' },
  { id: 'profile',       href: '/profile',       icon: 'user',     label: 'Profile' },
];

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
  const avatarSrc = profile?.avatar ?? null;
  const isCollapsed = pathname.startsWith('/messages');
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
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={32} height={32} style={{ borderRadius: 10, flexShrink: 0 }} />
          <span
            className="sidebar-label"
            style={{ fontWeight: 500, fontSize: 20, color: 'var(--text-on-dark)', letterSpacing: '-0.02em' }}
          >
            Sunibgram
          </span>
        </div>
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
              borderRadius: isCollapsed ? '50%' : 14,
              color: '#000',
              fontWeight: 600,
              fontSize: 17,
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
            <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text-on-dark)', letterSpacing: '-0.02em', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {profile?.name ?? user?.username ?? '—'}
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-on-dark-muted)' }}>
              {profile?.username ?? user?.username ?? ''}
            </div>
          </div>
          <div className="sidebar-label">
            <Icon name="more" size={16} color="var(--text-on-dark-muted)" />
          </div>
        </button>
      </div>
    </nav>
  );
}
