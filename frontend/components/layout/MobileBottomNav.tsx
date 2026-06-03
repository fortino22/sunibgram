'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Icon } from '@/components/ui/Icon';
import { useNotificationStore } from '@/stores/notificationStore';

interface MobileNavItem {
  id: string;
  href: string;
  icon: string;
  label: string;
}

// Order: home, explore, fox (center), notifications, messages.
const ITEMS: MobileNavItem[] = [
  { id: 'home',          href: '/',              icon: 'home',     label: 'Home' },
  { id: 'explore',       href: '/explore',       icon: 'explore',  label: 'Explore' },
  { id: 'fox',           href: '/summaries',     icon: 'fox',      label: 'Fox' },
  { id: 'notifications', href: '/notifications', icon: 'bell',     label: 'Notifications' },
  { id: 'messages',      href: '/messages',      icon: 'messages', label: 'Messages' },
];

function isActive(pathname: string, href: string): boolean {
  if (href === '/') return pathname === '/';
  return pathname === href || pathname.startsWith(href + '/');
}

export function MobileBottomNav() {
  const pathname = usePathname();
  const unread = useNotificationStore((s) => s.unreadCount);

  return (
    <nav className="mobile-bottom-nav" aria-label="Primary">
      {ITEMS.map((item) => {
        const active = isActive(pathname, item.href);
        const badge = item.id === 'notifications' && unread > 0 ? (unread > 99 ? 99 : unread) : undefined;
        const color = active ? 'var(--text)' : 'var(--text-tertiary)';
        return (
          <Link
            key={item.id}
            href={item.href}
            className="mobile-bottom-nav-item"
            aria-label={item.label}
            aria-current={active ? 'page' : undefined}
          >
            <span style={{ position: 'relative', display: 'inline-flex' }}>
              {item.icon === 'fox' ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src="/assets/fox.png"
                  alt=""
                  width={26}
                  height={26}
                  style={{
                    display: 'block',
                    opacity: active ? 1 : 0.7,
                    filter: 'var(--fox-filter, none)',
                  }}
                />
              ) : (
                <Icon name={item.icon} size={26} color={color} strokeWidth={active ? 2.2 : 1.8} />
              )}
              {badge !== undefined && (
                <span className="mobile-bottom-nav-badge">{badge}</span>
              )}
            </span>
          </Link>
        );
      })}
    </nav>
  );
}
