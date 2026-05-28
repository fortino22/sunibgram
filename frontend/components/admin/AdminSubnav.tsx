'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface NavLink {
  href: string;
  label: string;
}

const LINKS: NavLink[] = [
  { href: '/admin', label: 'Overview' },
  { href: '/admin/users', label: 'Users' },
  { href: '/admin/reports', label: 'Reports' },
  { href: '/admin/trending-reports', label: 'Trending Reports' },
  { href: '/admin/events', label: 'Events' },
  { href: '/admin/faculties', label: 'Faculties' },
  { href: '/admin/violations', label: 'Report Types' },
  { href: '/admin/categories', label: 'Categories' },
  { href: '/admin/orders', label: 'Orders' },
  { href: '/admin/feedback', label: 'Feedback' },
];

export function AdminSubnav() {
  const pathname = usePathname();

  const isActive = (href: string): boolean => {
    if (href === '/admin') return pathname === '/admin';
    return pathname.startsWith(href);
  };

  return (
    <nav
      style={{
        display: 'flex',
        gap: 4,
        padding: '12px 24px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg)',
        position: 'sticky',
        top: 0,
        zIndex: 10,
        overflowX: 'auto',
        flexShrink: 0,
      }}
    >
      {LINKS.map((l) => {
        const active = isActive(l.href);
        return (
          <Link
            key={l.href}
            href={l.href}
            style={{
              padding: '8px 14px',
              borderRadius: 9999,
              background: active ? 'var(--sidebar-active)' : 'transparent',
              color: active ? 'var(--text)' : 'var(--text-secondary)',
              fontSize: 14,
              fontWeight: active ? 600 : 500,
              textDecoration: 'none',
              whiteSpace: 'nowrap',
              transition: 'background 0.12s ease',
            }}
            className="btn-ghost"
          >
            {l.label}
          </Link>
        );
      })}
    </nav>
  );
}
