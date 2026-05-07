'use client';

import Link from 'next/link';
import Image from 'next/image';
import { usePathname, useRouter } from 'next/navigation';
import { useState } from 'react';
import { LogOut } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { Icon } from '@/components/ui/Icon';
import { useAuthStore } from '@/stores/authStore';

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
  { id: 'notifications', href: '/notifications', icon: 'bell',     label: 'Notifications'},
  { id: 'bookmarks',     href: '/bookmarks',     icon: 'bookmark', label: 'Bookmarks' },
  { id: 'profile',       href: '/profile',       icon: 'user',     label: 'Profile' },
];

interface SidebarItemProps {
  item: NavItem;
  active: boolean;
}

function SidebarItem({ item, active }: SidebarItemProps) {
  const [hovered, setHovered] = useState(false);

  const iconColor =
    active || hovered
      ? 'var(--text-on-dark)'
      : 'var(--text-on-dark-muted)';

  return (
    <Link
      href={item.href}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '10px 12px',
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
      <Icon name={item.icon} size={20} color={iconColor} strokeWidth={active ? 2 : 1.7} />
      <span
        className="sidebar-label"
        style={{
          fontWeight: active ? 600 : 400,
          fontSize: 15,
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
            background: 'var(--coral)',
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
  const [postHovered, setPostHovered] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const userWithAvatar = user as (typeof user & {
    avatar_url?: string | null;
    image_url?: string | null;
    profile_image_url?: string | null;
    profileImageUrl?: string | null;
  }) | null;
  const avatarSrc =
    userWithAvatar?.profile_image_url ??
    userWithAvatar?.profileImageUrl ??
    userWithAvatar?.avatar_url ??
    userWithAvatar?.image_url ??
    null;

  const handleLogout = async () => {
    await fetch('/api/auth/clear-session', { method: 'POST' });
    logout();
    router.replace('/login');
  };

  const handleNewPostClick = () => {
    if (pathname === '/') {
      const composer = document.getElementById('create-post-box');
      composer?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      const textarea = composer?.querySelector<HTMLTextAreaElement>('textarea');
      textarea?.focus();
      return;
    }

    router.push('/#create-post-box');
  };

  const isActive = (item: NavItem): boolean => {
    if (item.href === '/') return pathname === '/';
    return pathname.startsWith(item.href);
  };

  return (
    <nav className="sidebar-nav">
      {/* Logo */}
      <div style={{ padding: '24px 20px 20px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={32} height={32} style={{ borderRadius: 10, flexShrink: 0 }} />
          <span
            className="sidebar-label"
            style={{
              fontWeight: 500,
              fontSize: 20,
              color: 'var(--text-on-dark)',
              letterSpacing: '-0.02em',
            }}
          >
            Sunibgram
          </span>
        </div>
        <div
          className="sidebar-label"
          style={{
            fontSize: 10,
            color: 'var(--text-on-dark-muted)',
            marginTop: 4,
            marginLeft: 42,
            letterSpacing: '0.08em',
            textTransform: 'uppercase',
            fontWeight: 500,
          }}
        >
          Campus Discourse
        </div>
      </div>

      {/* Nav items */}
      <div style={{ padding: '0 10px', flex: 1 }}>
        {NAV_ITEMS.map((item) => (
          <SidebarItem key={item.id} item={item} active={isActive(item)} />
        ))}

        {/* New Post button */}
        <div style={{ marginTop: 16, padding: '0 2px' }}>
          <button
            type="button"
            onClick={handleNewPostClick}
            onMouseEnter={() => setPostHovered(true)}
            onMouseLeave={() => setPostHovered(false)}
            style={{
              width: '100%',
              padding: '11px 0',
              background: postHovered ? 'var(--accent-hover)' : 'var(--accent)',
              border: 'none',
              cursor: 'pointer',
              borderRadius: 14,
              color: '#fff',
              fontWeight: 500,
              fontSize: 15,
              letterSpacing: '-0.02em',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 8,
              transition: 'background 0.15s ease',
              boxShadow: '0 4px 16px oklch(64% 0.22 270 / 0.35)',
            }}
          >
            <Icon name="plus" size={18} color="#fff" />
            <span className="sidebar-label">New Post</span>
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
          onClick={() => setMenuOpen((v) => !v)}
          style={{
            width: '100%',
            padding: '14px 12px',
            display: 'flex',
            alignItems: 'center',
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
            <div
              style={{
                fontWeight: 500,
                fontSize: 14,
                color: 'var(--text-on-dark)',
                letterSpacing: '-0.02em',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {user?.username ?? '—'}
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-on-dark-muted)' }}>
              {user?.email ?? ''}
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
